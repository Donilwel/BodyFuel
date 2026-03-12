import Foundation

protocol HealthServiceProtocol {
    func fetchTodaySteps() async throws -> Int
}

protocol NutritionServiceProtocol {
    func fetchTodayConsumedCalories() async throws -> Int
    func fetchTodayBurnedCalories() async throws -> Int
    func fetchTodayMeals() async throws -> [MealPreview]
}

protocol GoalsServiceProtocol {
//    func fetchGoals() async throws -> GoalTargets
    func fetchTodayWorkout() async throws -> WorkoutModel
}

final class MockHealthService: HealthServiceProtocol {
    func fetchTodaySteps() async throws -> Int { 6540 }
}

final class MockNutritionService: NutritionServiceProtocol {
    private let sharedWidgetStorage = SharedWidgetStorage.shared
    
    func fetchTodayConsumedCalories() async throws -> Int {
        sharedWidgetStorage.saveTodayConsumedCalories(1600)
        return 1600
    }
    
    func fetchTodayBurnedCalories() async throws -> Int {
        sharedWidgetStorage.saveTodayBurnedCalories(345)
        return 345
    }
    
    func fetchTodayMeals() async throws -> [MealPreview] {
        [
            MealPreview(title: "Завтрак", calories: 420),
            MealPreview(title: "Обед", calories: 650),
            MealPreview(title: "Перекус", calories: 350)
        ]
    }
}

final class MockGoalsService: GoalsServiceProtocol {
    private let sharedWidgetStorage = SharedWidgetStorage.shared
    
//    func fetchGoals() async throws -> GoalTargets {
////        sharedWidgetStorage.saveTargetCalories(2437)
//        return GoalTargets(
//            steps: 10000,
//            calories: sharedWidgetStorage.getTargetCalories()
//        )
//    }

    func fetchTodayWorkout() async throws -> WorkoutModel {
        let workout = WorkoutModel(
            name: "Кардио + пресс",
            duration: 45,
            calories: 320,
            location: "Зал",
            type: "Кардио"
        )
        
        sharedWidgetStorage.saveWorkout(workout)
        
        return workout
    }
}
