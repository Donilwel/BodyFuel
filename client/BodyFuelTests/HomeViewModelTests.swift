import XCTest
import Combine
@testable import BodyFuel

@MainActor
final class HomeViewModelTests: XCTestCase {

    var mockNutritionStore: MockNutritionStore!
    var mockUserStore: MockUserStore!
    var mockHistoryStore: MockWorkoutHistoryStore!
    var mockHealthKit: MockHealthKitService!
    var sut: HomeViewModel!

    override func setUp() async throws {
        mockNutritionStore = MockNutritionStore()
        mockUserStore = MockUserStore()
        mockHistoryStore = MockWorkoutHistoryStore()
        mockHealthKit = MockHealthKitService()
        makeSUT()
    }

    override func tearDown() async throws {
        sut = nil
        mockNutritionStore = nil
        mockUserStore = nil
        mockHistoryStore = nil
        mockHealthKit = nil
    }

    private func makeSUT() {
        sut = HomeViewModel(
            nutritionStore: mockNutritionStore,
            userStore: mockUserStore,
            workoutHistoryStore: mockHistoryStore,
            health: mockHealthKit
        )
    }

    private func drainTasks() async {
        await Task.yield()
        await Task.yield()
        await Task.yield()
    }

    func test_load_callsAllThreeStores() async throws {
        await sut.load()

        XCTAssertEqual(mockNutritionStore.loadCallCount, 1)
        XCTAssertEqual(mockUserStore.loadCallCount, 1)
        XCTAssertEqual(mockHistoryStore.loadCallCount, 1)
    }

    func test_load_callsHealthKitRefresh() async throws {
        await sut.load()

        XCTAssertEqual(mockHealthKit.refreshDailyActivityCallCount, 1)
    }

    func test_load_success_setsLoadedState() async throws {
        await sut.load()

        XCTAssertEqual(sut.state, .loaded)
    }

    func test_load_populatesStats_withUserStoreValues() async throws {
        mockUserStore.setTodaySteps(8500)
        mockUserStore.setCaloriesBurned(400)
        let summary = NutritionDailySummary.stub(consumed: .stub(protein: 50, fat: 20, carbs: 100))
        mockNutritionStore.setSummary(summary)

        await sut.load()

        XCTAssertEqual(sut.stats?.steps, 8500)
        XCTAssertEqual(sut.stats?.caloriesBurned, 400)
        XCTAssertEqual(sut.stats?.caloriesConsumed, summary.consumed.calories)
    }

    func test_load_stats_usesZeroCaloriesConsumed_whenNoSummary() async throws {
        await sut.load()

        XCTAssertEqual(sut.stats?.caloriesConsumed, 0)
    }

    func test_load_populatesGoals_withTargetCalories() async throws {
        mockUserStore.setTargetCaloriesValue(2500)

        await sut.load()

        XCTAssertEqual(sut.goals?.calories, 2500)
        XCTAssertEqual(sut.goals?.steps, 10000)
    }

    func test_load_populatesBasalMetabolicRate_whenPositive() async throws {
        mockUserStore.basalMetabolicRate = 1600

        await sut.load()

        XCTAssertEqual(sut.basalMetabolicRate, 1600)
    }

    func test_load_basalMetabolicRateIsNil_whenZero() async throws {
        mockUserStore.basalMetabolicRate = 0

        await sut.load()

        XCTAssertNil(sut.basalMetabolicRate)
    }

    func test_load_nutritionError_setsErrorState() async throws {
        mockNutritionStore.loadResult = .failure(NetworkError.requestFailed(statusCode: 500, message: "Server error"))

        await sut.load()

        if case .error(let msg) = sut.state {
            XCTAssertFalse(msg.isEmpty)
        } else {
            XCTFail("Expected .error state, got \(sut.state)")
        }
    }

    func test_load_nutritionError_doesNotCallUserOrHistoryStore() async throws {
        mockNutritionStore.loadResult = .failure(NetworkError.requestFailed(statusCode: 500, message: ""))

        await sut.load()

        XCTAssertEqual(mockUserStore.loadCallCount, 0)
        XCTAssertEqual(mockHistoryStore.loadCallCount, 0)
    }

    func test_load_nutritionError_doesNotCallHealthKit() async throws {
        mockNutritionStore.loadResult = .failure(NetworkError.requestFailed(statusCode: 500, message: ""))

        await sut.load()

        XCTAssertEqual(mockHealthKit.refreshDailyActivityCallCount, 0)
    }

    func test_load_nutritionAuthError_doesNotSetErrorState() async throws {
        mockNutritionStore.loadResult = .failure(NetworkError.missingToken)

        await sut.load()

        if case .error = sut.state {
            XCTFail("Auth errors must not set .error state")
        }
    }

    func test_hasTodayWorkout_falseByDefault() {
        XCTAssertFalse(sut.hasTodayWorkout)
    }

    func test_hasTodayWorkout_true_afterLoad_whenTodayCountGT0() async throws {
        mockHistoryStore.todayCompletedCount = 1

        await sut.load()

        XCTAssertTrue(sut.hasTodayWorkout)
    }

    func test_hasTodayWorkout_false_afterLoad_whenNoneToday() async throws {
        mockHistoryStore.todayCompletedCount = 0

        await sut.load()

        XCTAssertFalse(sut.hasTodayWorkout)
    }

    func test_hasTodayWorkout_reactivelyUpdates_whenWorkoutsPublisherEmits() async throws {
        await sut.load()
        XCTAssertFalse(sut.hasTodayWorkout)

        mockHistoryStore.todayCompletedCount = 2
        mockHistoryStore.emitWorkoutsUpdate()
        await drainTasks()

        XCTAssertTrue(sut.hasTodayWorkout)
    }

