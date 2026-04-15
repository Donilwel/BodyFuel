import Foundation
import Combine

@MainActor
final class FoodViewModel: ObservableObject {
    enum ScreenState {
        case loading
        case loaded
        case error(String)
    }

    @Published var screenState: ScreenState = .loading
    @Published var dailySummary: NutritionDailySummary?
    @Published var meals: [Meal] = []

    @Published var showAddMeal = false
    @Published var showCamera = false
    @Published var showRecipes = false

    @Published var recipes: [Recipe] = []
    @Published var isLoadingRecipes = false

    @Published var addMealType: MealType = .breakfast
    @Published var isAddingMeal = false
    @Published var addMealError: String = ""

    var mealsByType: [(MealType, [Meal])] {
        MealType.allCases.compactMap { type in
            let group = meals.filter { $0.mealType == type }
            guard !group.isEmpty else { return nil }
            return (type, group)
        }
    }

    private let nutritionService: NutritionServiceProtocol = NutritionService.shared
    private let offService: OpenFoodFactsServiceProtocol = OpenFoodFactsService.shared

    func load() async {
        screenState = .loading
        do {
            async let summary = nutritionService.fetchDailySummary()
            async let meals = nutritionService.fetchMeals()

            self.dailySummary = try await summary
            self.meals = try await meals
            screenState = .loaded
        } catch {
            if AppRouter.shared.handleIfUnauthorized(error) { return }
            screenState = .error("Не удалось загрузить данные питания")
        }
    }

    func analyzeMealFromPhoto(_ imageData: Data, mealType: MealType) async -> Meal? {
        do {
            let meal = try await nutritionService.analyzeMealFromPhoto(imageData, mealType: mealType)
            return meal
        } catch {
            return nil
        }
    }

    func saveMeal(_ meal: Meal) async {
        do {
            try await nutritionService.saveMeal(meal)
            await load()
            showAddMeal = false
        } catch {
            if AppRouter.shared.handleIfUnauthorized(error) { return }
            addMealError = "Не удалось сохранить блюдо"
        }
    }

    func confirmAndSaveAnalyzedMeal(_ meal: Meal) async {
        do {
            try await nutritionService.saveMeal(meal)
            await load()
        } catch {
            if AppRouter.shared.handleIfUnauthorized(error) { return }
            addMealError = "Не удалось сохранить блюдо"
        }
        showCamera = false
    }

    func deleteMeal(_ meal: Meal) async {
        do {
            try await nutritionService.deleteFoodEntry(id: meal.id.uuidString)
            meals.removeAll { $0.id == meal.id }
            await refreshSummary()
        } catch {
            if AppRouter.shared.handleIfUnauthorized(error) { return }
            screenState = .error("Не удалось удалить блюдо")
        }
    }

    func loadRecipes() async {
        isLoadingRecipes = true
        do {
            recipes = try await nutritionService.generateRecipes()
        } catch {
            if AppRouter.shared.handleIfUnauthorized(error) { return }
            recipes = []
        }
        isLoadingRecipes = false
        showRecipes = true
    }

    func searchProducts(_ query: String) async throws -> [FoodProduct] {
        try await offService.searchProducts(query: query)
    }

    func fetchProductByBarcode(_ barcode: String) async throws -> FoodProduct? {
        try await offService.fetchProductByBarcode(barcode)
    }

    private func refreshSummary() async {
        if let summary = try? await nutritionService.fetchDailySummary() {
            dailySummary = summary
        }
    }
}
