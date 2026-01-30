import Foundation

protocol ProfileServiceProtocol {
    func fetchProfile() async throws -> UserProfile
    func updateProfile(_ profile: UserProfile) async throws
}

final class ProfileService: ProfileServiceProtocol {
    static let shared = ProfileService()
    
    private let networkClient = NetworkClient.shared
    
    private init() {}

    func fetchProfile() async throws -> UserProfile {
        return UserProfile(
            height: 165,
            photo: "http://localhost:9000/avatars/96805555-20a6-4cfc-b92f-e8d938d1dfa3",
            goal: .loseWeight,
            lifestyle: .active,
            currentWeight: 55,
            targetWeight: 50,
            targetCaloriesDaily: 1813,
            targetWorkoutsWeekly: 3
        )
    }

    func updateProfile(_ profile: UserProfile) async throws {
        
    }
}
