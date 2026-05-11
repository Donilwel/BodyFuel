import Foundation
@testable import BodyFuel

final class MockWorkoutService: WorkoutServiceProtocol {

    // MARK: - Call tracking

    var generateWorkoutCallCount = 0
    var fetchWorkoutCallCount = 0
    var fetchWorkoutHistoryCallCount = 0
    var updateWorkoutCallCount = 0
    var deleteWorkoutCallCount = 0

    var lastFetchWorkoutId: String?
    var lastUpdateWorkoutId: String?
    var lastUpdateStatus: WorkoutStatus?
    var lastUpdateDuration: Int64?

    // MARK: - Configurable responses

    var generateWorkoutResult: Result<(workoutID: String, workout: Workout), Error> =
        .success(("test-id", .stub()))
    var fetchWorkoutResult: Result<(workoutID: String, workout: Workout), Error> =
        .success(("test-id", .stub()))
    var fetchWorkoutHistoryResult: Result<WorkoutHistoryResponseBody, Error> =
        .success(WorkoutHistoryResponseBody(workouts: [], total: 0, limit: 100, offset: 0))
    var updateWorkoutResult: Result<Void, Error> = .success(())
    var deleteWorkoutResult: Result<Void, Error> = .success(())

    // MARK: - Protocol

    func generateWorkout(place: WorkoutPlace?, type: ExerciseType?, level: WorkoutLevel?) async throws -> (workoutID: String, workout: Workout) {
        generateWorkoutCallCount += 1
        return try generateWorkoutResult.get()
    }

    func generateWorkout() async throws -> (workoutID: String, workout: Workout) {
        generateWorkoutCallCount += 1
        return try generateWorkoutResult.get()
    }

    func fetchWorkout(id: String) async throws -> (workoutID: String, workout: Workout) {
        fetchWorkoutCallCount += 1
        lastFetchWorkoutId = id
        return try fetchWorkoutResult.get()
    }

    func fetchWorkoutHistory(limit: Int, offset: Int) async throws -> WorkoutHistoryResponseBody {
        fetchWorkoutHistoryCallCount += 1
        return try fetchWorkoutHistoryResult.get()
    }

    func updateWorkout(id: String, status: WorkoutStatus?, duration: Int64?, totalCalories: Int?, exercises: [UpdateWorkoutExerciseItem]?) async throws {
        updateWorkoutCallCount += 1
        lastUpdateWorkoutId = id
        lastUpdateStatus = status
        lastUpdateDuration = duration
        _ = try updateWorkoutResult.get()
    }

    func updateWorkout(id: String, status: WorkoutStatus?, duration: Int64?) async throws {
        try await updateWorkout(id: id, status: status, duration: duration, totalCalories: nil, exercises: nil)
    }

    func deleteWorkout(id: String) async throws {
        deleteWorkoutCallCount += 1
        _ = try deleteWorkoutResult.get()
    }
}
