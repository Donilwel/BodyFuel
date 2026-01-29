import Foundation

protocol UserParametersServiceProtocol {
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
    private let tokenStorage = TokenStorage.shared
    
    private let photoService: PhotoServiceProtocol = PhotoService.shared
    
    private init() {}
    
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
            throw UserParametersError.invalidData(error.localizedDescription)
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
            
            print("[INFO] [UserParametersService/sendCurrentWeight] Successfully sent user weight")
        } catch {
            throw UserParametersError.invalidData(error.localizedDescription)
        }
    }
}
