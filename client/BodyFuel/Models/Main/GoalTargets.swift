import Foundation

struct GoalTargets {
    let steps: Int
    let calories: Int
}

struct DayStats {
    let steps: Int
    let caloriesConsumed: Int
    let caloriesBurned: Int
}

struct WorkoutPreview {
    let title: String
    let durationMinutes: Int
    let caloriesBurn: Int
}

struct MealPreview: Identifiable {
    let id = UUID()
    let title: String
    let calories: Int
}
