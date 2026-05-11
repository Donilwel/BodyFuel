import Foundation
@testable import BodyFuel

final class MockUserSessionManager {

    // MARK: - State

    var currentUserId: String? = "test-user-id"
    var hasCompletedOnboarding: Bool = true
    var hasCompletedParametersSetup: Bool = true

    // MARK: - Call tracking

    var loginCallCount = 0
    var logoutCallCount = 0
    var deleteUserCallCount = 0

    var lastLoginUserId: String?
    var lastLogoutUserId: String?
    var lastDeletedUserId: String?

    // MARK: - Methods

    func login(userId: String, accessToken: String, refreshToken: String) {
        loginCallCount += 1
        lastLoginUserId = userId
        currentUserId = userId
    }

    func logout(userId: String? = nil) {
        logoutCallCount += 1
        lastLogoutUserId = userId ?? currentUserId
        currentUserId = nil
    }

    func deleteUser(userId: String) {
        deleteUserCallCount += 1
        lastDeletedUserId = userId
        if userId == currentUserId { currentUserId = nil }
    }

    func authToken(for userId: String) -> String? {
        "mock-token-\(userId)"
    }

    func hasCompletedParametersSetup(for userId: String) -> Bool {
        hasCompletedParametersSetup
    }
}
