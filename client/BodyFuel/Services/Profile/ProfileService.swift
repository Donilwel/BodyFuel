import Foundation

protocol ProfileServiceProtocol {
    func fetchProfile() async throws -> UserProfile
    func updateProfile(_ profile: UserProfile) async throws
    func deleteProfile() async throws
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
    private let sessionManager = UserSessionManager.shared

    private init() {}

    func fetchProfile() async throws -> UserProfile {
        guard let url = URLComponents(string: API.baseURLString + API.userParameters)?.url else {
            print("[ERROR] [ProfileService/fetchProfile]: Invalid user parameters URL")
            throw NetworkError.invalidURL
        }

        do {
            let response: UserParametersResponseBody = try await networkClient.request(
                url: url,
                method: .get
            )
            print("[INFO] [ProfileService/fetchProfile]: Successfully fetched user parameters")
            return UserProfile(from: response)
        } catch {
            print("[ERROR] [ProfileService/fetchProfile]: \(error.localizedDescription)")
            throw mapToProfileError(error)
        }
    }

    func updateProfile(_ profile: UserProfile) async throws {
        guard let url = URLComponents(string: API.baseURLString + API.userParameters)?.url else {
            print("[ERROR] [ProfileService/updateProfile]: Invalid user parameters URL")
            throw NetworkError.invalidURL
        }

        do {
            let response: APIMessageResponse = try await networkClient.request(
                url: url,
                method: .patch,
                requestBody: UserParametersRequestBody(from: profile)
            )
            print("[INFO] [ProfileService/updateProfile]: Successfully updated user parameters: \(response.message)")
        } catch {
            print("[ERROR] [ProfileService/updateProfile]: \(error.localizedDescription)")
            throw mapToProfileError(error)
        }
    }

    func deleteProfile() async throws {
        guard let url = URLComponents(string: API.baseURLString + API.userInfo)?.url else {
            print("[ERROR] [ProfileService/deleteProfile] Invalid delete user info URL")
            throw NetworkError.invalidURL
        }

        do {
            let _: APIMessageResponse = try await networkClient.request(
                url: url,
                method: .delete
            )
            if let userId = sessionManager.currentUserId {
                sessionManager.deleteUser(userId: userId)
            }
            print("[INFO] [ProfileService/deleteProfile]: Successfully deleted profile")
        } catch {
            print("[ERROR] [ProfileService/deleteProfile]: \(error.localizedDescription)")
            throw mapToProfileError(error)
        }
    }
}

extension ProfileService {
    private func mapToProfileError(_ error: Error) -> ProfileError {
        guard let networkError = error as? NetworkError else {
            return .invalidData("Не удалось загрузить профиль")
        }

        switch networkError {
        case .requestFailed(let statusCode, _):
            switch statusCode {
            case 400: return .validation
            case 401: return .unauthorized
            default: return .invalidData("Ошибка сервера, попробуйте позже")
            }
        case .decodingFailed, .encodingFailed:
            return .invalidData("Ошибка обработки данных")
        case .missingToken:
            return .unauthorized
        case .invalidURL:
            return .invalidData("Сервер временно недоступен")
        case .network(let underlying):
            if let urlError = underlying as? URLError,
               urlError.code == .notConnectedToInternet || urlError.code == .networkConnectionLost {
                return .invalidData("Нет подключения к интернету")
            }
            return .invalidData("Сервер временно недоступен")
        }
    }
}
