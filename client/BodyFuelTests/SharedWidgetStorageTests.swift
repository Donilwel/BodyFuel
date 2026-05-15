import XCTest
@testable import BodyFuel

final class SharedWidgetStorageTests: XCTestCase {

    var sut: SharedWidgetStorage!
    var testDefaults: UserDefaults!
    private let suiteName = "com.bodyfuel.test.widget"

    override func setUp() {
        testDefaults = UserDefaults(suiteName: suiteName)!
        testDefaults.removePersistentDomain(forName: suiteName)
        sut = SharedWidgetStorage(defaults: testDefaults)
    }

    override func tearDown() {
        testDefaults.removePersistentDomain(forName: suiteName)
        testDefaults = nil
        sut = nil
    }

    // MARK: - Workout roundtrip

    func test_saveWorkout_getWorkout_roundtrip() {
        let workout = WorkoutModel(name: "Утренняя зарядка", duration: 30, calories: 250, location: "home", type: "cardio")

        sut.saveWorkout(workout)
        let loaded = sut.getWorkout()

        XCTAssertEqual(loaded?.name, workout.name)
        XCTAssertEqual(loaded?.duration, workout.duration)
        XCTAssertEqual(loaded?.calories, workout.calories)
        XCTAssertEqual(loaded?.location, workout.location)
        XCTAssertEqual(loaded?.type, workout.type)
    }

    func test_saveWorkout_nil_getWorkoutReturnsNil() {
        let workout = WorkoutModel(name: "Test", duration: 10, calories: 100, location: "gym", type: "strength")
        sut.saveWorkout(workout)
        XCTAssertNotNil(sut.getWorkout())

        sut.saveWorkout(nil)

        XCTAssertNil(sut.getWorkout())
    }

    func test_getWorkout_beforeAnyWrite_returnsNil() {
        XCTAssertNil(sut.getWorkout())
    }

    func test_saveWorkout_overwritesPreviousWorkout() {
        sut.saveWorkout(WorkoutModel(name: "First", duration: 10, calories: 100, location: "home", type: "yoga"))
        sut.saveWorkout(WorkoutModel(name: "Second", duration: 45, calories: 400, location: "gym", type: "strength"))

        XCTAssertEqual(sut.getWorkout()?.name, "Second")
        XCTAssertEqual(sut.getWorkout()?.duration, 45)
    }

    // MARK: - isTodayWorkoutDone

    func test_saveTodayWorkoutDone_true_isTodayWorkoutDoneReturnsTrue() {
        sut.saveTodayWorkoutDone(true)
        XCTAssertTrue(sut.isTodayWorkoutDone())
    }

    func test_saveTodayWorkoutDone_false_isTodayWorkoutDoneReturnsFalse() {
        sut.saveTodayWorkoutDone(true)
        sut.saveTodayWorkoutDone(false)
        XCTAssertFalse(sut.isTodayWorkoutDone())
    }

    func test_isTodayWorkoutDone_beforeAnyWrite_returnsFalse() {
        XCTAssertFalse(sut.isTodayWorkoutDone())
    }

    // MARK: - Calories burned

    func test_saveTodayBurnedCalories_roundtrip() {
        sut.saveTodayBurnedCalories(350)
        XCTAssertEqual(sut.getTodayBurnedCalories(), 350)
    }

    func test_saveTodayBurnedCalories_zero_roundtrip() {
        sut.saveTodayBurnedCalories(0)
        XCTAssertEqual(sut.getTodayBurnedCalories(), 0)
    }

    func test_saveTodayBurnedCalories_overwritesPreviousValue() {
        sut.saveTodayBurnedCalories(200)
        sut.saveTodayBurnedCalories(500)
        XCTAssertEqual(sut.getTodayBurnedCalories(), 500)
    }

    func test_getTodayBurnedCalories_beforeAnyWrite_returnsNil() {
        XCTAssertNil(sut.getTodayBurnedCalories())
    }

    // MARK: - Calories consumed

    func test_saveTodayConsumedCalories_roundtrip() {
        sut.saveTodayConsumedCalories(1800)
        XCTAssertEqual(sut.getTodayConsumedCalories(), 1800)
    }

    func test_saveTodayConsumedCalories_overwritesPreviousValue() {
        sut.saveTodayConsumedCalories(1500)
        sut.saveTodayConsumedCalories(2100)
        XCTAssertEqual(sut.getTodayConsumedCalories(), 2100)
    }

    func test_getTodayConsumedCalories_beforeAnyWrite_returnsNil() {
        XCTAssertNil(sut.getTodayConsumedCalories())
    }

    // MARK: - Target calories

    func test_saveTargetCalories_roundtrip() {
        sut.saveTargetCalories(2500)
        XCTAssertEqual(sut.getTargetCalories(), 2500)
    }

