import Foundation

protocol AuthServiceProtocol {
    func login(user: LoginPayload) async throws
    func register(user: RegisterPayload) async throws
    func sendRecoveryCode(email: String) async throws
    func confirmRecovery(email: String, code: String, newPassword: String) async throws
    func sendUserParameters() async throws
}

final class AuthService: AuthServiceProtocol {
    static let shared = AuthService()

    private let networkClient = NetworkClient.shared

    private init() {}

    func login(user: LoginPayload) async throws {
        do {
            let urlComponents = URLComponents(string: API.baseURLString + API.Auth.login)
            guard let urlComponents, let url = urlComponents.url else {
                print("[ERROR] [AuthService/login] Invalid login URL")
                throw NetworkError.invalidURL
            }

            let response: LoginResponseBody = try await networkClient.request(
                requiresAuthorization: false,
                url: url,
                method: .post,
                requestBody: user
            )

            UserSessionManager.shared.login(userId: user.username, token: response.token)

            print("[INFO] [AuthService/login]: Successfully logged in, token: \(response.token)")
        } catch {
            print("[ERROR] [AuthService/login]: \(error.localizedDescription)")
            throw mapToAuthError(error)
        }
    }

    func register(user: RegisterPayload) async throws {
        do {
            let urlComponents = URLComponents(string: API.baseURLString + API.Auth.register)!
            guard let url = urlComponents.url else {
                print("[ERROR] [AuthService/register] Invalid register URL")
                throw NetworkError.invalidURL
            }

            let response: APIMessageResponse = try await networkClient.request(
                requiresAuthorization: false,
                url: url,
                method: .post,
                requestBody: user
            )

            print("[INFO] [AuthService/register]: \(response.message)")
        } catch {
            print("[ERROR] [AuthService/register]: \(error.localizedDescription)")
            throw mapToAuthError(error)
        }
    }

    func sendRecoveryCode(email: String) async throws {
        do {
            guard let url = URLComponents(string: API.baseURLString + API.Auth.recover)?.url else {
                print("[ERROR] [AuthService/sendRecoveryCode] Invalid recover URL")
                throw NetworkError.invalidURL
            }

            let response: APIMessageResponse = try await networkClient.request(
                requiresAuthorization: false,
                url: url,
                method: .post,
                requestBody: RecoverPasswordRequestBody(email: email)
            )

            print("[INFO] [AuthService/sendRecoveryCode]: \(response.message)")
        } catch {
            print("[ERROR] [AuthService/sendRecoveryCode]: \(error.localizedDescription)")
            throw mapToAuthError(error)
        }
    }

    func confirmRecovery(email: String, code: String, newPassword: String) async throws {
        do {
            guard let url = URLComponents(string: API.baseURLString + API.Auth.resetPassword)?.url else {
                print("[ERROR] [AuthService/confirmRecovery] Invalid reset-password URL")
                throw NetworkError.invalidURL
            }

            let response: APIMessageResponse = try await networkClient.request(
                requiresAuthorization: false,
                url: url,
                method: .post,
                requestBody: ResetPasswordRequestBody(email: email, code: code, newPassword: newPassword)
            )

            print("[INFO] [AuthService/confirmRecovery]: \(response.message)")
        } catch {
            print("[ERROR] [AuthService/confirmRecovery]: \(error.localizedDescription)")
            throw mapToAuthError(error)
        }
    }

    func sendUserParameters() async throws {}
}

extension AuthService {
    private func mapToAuthError(_ error: Error) -> AuthError {
        guard let networkError = error as? NetworkError else {
            return .invalidData(error.localizedDescription)
        }

        switch networkError {
        case .requestFailed(let statusCode, let message):
            switch statusCode {
            case 400:
                return .validation
            case 401:
                return .invalidCredentials
            case 409:
                return .userExists
            default:
                return .invalidData(message)
            }

        default:
            return .invalidData(error.localizedDescription)
        }
    }
}
