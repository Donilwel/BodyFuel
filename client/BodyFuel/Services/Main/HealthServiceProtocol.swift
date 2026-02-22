import Foundation

protocol HealthServiceProtocol {
    func fetchTodaySteps() async throws -> Int
}

protocol NutritionServiceProtocol {
    func fetchTodayCalories() async throws -> Int
    func fetchTodayMeals() async throws -> [MealPreview]
}

protocol GoalsServiceProtocol {
    func fetchGoals() async throws -> GoalTargets
    func fetchTodayWorkout() async throws -> WorkoutPreview
}

final class MockHealthService: HealthServiceProtocol {
    func fetchTodaySteps() async throws -> Int { 6_540 }
}

final class MockNutritionService: NutritionServiceProtocol {
    func fetchTodayCalories() async throws -> Int { 1_420 }
    func fetchTodayMeals() async throws -> [MealPreview] {
        [
            MealPreview(title: "Завтрак", calories: 420),
            MealPreview(title: "Обед", calories: 650),
            MealPreview(title: "Перекус", calories: 350)
        ]
    }
}

final class MockGoalsService: GoalsServiceProtocol {
    func fetchGoals() async throws -> GoalTargets {
        GoalTargets(steps: 10_000, calories: 2_200)
    }

    func fetchTodayWorkout() async throws -> WorkoutPreview {
        WorkoutPreview(title: "Кардио + пресс", durationMinutes: 45, caloriesBurn: 380)
    }
}
