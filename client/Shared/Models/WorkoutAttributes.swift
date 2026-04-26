import ActivityKit

struct WorkoutAttributes: ActivityAttributes {
    public struct ContentState: Codable, Hashable {
        var exerciseName: String
        var exerciseType: ExerciseType
        var exerciseDuration: Int
        var workoutPhase: WorkoutPhase
        var workoutProgress: Double
    }
    
    var workoutName: String
}
