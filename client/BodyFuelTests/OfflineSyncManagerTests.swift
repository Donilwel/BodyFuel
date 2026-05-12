import XCTest
@testable import BodyFuel

@MainActor
final class OfflineSyncManagerTests: XCTestCase {

    var mockNutrition: MockNutritionService!
    var mockStats: MockStatsService!
    var mockWorkout: MockWorkoutService!
    var mockFeedback: MockFeedbackService!
    var sut: OfflineSyncManager!

    override func setUp() async throws {
        mockNutrition = MockNutritionService()
        mockStats = MockStatsService()
        mockWorkout = MockWorkoutService()
        mockFeedback = MockFeedbackService()
        MutationQueue.shared.clear()
        makeSUT()
    }

    override func tearDown() async throws {
        MutationQueue.shared.clear()
        sut = nil
    }

    private func makeSUT() {
        sut = OfflineSyncManager(
            nutritionService: mockNutrition,
            statsService: mockStats,
            workoutService: mockWorkout,
            feedbackService: mockFeedback
        )
    }

    // MARK: - flush — empty queue

    func test_flush_emptyQueue_doesNothing() async throws {
        await sut.flush()

        XCTAssertEqual(mockNutrition.saveMealCallCount, 0)
        XCTAssertEqual(mockStats.addWeightCallCount, 0)
        XCTAssertFalse(sut.isSyncing)
    }

    func test_flush_setsIsSyncingFalse_afterCompletion() async throws {
        MutationQueue.shared.enqueue(type: .addWeight, payload: AddWeightPayload(weight: 70.0))

        await sut.flush()

        XCTAssertFalse(sut.isSyncing)
    }

    // MARK: - flush — addMeal

    func test_flush_addMeal_callsSaveMeal() async throws {
        let meal = Meal.stub()
        MutationQueue.shared.enqueue(type: .addMeal, payload: AddMealPayload(meal: meal))

        await sut.flush()

        XCTAssertEqual(mockNutrition.saveMealCallCount, 1)
        XCTAssertEqual(mockNutrition.lastSavedMeal?.id, meal.id)
    }

