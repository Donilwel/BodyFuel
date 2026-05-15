import Foundation
@testable import BodyFuel

final class MockSharedWidgetStorage: SharedWidgetStorageProtocol {

    // MARK: - Call tracking

    var saveWorkoutCallCount = 0
    var saveTodayWorkoutDoneCallCount = 0

    // MARK: - Captured values

    var lastSavedWorkout: WorkoutModel?
    var savedWorkoutWasNil = false
    var lastSavedWorkoutName: String?
    var lastSavedWorkoutDoneValue: Bool?

    // MARK: - Protocol

    func saveWorkout(_ workout: WorkoutModel?) {
        saveWorkoutCallCount += 1
        lastSavedWorkout = workout
        if let workout {
            savedWorkoutWasNil = false
            lastSavedWorkoutName = workout.name
        } else {
            savedWorkoutWasNil = true
        }
    }

    func saveTodayWorkoutDone(_ done: Bool) {
        saveTodayWorkoutDoneCallCount += 1
        lastSavedWorkoutDoneValue = done
    }

    func isTodayWorkoutDone() -> Bool { false }
}
