import Foundation

enum NetworkError: Error, LocalizedError {
    case invalidURL
    case missingToken
    case requestFailed(statusCode: Int, message: String)
    case decodingFailed
    case encodingFailed
    case network(Error)

    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "Invalid URL"
        case .missingToken:
            return "Token is missing"
        case .requestFailed(let statusCode, let msg):
            return "Invalid request: HTTP \(statusCode), \(msg)"
        case .decodingFailed:
            return "Decoding failed"
        case .encodingFailed:
            return "Encoding failed"
        case .network(let error):
            return "Network error: \(error.localizedDescription)"
        }
    }
}

enum HTTPMethod: String {
    case get = "GET"
    case post = "POST"
    case put = "PUT"
    case delete = "DELETE"
    case patch = "PATCH"
}

private actor TokenRefreshCoordinator {
    private var task: Task<String, Error>?

    func refresh(using perform: @escaping () async throws -> String) async throws -> String {
        if let existing = task {
            return try await existing.value
        }
        let newTask = Task { try await perform() }
        task = newTask
        defer { task = nil }
        return try await newTask.value
    }
}

final class NetworkClient {
    static let shared = NetworkClient()

    private let session: URLSession = {
        let config = URLSessionConfiguration.default
        config.timeoutIntervalForRequest = 10
        config.timeoutIntervalForResource = 30
        return URLSession(configuration: config)
    }()
    private let userSessionManager = UserSessionManager.shared
    private let refreshCoordinator = TokenRefreshCoordinator()

    private init() {}

    func request<T: Decodable, U: Encodable>(
        requiresAuthorization: Bool = true,
        url: URL,
        method: HTTPMethod,
        requestBody: U? = Optional<DefaultEncodable>.none
    ) async throws -> T {
        var request = URLRequest(url: url)

        if requiresAuthorization {
            guard let currentUserId = userSessionManager.currentUserId,
                  let token = userSessionManager.authToken(for: currentUserId) else {
                throw NetworkError.missingToken
            }
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        request.httpMethod = method.rawValue
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        if let body = requestBody {
            do {
                let jsonEncoder = JSONEncoder()
                jsonEncoder.dateEncodingStrategy = .iso8601
                request.httpBody = try await jsonEncoder.encodeAsync(body)
            } catch {
                throw NetworkError.encodingFailed
            }
        }

        do {
            let (data, response) = try await session.data(for: request)

            guard let httpResponse = response as? HTTPURLResponse else {
                let errorMessage = extractErrorMessage(from: data)
                throw NetworkError.requestFailed(statusCode: -1, message: errorMessage)
            }

            if httpResponse.statusCode == 401 && requiresAuthorization {
                let newToken = try await refreshAccessToken()
                request.setValue("Bearer \(newToken)", forHTTPHeaderField: "Authorization")

                let (retryData, retryResponse) = try await session.data(for: request)
                guard let retryHTTP = retryResponse as? HTTPURLResponse else {
                    throw NetworkError.requestFailed(statusCode: -1, message: "Retry failed")
                }
                guard 200..<300 ~= retryHTTP.statusCode else {
                    if retryHTTP.statusCode == 401 { throw NetworkError.missingToken }
                    let errorMessage = extractErrorMessage(from: retryData)
                    throw NetworkError.requestFailed(statusCode: retryHTTP.statusCode, message: errorMessage)
                }
                return try decodeResponse(T.self, from: retryData)
            }

            guard 200..<300 ~= httpResponse.statusCode else {
                let errorMessage = extractErrorMessage(from: data)
                throw NetworkError.requestFailed(statusCode: httpResponse.statusCode, message: errorMessage)
            }

            return try decodeResponse(T.self, from: data)
        } catch let error as NetworkError {
            throw error
        } catch {
            throw NetworkError.network(error)
        }
    }

    private func refreshAccessToken() async throws -> String {
        try await refreshCoordinator.refresh { [weak self] in
            guard let self else { throw NetworkError.missingToken }
            return try await self.performTokenRefresh()
        }
    }

    private func performTokenRefresh() async throws -> String {
        guard let userId = userSessionManager.currentUserId,
              let storedRefreshToken = userSessionManager.refreshToken(for: userId) else {
            throw NetworkError.missingToken
        }

        guard let url = URL(string: API.baseURLString + API.Auth.refresh) else {
            throw NetworkError.invalidURL
        }

        var refreshRequest = URLRequest(url: url)
        refreshRequest.httpMethod = HTTPMethod.post.rawValue
        refreshRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")
        refreshRequest.httpBody = try? JSONEncoder().encode(["refresh_token": storedRefreshToken])

        let (data, response) = try await session.data(for: refreshRequest)

        guard let httpResponse = response as? HTTPURLResponse,
              200..<300 ~= httpResponse.statusCode else {
            throw NetworkError.missingToken
        }

        let tokens = try JSONDecoder().decode(LoginResponseBody.self, from: data)
        userSessionManager.login(userId: userId, accessToken: tokens.accessToken, refreshToken: tokens.refreshToken)

        print("[INFO] [NetworkClient/refreshAccessToken]: Token refreshed for user \(userId)")
        return tokens.accessToken
    }

    private func decodeResponse<T: Decodable>(_ type: T.Type, from data: Data) throws -> T {
        if data.isEmpty {
            if T.self == DefaultDecodable.self {
                guard let emptyResponse = DefaultDecodable() as? T else {
                    throw NetworkError.decodingFailed
                }
                return emptyResponse
            } else {
                throw NetworkError.decodingFailed
            }
        }

        do {
            let jsonDecoder = JSONDecoder()
            return try jsonDecoder.decode(T.self, from: data)
        } catch {
            throw NetworkError.decodingFailed
        }
    }
    
    func uploadPhoto(data: Data, to url: URL) async throws {
        do {
            var request = URLRequest(url: url)
            request.httpMethod = "PUT"
            request.setValue("image/jpeg", forHTTPHeaderField: "Content-Type")
            
            let (_, response) = try await URLSession.shared.upload(for: request, from: data)
            
            guard let httpResponse = response as? HTTPURLResponse else {
                let errorMessage = extractErrorMessage(from: data)
                throw NetworkError.requestFailed(statusCode: -1, message: errorMessage)
            }
            
            guard 200..<300 ~= httpResponse.statusCode else {
                let errorMessage = extractErrorMessage(from: data)
                throw NetworkError.requestFailed(statusCode: httpResponse.statusCode, message: errorMessage)
            }
        } catch let error as NetworkError {
            throw error
        } catch {
            throw NetworkError.network(error)
        }
    }
    
    private func extractErrorMessage(from data: Data) -> String {
        if let apiError = try? JSONDecoder().decode(APIMessageResponse.self, from: data) {
            return apiError.message
        }

        if let text = String(data: data, encoding: .utf8), !text.isEmpty {
            return text
        }

        return "Ошибка сервера"
    }
}
