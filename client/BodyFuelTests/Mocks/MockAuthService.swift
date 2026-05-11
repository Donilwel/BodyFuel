import Foundation
@testable import BodyFuel

final class MockAuthService: AuthServiceProtocol {

    // MARK: - Call tracking

    var loginCallCount = 0
    var registerCallCount = 0
    var sendRecoveryCodeCallCount = 0
    var confirmRecoveryCallCount = 0
    var sendUserParametersCallCount = 0

    var lastLoginPayload: LoginPayload?
    var lastRegisterPayload: RegisterPayload?
    var lastRecoveryEmail: String?
    var lastConfirmRecoveryArgs: (email: String, code: String, newPassword: String)?

    // MARK: - Configurable responses

    var loginResult: Result<Void, Error> = .success(())
    var registerResult: Result<Void, Error> = .success(())
    var sendRecoveryCodeResult: Result<Void, Error> = .success(())
    var confirmRecoveryResult: Result<Void, Error> = .success(())
    var sendUserParametersResult: Result<Void, Error> = .success(())

    // MARK: - Protocol

    func login(user: LoginPayload) async throws {
        loginCallCount += 1
        lastLoginPayload = user
        _ = try loginResult.get()
    }

    func register(user: RegisterPayload) async throws {
        registerCallCount += 1
        lastRegisterPayload = user
        _ = try registerResult.get()
    }

    func sendRecoveryCode(email: String) async throws {
        sendRecoveryCodeCallCount += 1
        lastRecoveryEmail = email
        _ = try sendRecoveryCodeResult.get()
    }

    func confirmRecovery(email: String, code: String, newPassword: String) async throws {
        confirmRecoveryCallCount += 1
        lastConfirmRecoveryArgs = (email, code, newPassword)
        _ = try confirmRecoveryResult.get()
    }

    func sendUserParameters() async throws {
        sendUserParametersCallCount += 1
        _ = try sendUserParametersResult.get()
    }
}
