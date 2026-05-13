#if DEBUG
import Foundation

struct UITestingAuthService: AuthServiceProtocol {

    private var env: [String: String] { ProcessInfo.processInfo.environment }

    func login(user: LoginPayload) async throws {
        if env["UI_TESTING_AUTH_RESULT"] == "error" {
            throw NetworkError.requestFailed(statusCode: 401, message: "Неверный логин или пароль")
        }
        UserSessionManager.shared.login(
            userId: "uitest_user",
            accessToken: "test_access_token",
            refreshToken: "test_refresh_token"
        )
        UserSessionManager.shared.setHasCompletedParametersSetup(true, for: "uitest_user")
    }

    func register(user: RegisterPayload) async throws {
        if env["UI_TESTING_AUTH_RESULT"] == "error" {
            throw NetworkError.requestFailed(statusCode: 409, message: "Пользователь уже существует")
        }
    }

    func sendRecoveryCode(email: String) async throws {
        if env["UI_TESTING_RECOVERY_RESULT"] == "error" {
            throw NetworkError.requestFailed(statusCode: 404, message: "Email не найден")
        }
    }

    func confirmRecovery(email: String, code: String, newPassword: String) async throws {
        if env["UI_TESTING_RECOVERY_RESULT"] == "error" {
            throw NetworkError.requestFailed(statusCode: 400, message: "Неверный код подтверждения")
        }
    }

    func sendUserParameters() async throws {}
}
#endif
