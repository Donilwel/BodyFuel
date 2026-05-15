import Foundation
import Combine
import UIKit

@MainActor
final class FoodViewModel: ObservableObject {
    @Published var screenState: ScreenState = .loading
    @Published var dailySummary: NutritionDailySummary?
    @Published var meals: [Meal] = []

    @Published var showAddMeal = false
    @Published var showCamera = false
    @Published var showRecipes = false
    @Published var galleryImage: UIImage? = nil

    @Published var recipes: [Recipe] = []
    @Published var isLoadingRecipes = false

    private var recipesSessionLoaded = false

    @Published var addMealType: MealType = .breakfast
    @Published var isAddingMeal = false
    @Published var addMealError: String = ""
    @Published var analysisNetworkError = false

    var mealsByType: [(MealType, [Meal])] {
        MealType.allCases.compactMap { type in
            let group = meals.filter { $0.mealType == type }
            guard !group.isEmpty else { return nil }
            return (type, group)
        }
    }

    private let nutritionService: NutritionServiceProtocol
    private let offService: OpenFoodFactsServiceProtocol
    private let nutritionStore: NutritionStoreProtocol

    private var mealsCancellable: AnyCancellable?
    private var summaryCancellable: AnyCancellable?

    init() {
        self.nutritionService = NutritionService.shared
        self.offService = OpenFoodFactsService.shared
        self.nutritionStore = NutritionStore.shared
        setupSubscriptions()
    }

    init(
        nutritionService: NutritionServiceProtocol,
        offService: OpenFoodFactsServiceProtocol,
        nutritionStore: NutritionStoreProtocol
    ) {
        self.nutritionService = nutritionService
        self.offService = offService
        self.nutritionStore = nutritionStore
        setupSubscriptions()
    }

    private func setupSubscriptions() {
        mealsCancellable = nutritionStore.mealsPublisher
            .receive(on: RunLoop.main)
            .sink { [weak self] newMeals in
                self?.meals = newMeals
            }

        summaryCancellable = nutritionStore.dailySummaryPublisher
            .receive(on: RunLoop.main)
            .sink { [weak self] summary in
                self?.dailySummary = summary
            }
    }

    func load() async {
        screenState = .loading
        do {
            try await nutritionStore.load()
            screenState = .loaded
        } catch {
            if AppRouter.shared.handleIfUnauthorized(error) { return }
            screenState = .error("Не удалось загрузить данные питания")
            return
        }
        Task { await preloadRecipesIfNeeded() }
    }

    private func preloadRecipesIfNeeded() async {
        guard !recipesSessionLoaded, !isLoadingRecipes, recipes.isEmpty else { return }
        isLoadingRecipes = true
        do {
            recipes = try await nutritionService.generateRecipes()
            recipesSessionLoaded = true
        } catch {
            print("[ERROR] [FoodViewModel/preloadRecipesIfNeeded]: Error preloading recipes: \(error)")
        }
        isLoadingRecipes = false
    }

    func analyzeMealFromPhoto(_ imageData: Data, mealType: MealType) async -> Meal? {
        analysisNetworkError = false
        do {
            let meal = try await nutritionService.analyzeMealFromPhoto(imageData, mealType: mealType)
            return meal
        } catch {
            analysisNetworkError = isTransportError(error)
            return nil
        }
    }

    func saveMeal(_ meal: Meal) async {
        guard !isAddingMeal else { return }
        isAddingMeal = true
        defer { isAddingMeal = false }
        do {
            try await nutritionStore.addMeal(meal)
            HapticService.notification(.success)
            showAddMeal = false
        } catch {
            if AppRouter.shared.handleIfUnauthorized(error) { return }
            HapticService.notification(.error)
            addMealError = "Не удалось сохранить блюдо"
        }
    }

    func confirmAndSaveAnalyzedMeal(_ meal: Meal) async {
        guard !isAddingMeal else { return }
        isAddingMeal = true
        defer { isAddingMeal = false }
        do {
            try await nutritionStore.addMeal(meal)
            HapticService.notification(.success)
        } catch {
            if AppRouter.shared.handleIfUnauthorized(error) { return }
            HapticService.notification(.error)
            addMealError = "Не удалось сохранить блюдо"
        }
        showCamera = false
    }

    func deleteMeal(_ meal: Meal) async {
        await nutritionStore.deleteMeal(meal)
    }

    func loadRecipes() async {
        if recipesSessionLoaded || isLoadingRecipes {
            showRecipes = true
            return
        }
        isLoadingRecipes = true
        showRecipes = true
        do {
            recipes = try await nutritionService.generateRecipes()
            recipesSessionLoaded = true
        } catch {
            if AppRouter.shared.handleIfUnauthorized(error) { return }
            recipes = []
        }
        isLoadingRecipes = false
    }

    func searchProducts(_ query: String) async throws -> [FoodProduct] {
        try await offService.searchProducts(query: query)
    }

    func fetchProductByBarcode(_ barcode: String) async throws -> FoodProduct? {
        try await offService.fetchProductByBarcode(barcode)
    }
}
