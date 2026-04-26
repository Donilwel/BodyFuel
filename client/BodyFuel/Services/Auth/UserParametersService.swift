import Foundation

protocol UserParametersServiceProtocol {
    func hasUserParameters() async -> Bool
    func sendUserParameters(_ parameters: UserParametersPayload) async throws
    func sendCurrentWeight(_ weight: Float) async throws
}

enum UserParametersError: LocalizedError {
    case invalidData(String)
    
    var errorDescription: String? {
        switch self {
        case .invalidData(let message): return message
        }
    }
}

final class UserParametersService: UserParametersServiceProtocol {
    static let shared = UserParametersService()
    
    private let networkClient = NetworkClient.shared
    
    private let photoService: PhotoServiceProtocol = PhotoService.shared
    
    private init() {}
    
    func hasUserParameters() async -> Bool {
        guard let url = URL(string: API.baseURLString + API.userParameters) else { return false }
        do {
            let _: DefaultDecodable = try await networkClient.request(url: url, method: .get)
            return true
        } catch {
            return false
        }
    }

    func sendUserParameters(_ parametersPayload: UserParametersPayload) async throws {
        let urlComponents = URLComponents(string: API.baseURLString + API.userParameters)
        
        guard let urlComponents, let url = urlComponents.url else {
            print("[ERROR] [UserParametersService/sendUserParameters]: Invalid user parameters URL")
            throw NetworkError.invalidURL
        }
        
        do {
            let avatarURL = try await photoService.uploadUserAvatar(data: parametersPayload.avatarData)
            
            let request = UserParametersRequestBody(
                from: parametersPayload,
                avatarURL: avatarURL
            )

            let response: APIMessageResponse = try await networkClient.request(
                url: url,
                method: .post,
                requestBody: request
            )
            
            print("[INFO] [UserParametersService/sendUserParameters]: Successfully sent user parameters: \(response.message)")
        } catch {
            throw UserParametersError.invalidData(russianMessage(for: error))
        }
    }

    func sendCurrentWeight(_ weight: Float) async throws {
        let urlComponents = URLComponents(string: API.baseURLString + API.weight)

        guard let urlComponents, let url = urlComponents.url else {
            print("[ERROR] [UserParametersService/sendCurrentWeight]: Invalid user weight URL")
            throw NetworkError.invalidURL
        }

        do {
            let request = UserWeightRequestBody(weight: weight)

            let response: APIMessageResponse = try await networkClient.request(
                url: url,
                method: .post,
                requestBody: request
            )

            print("[INFO] [UserParametersService/sendCurrentWeight] Successfully sent user weight: \(response)")
        } catch {
            throw UserParametersError.invalidData(russianMessage(for: error))
        }
    }

    private func russianMessage(for error: Error) -> String {
        guard let networkError = error as? NetworkError else {
            return "Произошла ошибка, попробуйте позже"
        }
        switch networkError {
        case .requestFailed: return "Ошибка сервера, попробуйте позже"
        case .decodingFailed, .encodingFailed: return "Ошибка обработки данных"
        case .missingToken: return "Требуется повторный вход"
        case .invalidURL: return "Сервер временно недоступен"
        case .network(let underlying):
            if let urlError = underlying as? URLError,
               urlError.code == .notConnectedToInternet || urlError.code == .networkConnectionLost {
                return "Нет подключения к интернету"
            }
            return "Сервер временно недоступен"
        }
    }
}
