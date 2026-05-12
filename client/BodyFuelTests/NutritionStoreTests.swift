import XCTest
@testable import BodyFuel

@MainActor
final class NutritionStoreTests: XCTestCase {

    var sut: NutritionStore!

    private let cacheKey = "nutrition_meals_anon"

    override func setUp() async throws {
        sut = NutritionStore.shared
        sut.reset()
        MutationQueue.shared.clear()
        DiskCache.shared.remove(key: cacheKey)
        NetworkMonitor.shared.markServerReachable()
    }

    override func tearDown() async throws {
        sut.reset()
        MutationQueue.shared.clear()
        DiskCache.shared.remove(key: cacheKey)
        NetworkMonitor.shared.markServerReachable()
    }

    // MARK: - addMeal

    func test_addMeal_appendsMealToArray() async throws {
        let meal = Meal.stub()
        try await sut.addMeal(meal)
        XCTAssertTrue(sut.meals.contains { $0.id == meal.id })
    }

    func test_addMeal_updatesDailySummaryMacros() async throws {
        let meal = Meal.stub(macros: .stub(protein: 40, fat: 10, carbs: 80))
        try await sut.addMeal(meal)
        XCTAssertEqual(sut.dailySummary?.consumed.protein ?? 0, 40, accuracy: 0.001)
        XCTAssertEqual(sut.dailySummary?.consumed.fat ?? 0, 10, accuracy: 0.001)
        XCTAssertEqual(sut.dailySummary?.consumed.carbs ?? 0, 80, accuracy: 0.001)
    }

    func test_addMeal_accumulatesMacros_forMultipleMeals() async throws {
        let meal1 = Meal.stub(macros: .stub(protein: 30, fat: 5, carbs: 50))
        let meal2 = Meal.stub(macros: .stub(protein: 20, fat: 8, carbs: 30))
        try await sut.addMeal(meal1)
        try await sut.addMeal(meal2)
        XCTAssertEqual(sut.dailySummary?.consumed.protein ?? 0, 50, accuracy: 0.001)
    }

    func test_addMeal_updatesMealPreviews() async throws {
        let meal = Meal.stub(mealType: .lunch)
        try await sut.addMeal(meal)
        XCTAssertFalse(sut.mealPreviews.isEmpty)
    }

    func test_addMeal_mealPreview_reflectsCorrectMealType() async throws {
        let meal = Meal.stub(mealType: .breakfast)
        try await sut.addMeal(meal)
        XCTAssertTrue(sut.mealPreviews.contains { $0.title == MealType.breakfast.displayName })
    }

    func test_addMeal_duplicate_doesNotAddTwice() async throws {
        let meal = Meal.stub()
        try await sut.addMeal(meal)
        try await sut.addMeal(meal)
        let count = sut.meals.filter { $0.id == meal.id }.count
        XCTAssertEqual(count, 1)
    }

    func test_addMeal_duplicate_doesNotUpdateSummaryTwice() async throws {
        let meal = Meal.stub(macros: .stub(protein: 30, fat: 0, carbs: 0))
        try await sut.addMeal(meal)
        try await sut.addMeal(meal)
        XCTAssertEqual(sut.dailySummary?.consumed.protein ?? 0, 30, accuracy: 0.001)
    }

    func test_addMeal_offline_addsMealToArray() async throws {
        NetworkMonitor.shared.markServerUnreachable()
        let meal = Meal.stub()
        try await sut.addMeal(meal)
        XCTAssertTrue(sut.meals.contains { $0.id == meal.id })
    }

    func test_addMeal_offline_enqueuesAddMealMutation() async throws {
        NetworkMonitor.shared.markServerUnreachable()
        let meal = Meal.stub()
        try await sut.addMeal(meal)
        XCTAssertEqual(MutationQueue.shared.mutations.count, 1)
        XCTAssertEqual(MutationQueue.shared.mutations.first?.type, .addMeal)
    }

    func test_addMeal_offline_twoMeals_enqueueTwoMutations() async throws {
        NetworkMonitor.shared.markServerUnreachable()
        try await sut.addMeal(Meal.stub())
        try await sut.addMeal(Meal.stub())
        XCTAssertEqual(MutationQueue.shared.mutations.count, 2)
    }

    func test_addMeal_online_doesNotEnqueueMutation_immediately() async throws {
        let meal = Meal.stub()
        try await sut.addMeal(meal)
        XCTAssertEqual(MutationQueue.shared.mutations.count, 0)
    }

    // MARK: - deleteMeal

    func test_deleteMeal_removesFromArray() async throws {
        let meal = Meal.stub()
        try await sut.addMeal(meal)
        await sut.deleteMeal(meal)
        XCTAssertFalse(sut.meals.contains { $0.id == meal.id })
    }

