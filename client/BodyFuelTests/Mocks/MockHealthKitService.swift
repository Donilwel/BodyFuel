import Foundation
import HealthKit
@testable import BodyFuel

final class MockHealthKitService: HealthKitServiceProtocol {

    // MARK: - Call tracking

    var requestAuthorizationCallCount = 0
    var startWorkoutCallCount = 0
    var pauseWorkoutCallCount = 0
    var resumeWorkoutCallCount = 0
    var endWorkoutCallCount = 0
    var discardWorkoutCallCount = 0
    var refreshDailyActivityCallCount = 0
    var startBackgroundObserversCallCount = 0

    var lastStartedActivityType: HKWorkoutActivityType?

    // MARK: - Configurable responses

    var hasGrantedPermission: Bool = true
    var endWorkoutResult: (calories: Double, workout: HKWorkout?) = (250, nil)
    var fetchGenderResult: Result<HKBiologicalSex, Error> = .success(.female)
    var fetchDateOfBirthResult: Result<Date, Error> = .success(
        Calendar.current.date(byAdding: .year, value: -28, to: Date())!
    )
    var fetchTodayActiveCaloriesResult: Result<Double, Error> = .success(300)
    var fetchTodayStepsResult: Result<Int, Error> = .success(5000)
    var fetchDailyStepsResult: [DailySteps] = []

    // MARK: - Protocol

    func requestAuthorization() async {
        requestAuthorizationCallCount += 1
    }

    func fetchGender() throws -> HKBiologicalSex {
        try fetchGenderResult.get()
    }

    func fetchDateOfBirth() throws -> Date {
        try fetchDateOfBirthResult.get()
    }

    func fetchTodayActiveCalories() async throws -> Double {
        try fetchTodayActiveCaloriesResult.get()
    }

    func fetchTodaySteps() async throws -> Int {
        try fetchTodayStepsResult.get()
    }

    func fetchDailySteps(from startDate: Date, to endDate: Date) async -> [DailySteps] {
        fetchDailyStepsResult
    }

    func refreshDailyActivity() async {
        refreshDailyActivityCallCount += 1
    }

    func startBackgroundObservers() async {
        startBackgroundObserversCallCount += 1
    }

    func startWorkout(activityType: HKWorkoutActivityType) async {
        startWorkoutCallCount += 1
        lastStartedActivityType = activityType
    }

    func startWorkout() async {
        startWorkoutCallCount += 1
    }

    func pauseWorkout() {
        pauseWorkoutCallCount += 1
    }

    func resumeWorkout() {
        resumeWorkoutCallCount += 1
    }

    func endWorkout() async -> (calories: Double, workout: HKWorkout?) {
        endWorkoutCallCount += 1
        return endWorkoutResult
    }

    func discardWorkout() async {
        discardWorkoutCallCount += 1
    }
}
