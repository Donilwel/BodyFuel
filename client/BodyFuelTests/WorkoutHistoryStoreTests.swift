import XCTest
@testable import BodyFuel
import Combine

@MainActor
final class WorkoutHistoryStoreTests: XCTestCase {

    var sut: WorkoutHistoryStore!

    override func setUp() async throws {
        sut = WorkoutHistoryStore.shared
        sut.workouts = []
    }

    override func tearDown() async throws {
        sut.workouts = []
    }

    private func makeItem(
        id: String = UUID().uuidString,
        status: String,
        date: Date
    ) -> WorkoutHistoryItem {
        WorkoutHistoryItem(
            id: id, title: "Тренировка", level: "workout_middle",
            status: status, totalCalories: 0, duration: 0,
            date: date, exercisesCount: 3, completedCount: 3, exercises: []
        )
    }

    private var today: Date { Date() }

    private var yesterday: Date {
        Calendar.current.date(byAdding: .day, value: -1, to: Date())!
    }

    private var lastWeek: Date {
        Calendar.current.date(byAdding: .weekOfYear, value: -1, to: Date())!
    }

    private var tomorrow: Date {
        Calendar.current.date(byAdding: .day, value: 1, to: Date())!
    }

    // MARK: - todayCompletedCount

    func test_todayCompletedCount_zeroWhenNoWorkouts() {
        XCTAssertEqual(sut.todayCompletedCount, 0)
    }

    func test_todayCompletedCount_countsWorkoutDoneToday() {
        sut.workouts = [makeItem(status: "workout_done", date: today)]
        XCTAssertEqual(sut.todayCompletedCount, 1)
    }

    func test_todayCompletedCount_countsMultipleDoneToday() {
        sut.workouts = [
            makeItem(status: "workout_done", date: today),
            makeItem(status: "workout_done", date: today)
        ]
        XCTAssertEqual(sut.todayCompletedCount, 2)
    }

    func test_todayCompletedCount_excludesWrongStatus() {
        sut.workouts = [
            makeItem(status: "in_progress", date: today),
            makeItem(status: "scheduled", date: today),
            makeItem(status: "workout_done", date: today)
        ]
        XCTAssertEqual(sut.todayCompletedCount, 1)
    }

    func test_todayCompletedCount_excludesYesterdayDone() {
        sut.workouts = [makeItem(status: "workout_done", date: yesterday)]
        XCTAssertEqual(sut.todayCompletedCount, 0)
    }

    func test_todayCompletedCount_excludesTomorrowDone() {
        sut.workouts = [makeItem(status: "workout_done", date: tomorrow)]
        XCTAssertEqual(sut.todayCompletedCount, 0)
    }

    func test_todayCompletedCount_excludesLastWeekDone() {
        sut.workouts = [makeItem(status: "workout_done", date: lastWeek)]
        XCTAssertEqual(sut.todayCompletedCount, 0)
    }

    func test_todayCompletedCount_mixedDatesAndStatuses() {
        sut.workouts = [
            makeItem(status: "workout_done", date: today),
            makeItem(status: "workout_done", date: yesterday),
            makeItem(status: "in_progress", date: today),
            makeItem(status: "workout_done", date: today)
        ]
        XCTAssertEqual(sut.todayCompletedCount, 2)
    }

    func test_todayCompletedCount_zeroWithOnlyNonDoneStatuses() {
        sut.workouts = [
            makeItem(status: "scheduled", date: today),
            makeItem(status: "in_progress", date: today)
        ]
        XCTAssertEqual(sut.todayCompletedCount, 0)
    }

    // MARK: - thisWeekCompletedCount

    func test_thisWeekCompletedCount_zeroWhenNoWorkouts() {
        XCTAssertEqual(sut.thisWeekCompletedCount, 0)
    }

    func test_thisWeekCompletedCount_countsTodayDone() {
        sut.workouts = [makeItem(status: "workout_done", date: today)]
        XCTAssertEqual(sut.thisWeekCompletedCount, 1)
    }

    func test_thisWeekCompletedCount_excludesLastWeek() {
        sut.workouts = [makeItem(status: "workout_done", date: lastWeek)]
        XCTAssertEqual(sut.thisWeekCompletedCount, 0)
    }

    func test_thisWeekCompletedCount_excludesWrongStatus() {
        sut.workouts = [
            makeItem(status: "in_progress", date: today),
            makeItem(status: "scheduled", date: today)
        ]
        XCTAssertEqual(sut.thisWeekCompletedCount, 0)
    }

    func test_thisWeekCompletedCount_countsMultipleThisWeek() {
        sut.workouts = [
            makeItem(status: "workout_done", date: today),
            makeItem(status: "workout_done", date: today)
        ]
        XCTAssertEqual(sut.thisWeekCompletedCount, 2)
    }

    func test_thisWeekCompletedCount_onlyThisWeek_excludesLastWeek() {
        sut.workouts = [
            makeItem(status: "workout_done", date: today),
            makeItem(status: "workout_done", date: today),
            makeItem(status: "workout_done", date: lastWeek)
        ]
        XCTAssertEqual(sut.thisWeekCompletedCount, 2)
    }

    func test_thisWeekCompletedCount_mixedStatusesThisWeek() {
        sut.workouts = [
            makeItem(status: "workout_done", date: today),
            makeItem(status: "in_progress", date: today),
            makeItem(status: "workout_done", date: lastWeek)
        ]
        XCTAssertEqual(sut.thisWeekCompletedCount, 1)
    }

    func test_thisWeekCompletedCount_yesterdayInSameWeek_isIncluded() {
        guard let weekInterval = Calendar.current.dateInterval(of: .weekOfYear, for: Date()),
              weekInterval.contains(yesterday) else {
            return
        }
        sut.workouts = [makeItem(status: "workout_done", date: yesterday)]
        XCTAssertEqual(sut.thisWeekCompletedCount, 1)
    }

    // MARK: - workoutsPublisher

    func test_workoutsPublisher_emitsOnWorkoutsChange() {
        var emitted = false
        let cancellable = sut.workoutsPublisher.dropFirst().sink { _ in emitted = true }

        sut.workouts = [makeItem(status: "workout_done", date: today)]

        XCTAssertTrue(emitted)
        cancellable.cancel()
    }

    func test_workoutsPublisher_emitsCurrentValueImmediately() {
        sut.workouts = [makeItem(status: "workout_done", date: today)]
        var receivedCount = 0
        let cancellable = sut.workoutsPublisher.sink { receivedCount = $0.count }
        XCTAssertEqual(receivedCount, 1)
        cancellable.cancel()
    }
}
