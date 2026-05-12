import Foundation
import Combine
@testable import BodyFuel

@MainActor
final class MockUserStore: UserStoreProtocol {

    // MARK: - Publishers (backing subjects)

    private let profileSubject = CurrentValueSubject<UserProfile?, Never>(nil)
    private let targetCaloriesSubject = CurrentValueSubject<Int, Never>(0)
    private let caloriesBurnedSubject = CurrentValueSubject<Int, Never>(0)
    private let todayStepsSubject = CurrentValueSubject<Int, Never>(0)

    // MARK: - Protocol properties

    var profile: UserProfile? { profileSubject.value }
    var targetCalories: Int { targetCaloriesSubject.value }
    var caloriesBurned: Int { caloriesBurnedSubject.value }
    var todaySteps: Int { todayStepsSubject.value }
    var basalMetabolicRate: Int = 0

    var profilePublisher: AnyPublisher<UserProfile?, Never> { profileSubject.eraseToAnyPublisher() }
    var targetCaloriesPublisher: AnyPublisher<Int, Never> { targetCaloriesSubject.eraseToAnyPublisher() }
    var caloriesBurnedPublisher: AnyPublisher<Int, Never> { caloriesBurnedSubject.eraseToAnyPublisher() }
    var todayStepsPublisher: AnyPublisher<Int, Never> { todayStepsSubject.eraseToAnyPublisher() }

    // MARK: - Call tracking

    var loadCallCount = 0
    var setTargetCaloriesCallCount = 0
    var setBasalMetabolicRateCallCount = 0
    var setProfileCallCount = 0

    var lastSetTargetCalories: Int?
    var lastSetBasalMetabolicRate: Int?
    var lastSetProfile: UserProfile?

    // MARK: - Helpers to drive publishers

    func setProfileValue(_ profile: UserProfile?) {
        profileSubject.send(profile)
    }

    func setTargetCaloriesValue(_ value: Int) {
        targetCaloriesSubject.send(value)
    }

    func setCaloriesBurned(_ value: Int) {
        caloriesBurnedSubject.send(value)
    }

    func setTodaySteps(_ value: Int) {
        todayStepsSubject.send(value)
    }

    // MARK: - Protocol methods

    func load() async {
        loadCallCount += 1
    }

    func setTargetCalories(_ calories: Int) {
        setTargetCaloriesCallCount += 1
        lastSetTargetCalories = calories
        targetCaloriesSubject.send(calories)
    }

    func setBasalMetabolicRate(_ bmr: Int) {
        setBasalMetabolicRateCallCount += 1
        lastSetBasalMetabolicRate = bmr
        basalMetabolicRate = bmr
    }

    func setProfile(_ updated: UserProfile) {
        setProfileCallCount += 1
        lastSetProfile = updated
        profileSubject.send(updated)
    }
}