    func test_hasTodayWorkout_reactivelyClears_whenWorkoutsUpdated() async throws {
        mockHistoryStore.todayCompletedCount = 1
        await sut.load()
        XCTAssertTrue(sut.hasTodayWorkout)

        mockHistoryStore.todayCompletedCount = 0
        mockHistoryStore.emitWorkoutsUpdate()
        await drainTasks()

        XCTAssertFalse(sut.hasTodayWorkout)
    }

    func test_hasWeeklyGoalMet_falseByDefault() {
        XCTAssertFalse(sut.hasWeeklyGoalMet)
    }

    func test_hasWeeklyGoalMet_true_whenCompletedCountMeetsTarget() async throws {
        mockUserStore.setProfileValue(.stub(targetWorkoutsWeekly: 3))
        mockHistoryStore.thisWeekCompletedCount = 3

        await sut.load()

        XCTAssertTrue(sut.hasWeeklyGoalMet)
    }

    func test_hasWeeklyGoalMet_true_whenCompletedCountExceedsTarget() async throws {
        mockUserStore.setProfileValue(.stub(targetWorkoutsWeekly: 3))
        mockHistoryStore.thisWeekCompletedCount = 5

        await sut.load()

        XCTAssertTrue(sut.hasWeeklyGoalMet)
    }

    func test_hasWeeklyGoalMet_false_whenCompletedCountBelowTarget() async throws {
        mockUserStore.setProfileValue(.stub(targetWorkoutsWeekly: 4))
        mockHistoryStore.thisWeekCompletedCount = 2

        await sut.load()

        XCTAssertFalse(sut.hasWeeklyGoalMet)
    }

    func test_hasWeeklyGoalMet_false_whenTargetIsZero() async throws {
        mockUserStore.setProfileValue(.stub(targetWorkoutsWeekly: 0))
        mockHistoryStore.thisWeekCompletedCount = 5

        await sut.load()

        XCTAssertFalse(sut.hasWeeklyGoalMet)
    }

    func test_hasWeeklyGoalMet_false_whenProfileIsNil() async throws {
        mockHistoryStore.thisWeekCompletedCount = 10

        await sut.load()

        XCTAssertFalse(sut.hasWeeklyGoalMet)
    }

    func test_hasWeeklyGoalMet_exactlyAtTarget_isTrue() async throws {
        mockUserStore.setProfileValue(.stub(targetWorkoutsWeekly: 5))
        mockHistoryStore.thisWeekCompletedCount = 5

        await sut.load()

        XCTAssertTrue(sut.hasWeeklyGoalMet)
    }

    func test_hasWeeklyGoalMet_reactivelyUpdates_whenWorkoutsPublisherEmits() async throws {
        mockUserStore.setProfileValue(.stub(targetWorkoutsWeekly: 3))
        mockHistoryStore.thisWeekCompletedCount = 2
        await sut.load()
        XCTAssertFalse(sut.hasWeeklyGoalMet)

        mockHistoryStore.thisWeekCompletedCount = 3
        mockHistoryStore.emitWorkoutsUpdate()
        await drainTasks()

        XCTAssertTrue(sut.hasWeeklyGoalMet)
    }

    func test_meals_updatesWhenMealPreviewsPublisherEmits() async throws {
        let previews = [MealPreview(title: "Завтрак", calories: 450),
                        MealPreview(title: "Обед", calories: 700)]

        mockNutritionStore.setMealPreviews(previews)
        await drainTasks()

        XCTAssertEqual(sut.meals.count, 2)
        XCTAssertEqual(sut.meals.first?.title, "Завтрак")
    }

    func test_goals_updatesWhenTargetCaloriesPublisherEmits() async throws {
        mockUserStore.setTargetCaloriesValue(2200)
        await drainTasks()

        XCTAssertEqual(sut.goals?.calories, 2200)
    }

    func test_stats_updatesCaloriesConsumed_whenDailySummaryChanges() async throws {
        await sut.load()

        let newSummary = NutritionDailySummary.stub(
            consumed: .stub(protein: 60, fat: 25, carbs: 120)
        )
        mockNutritionStore.setSummary(newSummary)
        await drainTasks()

        XCTAssertEqual(sut.stats?.caloriesConsumed, newSummary.consumed.calories)
    }

    func test_stats_updatesBurned_whenCaloriesBurnedChanges() async throws {
        await sut.load()

        mockUserStore.setCaloriesBurned(550)
        await drainTasks()

        XCTAssertEqual(sut.stats?.caloriesBurned, 550)
    }

    func test_stats_updatesSteps_whenTodayStepsChanges() async throws {
        await sut.load()

        mockUserStore.setTodaySteps(12000)
        await drainTasks()

        XCTAssertEqual(sut.stats?.steps, 12000)
    }

    func test_stats_preservesExistingValues_whenOnlyOneFieldChanges() async throws {
        mockUserStore.setTodaySteps(9000)
        mockUserStore.setCaloriesBurned(300)
        mockNutritionStore.setSummary(.stub(consumed: .stub(protein: 40, fat: 15, carbs: 80)))
        await sut.load()

        mockUserStore.setTodaySteps(11000)
        await drainTasks()

        XCTAssertEqual(sut.stats?.steps, 11000)
        XCTAssertEqual(sut.stats?.caloriesBurned, 300)
    }

    func test_stats_doesNotUpdate_beforeLoadSetsStats() async throws {
        let newSummary = NutritionDailySummary.stub()
        mockNutritionStore.setSummary(newSummary)
        await drainTasks()

        XCTAssertNil(sut.stats)
    }
}
