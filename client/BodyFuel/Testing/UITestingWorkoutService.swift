#if DEBUG
import Foundation

struct UITestingWorkoutService: WorkoutServiceProtocol {

    private var env: [String: String] { ProcessInfo.processInfo.environment }

    private static let stubWorkout = Workout(
        title: "Тестовая тренировка",
        type: .fullBody,
        duration: 1800,
        calories: 300,
        place: .gym,
        exercises: []
    )

    func generateWorkout() async throws -> (workoutID: String, workout: Workout) {
        return ("uitest-workout-id", Self.stubWorkout)
    }

    func generateWorkout(place: WorkoutPlace?, type: ExerciseType?, level: WorkoutLevel?) async throws -> (workoutID: String, workout: Workout) {
        return ("uitest-workout-id", Self.stubWorkout)
    }

    func fetchWorkout(id: String) async throws -> (workoutID: String, workout: Workout) {
        return (id, Self.stubWorkout)
    }

    func fetchWorkoutHistory(limit: Int, offset: Int) async throws -> WorkoutHistoryResponseBody {
        return WorkoutHistoryResponseBody(workouts: [], total: 0, limit: limit, offset: offset)
    }

    func updateWorkout(id: String, status: WorkoutStatus?, duration: Int64?, totalCalories: Int?, exercises: [UpdateWorkoutExerciseItem]?) async throws {}

    func updateWorkout(id: String, status: WorkoutStatus?, duration: Int64?) async throws {}

    func deleteWorkout(id: String) async throws {}
}
#endif