    func test_flush_addMeal_success_removesMutationFromQueue() async throws {
        MutationQueue.shared.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))

        await sut.flush()

        XCTAssertTrue(MutationQueue.shared.mutations.isEmpty)
    }

    // MARK: - flush — deleteMeal

    func test_flush_deleteMeal_callsDeleteFoodEntry() async throws {
        let mealId = UUID().uuidString
        MutationQueue.shared.enqueue(type: .deleteMeal, payload: DeleteMealPayload(mealId: mealId))

        await sut.flush()

        XCTAssertEqual(mockNutrition.deleteFoodEntryCallCount, 1)
        XCTAssertEqual(mockNutrition.lastDeletedMealId, mealId)
    }

    func test_flush_deleteMeal_success_removesFromQueue() async throws {
        MutationQueue.shared.enqueue(type: .deleteMeal, payload: DeleteMealPayload(mealId: UUID().uuidString))

        await sut.flush()

        XCTAssertTrue(MutationQueue.shared.mutations.isEmpty)
    }

    // MARK: - flush — addWeight

    func test_flush_addWeight_callsStatsServiceAddWeight() async throws {
        MutationQueue.shared.enqueue(type: .addWeight, payload: AddWeightPayload(weight: 73.5))

        await sut.flush()

        XCTAssertEqual(mockStats.addWeightCallCount, 1)
        XCTAssertEqual(mockStats.lastAddedWeight, 73.5)
    }

    func test_flush_addWeight_success_removesFromQueue() async throws {
        MutationQueue.shared.enqueue(type: .addWeight, payload: AddWeightPayload(weight: 70.0))

        await sut.flush()

        XCTAssertTrue(MutationQueue.shared.mutations.isEmpty)
    }

    // MARK: - flush — markRecommendationRead

    func test_flush_markRecommendationRead_callsStatsService() async throws {
        MutationQueue.shared.enqueue(
            type: .markRecommendationRead,
            payload: MarkRecommendationReadPayload(id: "rec-42")
        )

        await sut.flush()

        XCTAssertEqual(mockStats.markRecommendationReadCallCount, 1)
        XCTAssertEqual(mockStats.lastMarkedReadId, "rec-42")
    }

    func test_flush_markRecommendationRead_success_removesFromQueue() async throws {
        MutationQueue.shared.enqueue(
            type: .markRecommendationRead,
            payload: MarkRecommendationReadPayload(id: "rec-1")
        )

        await sut.flush()

        XCTAssertTrue(MutationQueue.shared.mutations.isEmpty)
    }

    // MARK: - flush — updateWorkout

    func test_flush_updateWorkout_callsWorkoutService() async throws {
        let payload = UpdateWorkoutPayload(
            workoutID: "workout-99",
            status: WorkoutStatus.completed.rawValue,
            duration: 3600,
            totalCalories: 300,
            exercises: []
        )
        MutationQueue.shared.enqueue(type: .updateWorkout, payload: payload)

        await sut.flush()

        XCTAssertEqual(mockWorkout.updateWorkoutCallCount, 1)
        XCTAssertEqual(mockWorkout.lastUpdateWorkoutId, "workout-99")
    }

    func test_flush_updateWorkout_success_removesFromQueue() async throws {
        let payload = UpdateWorkoutPayload(
            workoutID: "workout-1",
            status: WorkoutStatus.completed.rawValue,
            duration: 1800,
            totalCalories: 150,
            exercises: []
        )
        MutationQueue.shared.enqueue(type: .updateWorkout, payload: payload)

        await sut.flush()

        XCTAssertTrue(MutationQueue.shared.mutations.isEmpty)
    }

    // MARK: - flush — sendFeedback

    func test_flush_sendFeedback_callsFeedbackService() async throws {
        MutationQueue.shared.enqueue(
            type: .sendFeedback,
            payload: SendFeedbackPayload(message: "Great app!", email: "user@example.com")
        )

        await sut.flush()

        XCTAssertEqual(mockFeedback.sendFeedbackCallCount, 1)
        XCTAssertEqual(mockFeedback.lastMessage, "Great app!")
        XCTAssertEqual(mockFeedback.lastEmail, "user@example.com")
    }

    func test_flush_sendFeedback_nilEmail_passesNilToService() async throws {
        MutationQueue.shared.enqueue(
            type: .sendFeedback,
            payload: SendFeedbackPayload(message: "Bug report", email: nil)
        )

        await sut.flush()

        XCTAssertNil(mockFeedback.lastEmail)
    }

    func test_flush_sendFeedback_success_removesFromQueue() async throws {
        MutationQueue.shared.enqueue(
            type: .sendFeedback,
            payload: SendFeedbackPayload(message: "Hello", email: nil)
        )

        await sut.flush()

        XCTAssertTrue(MutationQueue.shared.mutations.isEmpty)
    }

    // MARK: - flush — multiple mutations

    func test_flush_multipleMutations_allExecuted() async throws {
        MutationQueue.shared.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))
        MutationQueue.shared.enqueue(type: .addWeight, payload: AddWeightPayload(weight: 71.0))
        MutationQueue.shared.enqueue(
            type: .sendFeedback,
            payload: SendFeedbackPayload(message: "Test", email: nil)
        )

        await sut.flush()

        XCTAssertEqual(mockNutrition.saveMealCallCount, 1)
        XCTAssertEqual(mockStats.addWeightCallCount, 1)
        XCTAssertEqual(mockFeedback.sendFeedbackCallCount, 1)
        XCTAssertTrue(MutationQueue.shared.mutations.isEmpty)
    }

    // MARK: - flush — server error on one mutation → continues others

    func test_flush_serverError_incrementsRetry_andContinues() async throws {
        mockNutrition.saveMealResult = .failure(
            NetworkError.requestFailed(statusCode: 500, message: "Server error")
        )
        MutationQueue.shared.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))
        MutationQueue.shared.enqueue(type: .addWeight, payload: AddWeightPayload(weight: 70.0))

        await sut.flush()

        XCTAssertEqual(MutationQueue.shared.mutations.count, 1)
        XCTAssertEqual(MutationQueue.shared.mutations.first?.type, .addMeal)
        XCTAssertEqual(MutationQueue.shared.mutations.first?.retryCount, 1)
        XCTAssertEqual(mockStats.addWeightCallCount, 1)
    }

    func test_flush_serverError_doesNotRemoveFailedMutation() async throws {
        mockStats.addWeightResult = .failure(
            NetworkError.requestFailed(statusCode: 503, message: "Unavailable")
        )
        MutationQueue.shared.enqueue(type: .addWeight, payload: AddWeightPayload(weight: 70.0))

        await sut.flush()

        XCTAssertEqual(MutationQueue.shared.mutations.count, 1)
    }

    func test_flush_serverError_onFirstMutation_continuesWithSecond() async throws {
        mockNutrition.saveMealResult = .failure(
            NetworkError.requestFailed(statusCode: 500, message: "")
        )
        MutationQueue.shared.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))
        MutationQueue.shared.enqueue(
            type: .sendFeedback,
            payload: SendFeedbackPayload(message: "msg", email: nil)
        )

        await sut.flush()

        XCTAssertEqual(mockFeedback.sendFeedbackCallCount, 1)
    }

    // MARK: - flush — network error stops flush

    func test_flush_networkError_stopsFlush_doesNotContinueOthers() async throws {
        mockNutrition.saveMealResult = .failure(
            NetworkError.network(URLError(.notConnectedToInternet))
        )
        MutationQueue.shared.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))
        MutationQueue.shared.enqueue(type: .addWeight, payload: AddWeightPayload(weight: 70.0))

        await sut.flush()

        XCTAssertEqual(mockStats.addWeightCallCount, 0)
    }

    func test_flush_networkError_doesNotRemoveMutation() async throws {
        mockNutrition.saveMealResult = .failure(
            NetworkError.network(URLError(.networkConnectionLost))
        )
        MutationQueue.shared.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))

        await sut.flush()

        XCTAssertEqual(MutationQueue.shared.mutations.count, 1)
        XCTAssertEqual(MutationQueue.shared.mutations.first?.retryCount, 0)
    }

    // MARK: - flush — guard against concurrent flush

    func test_flush_whileAlreadySyncing_doesNotStartAgain() async throws {
        MutationQueue.shared.enqueue(type: .addWeight, payload: AddWeightPayload(weight: 70.0))

        await sut.flush()

        XCTAssertFalse(sut.isSyncing)
    }

    // MARK: - flush — maxRetries → mutation dropped

    func test_flush_maxRetries_mutationIsDropped() async throws {
        mockNutrition.saveMealResult = .failure(
            NetworkError.requestFailed(statusCode: 500, message: "")
        )
        MutationQueue.shared.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))

        for _ in 0..<QueuedMutation.maxRetries {
            await sut.flush()
        }

        XCTAssertTrue(MutationQueue.shared.mutations.isEmpty)
    }
}
