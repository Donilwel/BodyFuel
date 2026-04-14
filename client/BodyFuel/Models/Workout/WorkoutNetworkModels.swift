import Foundation

// MARK: - Enums

enum WorkoutLevel: String, Encodable {
    case beginner
    case intermediate
    case advanced
}

enum WorkoutStatus: String, Encodable {
    case pending
    case inProgress = "in_progress"
    case completed
    case cancelled
}

// MARK: - Request Bodies

struct GenerateWorkoutRequestBody: Encodable {
    let level: WorkoutLevel
    let duration: Int64?
}

struct UpdateWorkoutRequestBody: Encodable {
    let status: WorkoutStatus?
    let duration: Int64?

    func encode(to encoder: Encoder) throws {
        var container = encoder.container(keyedBy: CodingKeys.self)
        try container.encodeIfPresent(status, forKey: .status)
        try container.encodeIfPresent(duration, forKey: .duration)
    }

    private enum CodingKeys: String, CodingKey {
        case status
        case duration
    }
}

// MARK: - Response Bodies

struct WorkoutResponseBody: Decodable {
    let id: String
    let userID: String
    let level: String
    let totalCalories: Int
    let status: String
    let duration: Int64?
    let createdAt: String
    let updatedAt: String
    let exercises: [WorkoutExerciseResponseBody]?

    private enum CodingKeys: String, CodingKey {
        case id
        case userID = "user_id"
        case level
        case totalCalories = "total_calories"
        case status
        case duration
        case createdAt = "created_at"
        case updatedAt = "updated_at"
        case exercises
    }
}

struct WorkoutExerciseResponseBody: Decodable {
    let exerciseID: String
    let name: String
    let description: String
    let typeExercise: String
    let placeExercise: String
    let levelPreparation: String
    let linkGif: String
    let baseCountReps: Int
    let baseRelaxTime: Int
    let modifyReps: Int
    let modifyRelaxTime: Int
    let calories: Int
    let status: String
    let avgCaloriesPer: Double
    let steps: Int
    let completedAt: String?

    private enum CodingKeys: String, CodingKey {
        case exerciseID = "exercise_id"
        case name
        case description
        case typeExercise = "type_exercise"
        case placeExercise = "place_exercise"
        case levelPreparation = "level_preparation"
        case linkGif = "link_gif"
        case baseCountReps = "base_count_reps"
        case baseRelaxTime = "base_relax_time"
        case modifyReps = "modify_reps"
        case modifyRelaxTime = "modify_relax_time"
        case calories
        case status
        case avgCaloriesPer = "avg_calories_per"
        case steps
        case completedAt = "completed_at"
    }
}

struct WorkoutHistoryResponseBody: Decodable {
    let workouts: [WorkoutSummaryResponseBody]
    let total: Int
    let limit: Int
    let offset: Int
}

struct WorkoutSummaryResponseBody: Decodable {
    let id: String
    let level: String
    let totalCalories: Int
    let status: String
    let duration: Int64?
    let createdAt: String
    let exercisesCount: Int
    let completedCount: Int

    private enum CodingKeys: String, CodingKey {
        case id
        case level
        case totalCalories = "total_calories"
        case status
        case duration
        case createdAt = "created_at"
        case exercisesCount = "exercises_count"
        case completedCount = "completed_count"
    }
}
