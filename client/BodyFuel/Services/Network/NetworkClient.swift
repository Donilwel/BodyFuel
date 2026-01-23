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

final class NetworkClient {
    static let shared = NetworkClient()

    private let session = URLSession(configuration: .default)

    private init() {}

    func request<T: Decodable, U: Encodable>(
        requiresAuthorization: Bool = true,
        url: URL,
        method: HTTPMethod,
        requestBody: U? = Optional<DefaultEncodable>.none
    ) async throws -> T {
        var request = URLRequest(url: url)
        
        if requiresAuthorization {
            guard let token = TokenStorage.shared.token else {
                throw NetworkError.missingToken
            }
            
            request.setValue("OAuth \(token)", forHTTPHeaderField: "authorization")
        }
        
        request.httpMethod = method.rawValue
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        if let body = requestBody {
            do {
                let jsonEncoder = JSONEncoder()
                jsonEncoder.dateEncodingStrategy = .iso8601
                request.httpBody = try await JSONEncoder().encodeAsync(body)
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

            guard 200..<300 ~= httpResponse.statusCode else {
                let errorMessage = extractErrorMessage(from: data)
                throw NetworkError.requestFailed(statusCode: httpResponse.statusCode, message: errorMessage)
            }

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

                return try await jsonDecoder.decodeAsync(T.self, from: data)
            } catch {
                throw NetworkError.decodingFailed
            }
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
