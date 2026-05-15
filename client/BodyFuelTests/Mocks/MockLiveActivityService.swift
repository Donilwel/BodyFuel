import Foundation
@testable import BodyFuel

final class MockLiveActivityService: LiveActivityServiceProtocol {

    // MARK: - Call tracking

    var startCallCount = 0
    var updateCallCount = 0
    var endCallCount = 0

    var lastStartedWorkoutName: String?
    var lastStartedExerciseName: String?

    // MARK: - Protocol

    func start(workoutName: String, exerciseName: String, exerciseType: ExerciseType) {
        startCallCount += 1
        lastStartedWorkoutName = workoutName
        lastStartedExerciseName = exerciseName
    }

    func update(
        exerciseName: String,
        exerciseType: ExerciseType,
        exerciseDuration: Int,
        workoutPhase: WorkoutPhase,
        workoutProgress: Double
    ) {
        updateCallCount += 1
    }

    func end() {
        endCallCount += 1
    }
}
