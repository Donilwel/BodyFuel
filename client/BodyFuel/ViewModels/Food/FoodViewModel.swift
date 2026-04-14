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

    func load() async {
        screenState = .loading
        do {
            async let summary = nutritionService.fetchDailySummary()
            async let meals = nutritionService.fetchMeals()

            self.dailySummary = try await summary
            self.meals = try await meals
            screenState = .loaded
        } catch {
            screenState = .error("Не удалось загрузить данные питания")
        }
    }

    func addMealByText(description: String, mealType: MealType) async {
        guard !description.trimmingCharacters(in: .whitespaces).isEmpty else {
            addMealError = "Введите описание блюда"
            return
        }
        isAddingMeal = true
        addMealError = ""
        do {
            let meal = try await nutritionService.addMealByText(description: description, mealType: mealType)
            meals.append(meal)
            await refreshSummary()
            showAddMeal = false
        } catch {
            addMealError = "Не удалось добавить блюдо"
        }
        isAddingMeal = false
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
            meals.append(meal)
            await refreshSummary()
            showAddMeal = false
        } catch {
            addMealError = "Не удалось сохранить блюдо"
        }
    }

    func confirmAndSaveAnalyzedMeal(_ meal: Meal) async {
        meals.append(meal)
        await refreshSummary()
        showCamera = false
    }

    func loadRecipes() async {
        isLoadingRecipes = true
        do {
            recipes = try await nutritionService.generateRecipes()
        } catch {
            recipes = []
        }
        isLoadingRecipes = false
        showRecipes = true
    }

    private func refreshSummary() async {
        if let summary = try? await nutritionService.fetchDailySummary() {
            dailySummary = summary
        }
    }
}
