import Foundation
@testable import BodyFuel

final class MockNutritionService: NutritionServiceProtocol {

    // MARK: - Call tracking

    var fetchDailySummaryCallCount = 0
    var fetchMealsCallCount = 0
    var saveMealCallCount = 0
    var deleteFoodEntryCallCount = 0
    var analyzeMealCallCount = 0
    var generateRecipesCallCount = 0

    var lastSavedMeal: Meal?
    var lastDeletedMealId: String?
    var lastAnalyzedImageData: Data?
    var lastAnalyzedMealType: MealType?

    // MARK: - Configurable responses

    var fetchDailySummaryResult: Result<NutritionDailySummary, Error> = .success(.stub())
    var fetchMealsResult: Result<[Meal], Error> = .success([])
    var saveMealResult: Result<Void, Error> = .success(())
    var deleteFoodEntryResult: Result<Void, Error> = .success(())
    var analyzeMealResult: Result<Meal, Error> = .success(.stub())
    var generateRecipesResult: Result<[Recipe], Error> = .success([])

    // MARK: - Protocol (unused in ViewModel but required)

    func fetchTodayConsumedCalories() async throws -> Int { 0 }
    func fetchTodayBurnedCalories() async throws -> Int { 0 }
    func fetchTodayMeals() async throws -> [MealPreview] { [] }

    // MARK: - Protocol (used in ViewModels)

    func fetchDailySummary() async throws -> NutritionDailySummary {
        fetchDailySummaryCallCount += 1
        return try fetchDailySummaryResult.get()
    }

    func fetchMeals() async throws -> [Meal] {
        fetchMealsCallCount += 1
        return try fetchMealsResult.get()
    }

    func saveMeal(_ meal: Meal) async throws {
        saveMealCallCount += 1
        lastSavedMeal = meal
        _ = try saveMealResult.get()
    }

    func deleteFoodEntry(id: String) async throws {
        deleteFoodEntryCallCount += 1
        lastDeletedMealId = id
        _ = try deleteFoodEntryResult.get()
    }

    func analyzeMealFromPhoto(_ imageData: Data, mealType: MealType) async throws -> Meal {
        analyzeMealCallCount += 1
        lastAnalyzedImageData = imageData
        lastAnalyzedMealType = mealType
        return try analyzeMealResult.get()
    }

    func generateRecipes() async throws -> [Recipe] {
        generateRecipesCallCount += 1
        return try generateRecipesResult.get()
    }
}
