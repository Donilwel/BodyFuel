import Foundation
@testable import BodyFuel

final class MockProfileService: ProfileServiceProtocol {

    // MARK: - Call tracking

    var fetchProfileCallCount = 0
    var updateProfileCallCount = 0
    var deleteProfileCallCount = 0

    var lastUpdatedProfile: UserProfile?

    // MARK: - Configurable responses

    var fetchProfileResult: Result<UserProfile, Error> = .success(.stub())
    var updateProfileResult: Result<Void, Error> = .success(())
    var deleteProfileResult: Result<Void, Error> = .success(())

    // MARK: - Protocol

    func fetchProfile() async throws -> UserProfile {
        fetchProfileCallCount += 1
        return try fetchProfileResult.get()
    }

    func updateProfile(_ profile: UserProfile) async throws {
        updateProfileCallCount += 1
        lastUpdatedProfile = profile
        _ = try updateProfileResult.get()
    }

    func deleteProfile() async throws {
        deleteProfileCallCount += 1
        _ = try deleteProfileResult.get()
    }
}
