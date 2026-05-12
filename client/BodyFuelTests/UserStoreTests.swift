import XCTest
@testable import BodyFuel

@MainActor
final class UserStoreTests: XCTestCase {

    var sut: UserStore!

    override func setUp() async throws {
        sut = UserStore.shared
        sut.reset()
    }

    override func tearDown() async throws {
        sut.reset()
    }

    // MARK: - setCaloriesBurned

    func test_setCaloriesBurned_updatesCaloriesBurnedProperty() {
        sut.setCaloriesBurned(350)
        XCTAssertEqual(sut.caloriesBurned, 350)
    }

    func test_setCaloriesBurned_truncatesDecimalToInt() {
        sut.setCaloriesBurned(299.9)
        XCTAssertEqual(sut.caloriesBurned, 299)
    }

    func test_setCaloriesBurned_zero_setsToZero() {
        sut.setCaloriesBurned(500)
        sut.setCaloriesBurned(0)
        XCTAssertEqual(sut.caloriesBurned, 0)
    }

    func test_setCaloriesBurned_largeValue() {
        sut.setCaloriesBurned(9999)
        XCTAssertEqual(sut.caloriesBurned, 9999)
    }

    func test_setCaloriesBurned_overwritesPreviousValue() {
        sut.setCaloriesBurned(100)
        sut.setCaloriesBurned(250)
        XCTAssertEqual(sut.caloriesBurned, 250)
    }

    // MARK: - setTargetCalories

    func test_setTargetCalories_updatesProperty() {
        sut.setTargetCalories(2200)
        XCTAssertEqual(sut.targetCalories, 2200)
    }

    func test_setTargetCalories_zero_setsToZero() {
        sut.setTargetCalories(2000)
        sut.setTargetCalories(0)
        XCTAssertEqual(sut.targetCalories, 0)
    }

    func test_setTargetCalories_overwritesPreviousValue() {
        sut.setTargetCalories(1800)
        sut.setTargetCalories(2500)
        XCTAssertEqual(sut.targetCalories, 2500)
    }

    // MARK: - setBasalMetabolicRate

    func test_setBasalMetabolicRate_updatesProperty() {
        sut.setBasalMetabolicRate(1500)
        XCTAssertEqual(sut.basalMetabolicRate, 1500)
    }

    func test_setBasalMetabolicRate_overwritesPreviousValue() {
        sut.setBasalMetabolicRate(1400)
        sut.setBasalMetabolicRate(1700)
        XCTAssertEqual(sut.basalMetabolicRate, 1700)
    }

    // MARK: - setProfile

    func test_setProfile_updatesProfileProperty() {
        let profile = UserProfile.stub(goal: .loseWeight)
        sut.setProfile(profile)
        XCTAssertEqual(sut.profile?.goal, .loseWeight)
    }

    func test_setProfile_setsIsDataStaleToFalse() {
        sut.isDataStale = true
        sut.setProfile(.stub())
        XCTAssertFalse(sut.isDataStale)
    }

    func test_setProfile_preservesAllFields() {
        let profile = UserProfile.stub(
            height: 180,
            goal: .gainMuscle,
            currentWeight: 75,
            targetWeight: 80,
            targetCaloriesDaily: 3000,
            targetWorkoutsWeekly: 5
        )
        sut.setProfile(profile)
        XCTAssertEqual(sut.profile?.height, 180)
        XCTAssertEqual(sut.profile?.goal, .gainMuscle)
        XCTAssertEqual(sut.profile?.currentWeight, 75)
        XCTAssertEqual(sut.profile?.targetWeight, 80)
        XCTAssertEqual(sut.profile?.targetCaloriesDaily, 3000)
        XCTAssertEqual(sut.profile?.targetWorkoutsWeekly, 5)
    }

    func test_setProfile_overwritesPreviousProfile() {
        sut.setProfile(.stub(goal: .maintain))
        sut.setProfile(.stub(goal: .loseWeight))
        XCTAssertEqual(sut.profile?.goal, .loseWeight)
    }

    func test_setProfile_persistedToDisk_canBeLoaded() {
        let profile = UserProfile.stub(goal: .gainMuscle)
        sut.setProfile(profile)

        let loaded = DiskCache.shared.load(
            UserProfile.self,
            key: "user_profile_anon"
        )
        XCTAssertNotNil(loaded)
        XCTAssertEqual(loaded?.goal, .gainMuscle)
    }

    // MARK: - invalidateProfile

    func test_invalidateProfile_clearsProfileProperty() {
        sut.setProfile(.stub())
        sut.invalidateProfile()
        XCTAssertNil(sut.profile)
    }

    func test_invalidateProfile_removesDiskCache() {
        sut.setProfile(.stub())
        sut.invalidateProfile()

        let loaded = DiskCache.shared.load(UserProfile.self, key: "user_profile_anon")
        XCTAssertNil(loaded)
    }

    // MARK: - reset

    func test_reset_clearsProfile() {
        sut.setProfile(.stub())
        sut.reset()
        XCTAssertNil(sut.profile)
    }

    func test_reset_clearsTargetCalories() {
        sut.setTargetCalories(2000)
        sut.reset()
        XCTAssertEqual(sut.targetCalories, 0)
    }

    func test_reset_clearsCaloriesBurned() {
        sut.setCaloriesBurned(400)
        sut.reset()
        XCTAssertEqual(sut.caloriesBurned, 0)
    }

    func test_reset_clearsBasalMetabolicRate() {
        sut.setBasalMetabolicRate(1600)
        sut.reset()
        XCTAssertEqual(sut.basalMetabolicRate, 0)
    }

    func test_reset_setsIsDataStaleToFalse() {
        sut.isDataStale = true
        sut.reset()
        XCTAssertFalse(sut.isDataStale)
    }

    func test_reset_clearsDiskCachedProfile() {
        sut.setProfile(.stub())
        sut.reset()

        let loaded = DiskCache.shared.load(UserProfile.self, key: "user_profile_anon")
        XCTAssertNil(loaded)
    }

    func test_reset_clearsTodaySteps() {
        sut.todaySteps = 8000
        sut.reset()
        XCTAssertEqual(sut.todaySteps, 0)
    }
}
