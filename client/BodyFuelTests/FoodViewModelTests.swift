import XCTest
import Combine
@testable import BodyFuel

@MainActor
final class FoodViewModelTests: XCTestCase {

    var mockNutritionService: MockNutritionService!
    var mockOffService: MockOpenFoodFactsService!
    var mockStore: MockNutritionStore!
    var sut: FoodViewModel!

    private var cancellables = Set<AnyCancellable>()

    override func setUp() async throws {
        mockNutritionService = MockNutritionService()
        mockOffService = MockOpenFoodFactsService()
        mockStore = MockNutritionStore()
        makeSUT()
    }

    override func tearDown() async throws {
        cancellables.removeAll()
        sut = nil
        mockStore = nil
        mockNutritionService = nil
        mockOffService = nil
    }

    private func makeSUT() {
        sut = FoodViewModel(
            nutritionService: mockNutritionService,
            offService: mockOffService,
            nutritionStore: mockStore
        )
    }

    private func drainTasks() async {
        await Task.yield()
        await Task.yield()
        await Task.yield()
    }

    func test_load_success_setsLoadedState() async throws {
        await sut.load()

        XCTAssertEqual(sut.screenState, .loaded)
        XCTAssertEqual(mockStore.loadCallCount, 1)
    }

    func test_load_triggersPreloadRecipesInBackground() async throws {
        let stubRecipes = [Recipe.stub(), Recipe.stub(name: "Омлет")]
        mockNutritionService.generateRecipesResult = .success(stubRecipes)

        await sut.load()
        await drainTasks()

        XCTAssertEqual(mockNutritionService.generateRecipesCallCount, 1)
        XCTAssertEqual(sut.recipes.count, 2)
    }

    func test_load_networkError_setsErrorState() async throws {
        mockStore.loadResult = .failure(NetworkError.network(URLError(.notConnectedToInternet)))

        await sut.load()

        if case .error(let msg) = sut.screenState {
            XCTAssertFalse(msg.isEmpty)
        } else {
            XCTFail("Expected .error state, got \(sut.screenState)")
        }
    }

    func test_load_authError_doesNotSetErrorState() async throws {
        mockStore.loadResult = .failure(NetworkError.missingToken)

        await sut.load()

        XCTAssertNotEqual(sut.screenState, .loaded)
        if case .error = sut.screenState {
            XCTFail("Auth errors should not set .error state")
        }
    }

    func test_preloadRecipes_onlyOncePerSession() async throws {
        mockNutritionService.generateRecipesResult = .success([.stub()])

        await sut.load()
        await drainTasks()

        XCTAssertEqual(mockNutritionService.generateRecipesCallCount, 1)

        await sut.load()
        await drainTasks()

        XCTAssertEqual(mockNutritionService.generateRecipesCallCount, 1, "Should not re-fetch recipes in same session")
    }

    func test_preloadRecipes_skippedIfRecipesAlreadyLoaded() async throws {
        sut.recipes = [.stub()]
        mockNutritionService.generateRecipesResult = .success([.stub(), .stub()])

        await sut.load()
        await drainTasks()

        XCTAssertEqual(mockNutritionService.generateRecipesCallCount, 0)
        XCTAssertEqual(sut.recipes.count, 1)
    }

    func test_loadRecipes_whenNotLoaded_fetchesAndShowsRecipes() async throws {
        let recipes = [Recipe.stub(), Recipe.stub(name: "Борщ")]
        mockNutritionService.generateRecipesResult = .success(recipes)

        await sut.loadRecipes()

        XCTAssertTrue(sut.showRecipes)
        XCTAssertEqual(sut.recipes.count, 2)
        XCTAssertFalse(sut.isLoadingRecipes)
        XCTAssertEqual(mockNutritionService.generateRecipesCallCount, 1)
    }

    func test_loadRecipes_whenAlreadyLoaded_justSetsShowRecipes() async throws {
        mockNutritionService.generateRecipesResult = .success([.stub()])
        await sut.loadRecipes()
        mockNutritionService.generateRecipesCallCount = 0

        await sut.loadRecipes()

        XCTAssertTrue(sut.showRecipes)
        XCTAssertEqual(mockNutritionService.generateRecipesCallCount, 0)
    }