    func test_deleteMeal_updatesDailySummary_toZero() async throws {
        let meal = Meal.stub(macros: .stub(protein: 40, fat: 10, carbs: 80))
        try await sut.addMeal(meal)
        await sut.deleteMeal(meal)
        XCTAssertEqual(sut.dailySummary?.consumed.protein ?? 0, 0, accuracy: 0.001)
    }

    func test_deleteMeal_updatesMealPreviews_toEmpty() async throws {
        let meal = Meal.stub(mealType: .breakfast)
        try await sut.addMeal(meal)
        await sut.deleteMeal(meal)
        XCTAssertTrue(sut.mealPreviews.isEmpty)
    }

    func test_deleteMeal_nonExistentMeal_noChange() async throws {
        let meal = Meal.stub()
        try await sut.addMeal(meal)
        let other = Meal.stub()
        await sut.deleteMeal(other)
        XCTAssertEqual(sut.meals.count, 1)
    }

    func test_deleteMeal_updatesOnlyAffectedSummaryMacros() async throws {
        let meal1 = Meal.stub(macros: .stub(protein: 30, fat: 5, carbs: 0))
        let meal2 = Meal.stub(macros: .stub(protein: 20, fat: 0, carbs: 60))
        try await sut.addMeal(meal1)
        try await sut.addMeal(meal2)
        await sut.deleteMeal(meal1)
        XCTAssertEqual(sut.dailySummary?.consumed.protein ?? 0, 20, accuracy: 0.001)
        XCTAssertEqual(sut.dailySummary?.consumed.carbs ?? 0, 60, accuracy: 0.001)
    }

    func test_deleteMeal_offline_enqueuesDeleteMealMutation() async throws {
        let meal = Meal.stub()
        sut.meals = [meal]
        NetworkMonitor.shared.markServerUnreachable()
        await sut.deleteMeal(meal)
        XCTAssertEqual(MutationQueue.shared.mutations.count, 1)
        XCTAssertEqual(MutationQueue.shared.mutations.first?.type, .deleteMeal)
    }

    func test_deleteMeal_offline_removesFromArrayImmediately() async throws {
        let meal = Meal.stub()
        sut.meals = [meal]
        NetworkMonitor.shared.markServerUnreachable()
        await sut.deleteMeal(meal)
        XCTAssertTrue(sut.meals.isEmpty)
    }

    func test_deleteMeal_online_doesNotEnqueueMutation_immediately() async throws {
        let meal = Meal.stub()
        try await sut.addMeal(meal)
        MutationQueue.shared.clear()
        await sut.deleteMeal(meal)
        XCTAssertEqual(MutationQueue.shared.mutations.count, 0)
    }

    // MARK: - load 

    func test_load_offline_loadsFromDiskCache() async throws {
        let meals = [Meal.stub(name: "Cached Pasta")]
        DiskCache.shared.save(meals, key: cacheKey)
        NetworkMonitor.shared.markServerUnreachable()

        try await sut.load()

        XCTAssertTrue(sut.meals.contains { $0.name == "Cached Pasta" })
    }

    func test_load_offline_setsIsDataStale() async throws {
        DiskCache.shared.save([Meal.stub()], key: cacheKey)
        NetworkMonitor.shared.markServerUnreachable()

        try await sut.load()

        XCTAssertTrue(sut.isDataStale)
    }

    func test_load_offline_derivesDailySummaryFromCachedMeals() async throws {
        let meal = Meal.stub(macros: .stub(protein: 55, fat: 15, carbs: 90))
        DiskCache.shared.save([meal], key: cacheKey)
        NetworkMonitor.shared.markServerUnreachable()

        try await sut.load()

        XCTAssertEqual(sut.dailySummary?.consumed.protein ?? 0, 55, accuracy: 0.001)
    }

    func test_load_offline_emptyOrMissingCache_mealsRemainsEmpty() async throws {
        DiskCache.shared.remove(key: cacheKey)
        NetworkMonitor.shared.markServerUnreachable()

        try await sut.load()

        XCTAssertTrue(sut.meals.isEmpty)
    }

    func test_load_offline_buildsMealPreviews_fromCachedMeals() async throws {
        let meal = Meal.stub(mealType: .dinner)
        DiskCache.shared.save([meal], key: cacheKey)
        NetworkMonitor.shared.markServerUnreachable()

        try await sut.load()

        XCTAssertFalse(sut.mealPreviews.isEmpty)
    }

    // MARK: - reset

    func test_reset_clearsAllState() async throws {
        try await sut.addMeal(Meal.stub())

        sut.reset()

        XCTAssertTrue(sut.meals.isEmpty)
        XCTAssertNil(sut.dailySummary)
        XCTAssertTrue(sut.mealPreviews.isEmpty)
        XCTAssertFalse(sut.isDataStale)
    }

    func test_reset_removesDiskCache() async throws {
        DiskCache.shared.save([Meal.stub()], key: cacheKey)

        sut.reset()

        let cached = DiskCache.shared.load([Meal].self, key: cacheKey)
        XCTAssertNil(cached)
    }
}
