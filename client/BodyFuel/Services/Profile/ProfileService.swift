import Foundation

protocol ProfileServiceProtocol {
    func fetchProfile() async throws -> UserProfile
    func updateProfile(_ profile: UserProfile) async throws
    func deleteProfile() async throws
    func logout()
}

enum ProfileError: LocalizedError {
    case validation
    case unauthorized
    case invalidData(String)

    var errorDescription: String? {
        switch self {
        case .validation: return "Ошибка валидации"
        case .unauthorized: return "Требуется повторный вход"
        case .invalidData(let message): return message
        }
    }
}

final class ProfileService: ProfileServiceProtocol {
    static let shared = ProfileService()
    
    private let networkClient = NetworkClient.shared
    private let tokenStorage = TokenStorage.shared
    
    private init() {}

    func fetchProfile() async throws -> UserProfile {
        let urlComponents = URLComponents(string: API.baseURLString + API.userParameters)
        
        guard let urlComponents, let url = urlComponents.url else {
            print("[ERROR] [ProfileService/fetchProfile]: Invalid user parameters URL")
            throw NetworkError.invalidURL
        }
        
        do {
            let response: UserParametersResponseBody = try await networkClient.request(
                url: url,
                method: .get
            )
            
            let userProfile = UserProfile(from: response)
            
            print("[INFO] [ProfileService/fetchProfile]: Successfully fetched user parameters")
            return userProfile
        } catch {
            print("[ERROR] [ProfileService/fetchProfile]: \(error.localizedDescription)")
            throw mapToProfileError(error)
        }
    }

    func updateProfile(_ profile: UserProfile) async throws {
        let urlComponents = URLComponents(string: API.baseURLString + API.userParameters)
        
        guard let urlComponents, let url = urlComponents.url else {
            print("[ERROR] [ProfileService/updateProfile]: Invalid user parameters URL")
            throw NetworkError.invalidURL
        }
        
        do {
            let request = UserParametersRequestBody(from: profile)
            
            let response: APIMessageResponse = try await networkClient.request(
                url: url,
                method: .patch,
                requestBody: request
            )
            
            print("[INFO] [ProfileService/updateProfile]: Successfully updated user parameters: \(response.message)")
        } catch {
            print("[ERROR] [ProfileService/updateProfile]: \(error.localizedDescription)")
            throw mapToProfileError(error)
        }
    }
    
    func deleteProfile() async throws {
        do {
            let urlComponents = URLComponents(string: API.baseURLString + API.userInfo)
            guard let urlComponents, let url = urlComponents.url else {
                print("[ERROR] [ProfileService/deleteProfile] Invalid delete user info URL")
                throw NetworkError.invalidURL
            }
            
            let response: APIMessageResponse = try await networkClient.request(
                url: url,
                method: .delete
            )
            
            tokenStorage.deleteToken()
            
            print("[INFO] [ProfileService/deleteProfile]: Successfully deleted profile")
        } catch {
            print("[ERROR] [ProfileService/deleteProfile]: \(error.localizedDescription)")
            throw mapToProfileError(error)
        }
    }
    
    func logout() {
        tokenStorage.deleteToken()
    }
}

extension ProfileService {
    private func mapToProfileError(_ error: Error) -> ProfileError {
        guard let networkError = error as? NetworkError else {
            return .invalidData(error.localizedDescription)
        }
        
        switch networkError {
        case .requestFailed(let statusCode, let message):
            switch statusCode {
            case 400:
                return .validation
            case 401:
                return .unauthorized
            default:
                return .invalidData(message)
            }
            
        default:
            return .invalidData(error.localizedDescription)
        }
    }
}