    func test_loadRecipes_whenCurrentlyLoading_justSetsShowRecipes() async throws {
        sut.isLoadingRecipes = true

        await sut.loadRecipes()

        XCTAssertTrue(sut.showRecipes)
        XCTAssertEqual(mockNutritionService.generateRecipesCallCount, 0)
    }

    func test_loadRecipes_networkError_clearsRecipes() async throws {
        mockNutritionService.generateRecipesResult = .failure(NetworkError.network(URLError(.timedOut)))

        await sut.loadRecipes()

        XCTAssertEqual(sut.recipes, [])
        XCTAssertFalse(sut.isLoadingRecipes)
    }

    func test_saveMeal_success_dismissesAddMealSheet() async throws {
        sut.showAddMeal = true
        let meal = Meal.stub()

        await sut.saveMeal(meal)

        XCTAssertFalse(sut.showAddMeal)
        XCTAssertEqual(mockStore.addMealCallCount, 1)
        XCTAssertEqual(mockStore.lastAddedMeal?.id, meal.id)
    }

    func test_saveMeal_doubleTapPrevention_secondCallIgnored() async throws {
        sut.isAddingMeal = true

        await sut.saveMeal(.stub())

        XCTAssertEqual(mockStore.addMealCallCount, 0)
    }

    func test_saveMeal_error_setsAddMealError() async throws {
        mockStore.addMealResult = .failure(NetworkError.requestFailed(statusCode: 500, message: "Server error"))
        sut.showAddMeal = true

        await sut.saveMeal(.stub())

        XCTAssertFalse(sut.addMealError.isEmpty)
        XCTAssertTrue(sut.showAddMeal, "Sheet should remain open on error")
        XCTAssertFalse(sut.isAddingMeal, "isAddingMeal should be reset via defer")
    }

    func test_saveMeal_resetsIsAddingMealAfterCompletion() async throws {
        await sut.saveMeal(.stub())

        XCTAssertFalse(sut.isAddingMeal)
    }

    func test_confirmAndSaveAnalyzedMeal_success_dismissesCamera() async throws {
        sut.showCamera = true

        await sut.confirmAndSaveAnalyzedMeal(.stub())

        XCTAssertFalse(sut.showCamera)
        XCTAssertEqual(mockStore.addMealCallCount, 1)
    }

    func test_confirmAndSaveAnalyzedMeal_doubleTapPrevention() async throws {
        sut.isAddingMeal = true

        await sut.confirmAndSaveAnalyzedMeal(.stub())

        XCTAssertEqual(mockStore.addMealCallCount, 0)
    }

    func test_confirmAndSaveAnalyzedMeal_error_stillDismissesCamera() async throws {
        mockStore.addMealResult = .failure(NetworkError.requestFailed(statusCode: 422, message: "Validation failed"))
        sut.showCamera = true

        await sut.confirmAndSaveAnalyzedMeal(.stub())

        XCTAssertFalse(sut.showCamera, "Camera should be dismissed even on error")
        XCTAssertFalse(sut.addMealError.isEmpty)
    }

    func test_confirmAndSaveAnalyzedMeal_resetsIsAddingMeal() async throws {
        await sut.confirmAndSaveAnalyzedMeal(.stub())

        XCTAssertFalse(sut.isAddingMeal)
    }

    func test_analyzeMealFromPhoto_success_returnsMeal() async throws {
        let expected = Meal.stub(name: "Овсяная каша")
        mockNutritionService.analyzeMealResult = .success(expected)
        let data = Data("fake-image".utf8)

        let result = await sut.analyzeMealFromPhoto(data, mealType: .breakfast)

        XCTAssertEqual(result?.id, expected.id)
        XCTAssertEqual(mockNutritionService.analyzeMealCallCount, 1)
        XCTAssertEqual(mockNutritionService.lastAnalyzedMealType, .breakfast)
        XCTAssertFalse(sut.analysisNetworkError)
    }

    func test_analyzeMealFromPhoto_transportError_setsAnalysisNetworkError() async throws {
        mockNutritionService.analyzeMealResult = .failure(URLError(.notConnectedToInternet))

        let result = await sut.analyzeMealFromPhoto(Data(), mealType: .lunch)

        XCTAssertNil(result)
        XCTAssertTrue(sut.analysisNetworkError)
    }