    func test_getTargetCalories_zero_returnsNil() {
        sut.saveTargetCalories(0)
        XCTAssertNil(sut.getTargetCalories())
    }

    func test_getTargetCalories_beforeAnyWrite_returnsNil() {
        XCTAssertNil(sut.getTargetCalories())
    }

    func test_saveTargetCalories_overwritesPreviousValue() {
        sut.saveTargetCalories(2000)
        sut.saveTargetCalories(2800)
        XCTAssertEqual(sut.getTargetCalories(), 2800)
    }

    // MARK: - Basal metabolic rate

    func test_saveBasalMetabolicRate_roundtrip() {
        sut.saveBasalMetabolicRate(1600)
        XCTAssertEqual(sut.getBasalMetabolicRate(), 1600)
    }

    func test_getBasalMetabolicRate_zero_returnsNil() {
        sut.saveBasalMetabolicRate(0)
        XCTAssertNil(sut.getBasalMetabolicRate())
    }

    func test_getBasalMetabolicRate_beforeAnyWrite_returnsNil() {
        XCTAssertNil(sut.getBasalMetabolicRate())
    }

    // MARK: - Today steps

    func test_saveTodaySteps_roundtrip() {
        sut.saveTodaySteps(8500)
        XCTAssertEqual(sut.getTodaySteps(), 8500)
    }

    func test_saveTodaySteps_zero_roundtrip() {
        sut.saveTodaySteps(0)
        XCTAssertEqual(sut.getTodaySteps(), 0)
    }

    func test_getTodaySteps_beforeAnyWrite_returnsNil() {
        XCTAssertNil(sut.getTodaySteps())
    }

    func test_saveTodaySteps_overwritesPreviousValue() {
        sut.saveTodaySteps(5000)
        sut.saveTodaySteps(12000)
        XCTAssertEqual(sut.getTodaySteps(), 12000)
    }

    // MARK: - clearAll

    func test_clearAll_clearsCaloriesBurned() {
        sut.saveTodayBurnedCalories(400)
        sut.clearAll()
        XCTAssertNil(sut.getTodayBurnedCalories())
    }

    func test_clearAll_clearsCaloriesConsumed() {
        sut.saveTodayConsumedCalories(2000)
        sut.clearAll()
        XCTAssertNil(sut.getTodayConsumedCalories())
    }

    func test_clearAll_clearsTargetCalories() {
        sut.saveTargetCalories(2500)
        sut.clearAll()
        XCTAssertNil(sut.getTargetCalories())
    }

    func test_clearAll_clearsBasalMetabolicRate() {
        sut.saveBasalMetabolicRate(1700)
        sut.clearAll()
        XCTAssertNil(sut.getBasalMetabolicRate())
    }

    func test_clearAll_clearsSteps() {
        sut.saveTodaySteps(9000)
        sut.clearAll()
        XCTAssertNil(sut.getTodaySteps())
    }

    func test_clearAll_clearsWorkout() {
        sut.saveWorkout(WorkoutModel(name: "X", duration: 1, calories: 1, location: "home", type: "yoga"))
        sut.clearAll()
        XCTAssertNil(sut.getWorkout())
    }

    func test_clearAll_clearsTodayWorkoutDone() {
        sut.saveTodayWorkoutDone(true)
        sut.clearAll()
        XCTAssertFalse(sut.isTodayWorkoutDone())
    }

    func test_clearAll_clearsAllFieldsSimultaneously() {
        sut.saveTodayBurnedCalories(300)
        sut.saveTodayConsumedCalories(1800)
        sut.saveTargetCalories(2200)
        sut.saveBasalMetabolicRate(1500)
        sut.saveTodaySteps(7000)
        sut.saveWorkout(WorkoutModel(name: "Y", duration: 20, calories: 150, location: "gym", type: "strength"))
        sut.saveTodayWorkoutDone(true)

        sut.clearAll()

        XCTAssertNil(sut.getTodayBurnedCalories())
        XCTAssertNil(sut.getTodayConsumedCalories())
        XCTAssertNil(sut.getTargetCalories())
        XCTAssertNil(sut.getBasalMetabolicRate())
        XCTAssertNil(sut.getTodaySteps())
        XCTAssertNil(sut.getWorkout())
        XCTAssertFalse(sut.isTodayWorkoutDone())
    }

    func test_clearAll_emptyStorage_noError() {
        sut.clearAll()
    }

    func test_clearAll_afterClear_canSaveAndReadAgain() {
        sut.saveTodayBurnedCalories(200)
        sut.clearAll()
        sut.saveTodayBurnedCalories(999)
        XCTAssertEqual(sut.getTodayBurnedCalories(), 999)
    }
}
