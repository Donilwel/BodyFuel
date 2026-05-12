import Foundation
import Combine
@testable import BodyFuel

@MainActor
final class MockUserStore: UserStoreProtocol {

    // MARK: - Publishers

    private let profileSubject = CurrentValueSubject<UserProfile?, Never>(nil)

    var profilePublisher: AnyPublisher<UserProfile?, Never> { profileSubject.eraseToAnyPublisher() }
    var profile: UserProfile? { profileSubject.value }

    // MARK: - Call tracking

    var loadCallCount = 0
    var setTargetCaloriesCallCount = 0
    var setBasalMetabolicRateCallCount = 0
    var setProfileCallCount = 0

    var lastSetTargetCalories: Int?
    var lastSetBasalMetabolicRate: Int?
    var lastSetProfile: UserProfile?

    // MARK: - Helper to drive publisher

    func setProfileValue(_ profile: UserProfile?) {
        profileSubject.send(profile)
    }

    // MARK: - Protocol

    func load() async {
        loadCallCount += 1
    }

    func setTargetCalories(_ calories: Int) {
        setTargetCaloriesCallCount += 1
        lastSetTargetCalories = calories
    }

    func setBasalMetabolicRate(_ bmr: Int) {
        setBasalMetabolicRateCallCount += 1
        lastSetBasalMetabolicRate = bmr
    }

    func setProfile(_ updated: UserProfile) {
        setProfileCallCount += 1
        lastSetProfile = updated
        profileSubject.send(updated)
    }
}
