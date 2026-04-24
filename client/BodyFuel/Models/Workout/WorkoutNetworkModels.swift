import Foundation

// MARK: - Enums

enum WorkoutLevel: String, Encodable, CaseIterable {
    case light = "Лёгкий"
    case middle = "Средний"
    case hard = "Интенсивный"

    var apiValue: String {
        switch self {
        case .light: return "workout_light"
        case .middle: return "workout_middle"
        case .hard: return "workout_hard"
        }
    }
}

enum WorkoutStatus: String, Encodable {
    case created = "workout_created"
    case inProgress = "workout_in_active"
    case completed = "workout_done"
    case failed = "workout_failed"
}

// MARK: - Request Bodies

struct GenerateWorkoutRequestBody: Encodable {
    var placeExercise: String?
    var typeExercise: String?
    var level: String?
    var exercisesCount: Int?

    private enum CodingKeys: String, CodingKey {
        case placeExercise = "place_exercise"
        case typeExercise = "type_exercise"
        case level
        case exercisesCount = "exercises_count"
    }
}

struct UpdateWorkoutRequestBody: Encodable {
    let status: String?
    let duration: Int64?
}

// MARK: - Response Bodies

struct WorkoutResponseBody: Decodable {
    let id: String
    let userID: String
    let level: String
    let totalCalories: Int
    let predictionCalories: Int
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
        case predictionCalories = "prediction_calories"
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
    let modifyReps: Int
    let modifyRelaxTime: Int
    let calories: Int = 0
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
        case modifyReps = "modify_reps"
        case modifyRelaxTime = "modify_relax_time"
//        case calories
        case status
        case avgCaloriesPer = "avg_calories_per"
        case steps
        case completedAt = "completed_at"
    }
}

typealias WorkoutHistoryResponseBody = [WorkoutSummaryResponseBody]

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
