import XCTest
@testable import BodyFuel

@MainActor
final class MutationQueueTests: XCTestCase {

    var sut: MutationQueue!

    override func setUp() async throws {
        sut = MutationQueue.shared
        sut.clear()
    }

    override func tearDown() async throws {
        sut.clear()
    }

    // MARK: - enqueue

    func test_enqueue_addsMutationToArray() {
        sut.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))
        XCTAssertEqual(sut.mutations.count, 1)
    }

    func test_enqueue_setsCorrectType() {
        sut.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))
        XCTAssertEqual(sut.mutations.first?.type, .addMeal)
    }

    func test_enqueue_multipleTypes_allAdded() {
        sut.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))
        sut.enqueue(type: .addWeight, payload: AddWeightPayload(weight: 72.5))
        sut.enqueue(type: .markRecommendationRead, payload: MarkRecommendationReadPayload(id: "rec1"))
        XCTAssertEqual(sut.mutations.count, 3)
    }

    func test_enqueue_setsRetryCountToZero() {
        sut.enqueue(type: .addWeight, payload: AddWeightPayload(weight: 70.0))
        XCTAssertEqual(sut.mutations.first?.retryCount, 0)
    }

    func test_enqueue_assignsUniqueIDs() {
        sut.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))
        sut.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))
        let ids = sut.mutations.map(\.id)
        XCTAssertNotEqual(ids[0], ids[1])
    }

    func test_enqueue_persistsToDisk() {
        sut.enqueue(type: .addWeight, payload: AddWeightPayload(weight: 68.0))

        sut.reload()
        XCTAssertEqual(sut.mutations.count, 1)
        XCTAssertEqual(sut.mutations.first?.type, .addWeight)
    }

    func test_enqueue_preservesOrder() {
        sut.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))
        sut.enqueue(type: .addWeight, payload: AddWeightPayload(weight: 72.0))
        sut.enqueue(type: .deleteMeal, payload: DeleteMealPayload(mealId: UUID().uuidString))

        XCTAssertEqual(sut.mutations[0].type, .addMeal)
        XCTAssertEqual(sut.mutations[1].type, .addWeight)
        XCTAssertEqual(sut.mutations[2].type, .deleteMeal)
    }

    func test_pendingCount_matchesMutationsCount() {
        sut.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))
        sut.enqueue(type: .addWeight, payload: AddWeightPayload(weight: 72.0))
        XCTAssertEqual(sut.pendingCount, sut.mutations.count)
    }

    // MARK: - remove

    func test_remove_removesMutationById() {
        sut.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))
        let id = sut.mutations.first!.id

        sut.remove(id: id)

        XCTAssertTrue(sut.mutations.isEmpty)
    }

    func test_remove_leavesOtherMutations() {
        sut.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))
        sut.enqueue(type: .addWeight, payload: AddWeightPayload(weight: 72.0))
        let firstId = sut.mutations.first!.id

        sut.remove(id: firstId)

        XCTAssertEqual(sut.mutations.count, 1)
        XCTAssertEqual(sut.mutations.first?.type, .addWeight)
    }

    func test_remove_unknownId_noChange() {
        sut.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))

        sut.remove(id: UUID())

        XCTAssertEqual(sut.mutations.count, 1)
    }

    func test_remove_persistsAfterRemoval() {
        sut.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))
        sut.enqueue(type: .addWeight, payload: AddWeightPayload(weight: 72.0))
        let firstId = sut.mutations.first!.id

        sut.remove(id: firstId)
        sut.reload()

        XCTAssertEqual(sut.mutations.count, 1)
        XCTAssertEqual(sut.mutations.first?.type, .addWeight)
    }

    // MARK: - incrementRetry

    func test_incrementRetry_incrementsRetryCount() {
        sut.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))
        let id = sut.mutations.first!.id

        sut.incrementRetry(id: id)

        XCTAssertEqual(sut.mutations.first?.retryCount, 1)
    }

    func test_incrementRetry_twice_retryCountIsTwo() {
        sut.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))
        let id = sut.mutations.first!.id

        sut.incrementRetry(id: id)
        sut.incrementRetry(id: id)

        XCTAssertEqual(sut.mutations.first?.retryCount, 2)
    }

    func test_incrementRetry_atMaxRetries_removesMutation() {
        sut.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))
        let id = sut.mutations.first!.id

        for _ in 0..<QueuedMutation.maxRetries {
            sut.incrementRetry(id: id)
        }

        XCTAssertTrue(sut.mutations.isEmpty)
    }

    func test_incrementRetry_belowMaxRetries_doesNotRemove() {
        sut.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))
        let id = sut.mutations.first!.id

        for _ in 0..<(QueuedMutation.maxRetries - 1) {
            sut.incrementRetry(id: id)
        }

        XCTAssertEqual(sut.mutations.count, 1)
    }

    func test_incrementRetry_unknownId_noChange() {
        sut.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))

        sut.incrementRetry(id: UUID())

        XCTAssertEqual(sut.mutations.count, 1)
        XCTAssertEqual(sut.mutations.first?.retryCount, 0)
    }

    // MARK: - clear

    func test_clear_emptyMutations() {
        sut.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))
        sut.enqueue(type: .addWeight, payload: AddWeightPayload(weight: 72.0))

        sut.clear()

        XCTAssertTrue(sut.mutations.isEmpty)
    }

    func test_clear_persistsEmptyState() {
        sut.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))

        sut.clear()
        sut.reload()

        XCTAssertTrue(sut.mutations.isEmpty)
    }

    func test_clear_alreadyEmpty_noError() {
        sut.clear()
        XCTAssertTrue(sut.mutations.isEmpty)
    }

    // MARK: - reload

    func test_reload_loadsPersistedMutations() {
        sut.enqueue(type: .addWeight, payload: AddWeightPayload(weight: 75.0))
        sut.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))

        sut.reload()

        XCTAssertEqual(sut.mutations.count, 2)
    }

    func test_reload_afterClear_isEmpty() {
        sut.enqueue(type: .addMeal, payload: AddMealPayload(meal: .stub()))
        sut.clear()

        sut.reload()

        XCTAssertTrue(sut.mutations.isEmpty)
    }
}