    func test_analyzeMealFromPhoto_serverError_doesNotSetAnalysisNetworkError() async throws {
        mockNutritionService.analyzeMealResult = .failure(NetworkError.requestFailed(statusCode: 422, message: "Bad image"))

        let result = await sut.analyzeMealFromPhoto(Data(), mealType: .dinner)

        XCTAssertNil(result)
        XCTAssertFalse(sut.analysisNetworkError)
    }

    func test_analyzeMealFromPhoto_clearsNetworkErrorOnRetry() async throws {
        mockNutritionService.analyzeMealResult = .failure(URLError(.timedOut))
        _ = await sut.analyzeMealFromPhoto(Data(), mealType: .lunch)
        XCTAssertTrue(sut.analysisNetworkError)

        mockNutritionService.analyzeMealResult = .success(.stub())
        _ = await sut.analyzeMealFromPhoto(Data(), mealType: .lunch)
        XCTAssertFalse(sut.analysisNetworkError)
    }

    func test_deleteMeal_callsStore() async throws {
        let meal = Meal.stub()

        await sut.deleteMeal(meal)

        XCTAssertEqual(mockStore.deleteMealCallCount, 1)
        XCTAssertEqual(mockStore.lastDeletedMeal?.id, meal.id)
    }

    func test_mealsByType_groupsMealsByType() async throws {
        let breakfast1 = Meal.stub(mealType: .breakfast)
        let breakfast2 = Meal.stub(mealType: .breakfast)
        let dinner = Meal.stub(mealType: .dinner)
        sut.meals = [breakfast1, breakfast2, dinner]

        let result = sut.mealsByType

        XCTAssertEqual(result.count, 2)
        let breakfastGroup = result.first(where: { $0.0 == .breakfast })
        XCTAssertNotNil(breakfastGroup)
        XCTAssertEqual(breakfastGroup?.1.count, 2)
        let dinnerGroup = result.first(where: { $0.0 == .dinner })
        XCTAssertNotNil(dinnerGroup)
        XCTAssertEqual(dinnerGroup?.1.count, 1)
    }

    func test_mealsByType_excludesEmptyTypes() async throws {
        sut.meals = [Meal.stub(mealType: .lunch)]

        let result = sut.mealsByType

        XCTAssertEqual(result.count, 1)
        XCTAssertEqual(result.first?.0, .lunch)
    }

    func test_mealsByType_emptyMeals_returnsEmpty() async throws {
        sut.meals = []

        XCTAssertTrue(sut.mealsByType.isEmpty)
    }

    func test_mealsSubscription_updatesWhenStoreEmits() async throws {
        let meals = [Meal.stub(name: "Завтрак"), Meal.stub(name: "Обед")]

        mockStore.setMeals(meals)
        await drainTasks()

        XCTAssertEqual(sut.meals.count, 2)
    }

    func test_summarySubscription_updatesWhenStoreEmits() async throws {
        let summary = NutritionDailySummary.stub(burned: 500)

        mockStore.setSummary(summary)
        await drainTasks()

        XCTAssertEqual(sut.dailySummary?.burned, 500)
    }

    func test_summarySubscription_nilSummary_clearsProperty() async throws {
        mockStore.setSummary(.stub())
        await drainTasks()
        XCTAssertNotNil(sut.dailySummary)

        mockStore.setSummary(nil)
        await drainTasks()

        XCTAssertNil(sut.dailySummary)
    }

    func test_searchProducts_delegatesToOffService() async throws {
        let products = [FoodProduct.stub(name: "Молоко"), FoodProduct.stub(name: "Кефир")]
        mockOffService.searchResult = .success(products)

        let result = try await sut.searchProducts("молоко")

        XCTAssertEqual(result.count, 2)
        XCTAssertEqual(mockOffService.lastSearchQuery, "молоко")
        XCTAssertEqual(mockOffService.searchCallCount, 1)
    }

    func test_fetchProductByBarcode_delegatesToOffService() async throws {
        let product = FoodProduct.stub(code: "123456")
        mockOffService.barcodeResult = .success(product)

        let result = try await sut.fetchProductByBarcode("123456")

        XCTAssertEqual(result?.code, "123456")
        XCTAssertEqual(mockOffService.lastBarcode, "123456")
    }

    func test_fetchProductByBarcode_notFound_returnsNil() async throws {
        mockOffService.barcodeResult = .success(nil)

        let result = try await sut.fetchProductByBarcode("000")

        XCTAssertNil(result)
    }
}
