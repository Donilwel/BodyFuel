import Foundation

protocol NutritionServiceProtocol {
    func fetchTodayConsumedCalories() async throws -> Int
    func fetchTodayBurnedCalories() async throws -> Int
    func fetchTodayMeals() async throws -> [MealPreview]
}

final class NutritionService: NutritionServiceProtocol {
    static let shared = NutritionService()
    
    private let sharedWidgetStorage = SharedWidgetStorage.shared
    
    private init() {}
    
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
