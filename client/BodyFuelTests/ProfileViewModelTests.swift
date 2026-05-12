import XCTest
import Combine
import HealthKit
@testable import BodyFuel

@MainActor
final class ProfileViewModelTests: XCTestCase {

    var mockService: MockProfileService!
    var mockHealthKit: MockHealthKitService!
    var mockPhotoService: MockPhotoService!
    var mockUserStore: MockUserStore!
    var sut: ProfileViewModel!

    override func setUp() async throws {
        mockService = MockProfileService()
        mockHealthKit = MockHealthKitService()
        mockPhotoService = MockPhotoService()
        mockUserStore = MockUserStore()
        makeSUT()
        setValidFields()
    }

    override func tearDown() async throws {
        sut = nil
        mockService = nil
        mockHealthKit = nil
        mockPhotoService = nil
        mockUserStore = nil
    }

    private func makeSUT() {
        sut = ProfileViewModel(
            service: mockService,
            healthService: mockHealthKit,
            photoService: mockPhotoService,
            userStore: mockUserStore
        )
    }

    private func setValidFields() {
        sut.weight = 70.0
        sut.targetWeight = 70.0
        sut.height = 170
        sut.targetCaloriesDaily = 2000
        sut.targetWorkoutsWeekly = 3
        sut.lifestyle = .active
        sut.goal = .maintain
        sut.fitnessLevel = .intermediate
    }

    private func drainTasks() async {
        await Task.yield()
        await Task.yield()
        await Task.yield()
    }

    func test_load_callsUserStoreLoad() async throws {
        await sut.load()

        XCTAssertEqual(mockUserStore.loadCallCount, 1)
    }

    func test_load_whenStoreHasProfile_populatesFields() async throws {
        let profile = UserProfile.stub(height: 180, goal: .maintain, currentWeight: 85, targetWeight: 85)
        mockUserStore.setProfileValue(profile)

        await sut.load()

        XCTAssertEqual(sut.height, 180)
        XCTAssertEqual(sut.weight, 85.0, accuracy: 0.001)
        XCTAssertEqual(sut.lifestyle, profile.lifestyle)
    }

    func test_load_setsScreenStateToIdle_whenProfileAvailable() async throws {
        mockUserStore.setProfileValue(.stub())

        await sut.load()

        XCTAssertEqual(sut.screenState, .idle)
    }

    func test_load_setsScreenStateToIdle_whenNoProfile() async throws {
        await sut.load()

        XCTAssertEqual(sut.screenState, .idle)
    }

    func test_load_profilePublisher_updatesFieldsOnChange() async throws {
        await sut.load()

        let updatedProfile = UserProfile.stub(height: 190, goal: .maintain, currentWeight: 90, targetWeight: 90)
        mockUserStore.setProfileValue(updatedProfile)
        await drainTasks()

        XCTAssertEqual(sut.height, 190)
        XCTAssertEqual(sut.weight, 90.0, accuracy: 0.001)
    }

    func test_save_success_setsIsEditingToFalse() async throws {
        sut.isEditing = true

        await sut.save()

        XCTAssertFalse(sut.isEditing)
    }

    func test_save_success_setsScreenStateToIdle() async throws {
        await sut.save()

        XCTAssertEqual(sut.screenState, .idle)
    }

    func test_save_success_callsUpdateProfile() async throws {
        await sut.save()

        XCTAssertEqual(mockService.updateProfileCallCount, 1)
    }

    func test_save_success_savedProfileMatchesFields() async throws {
        sut.height = 175
        sut.weight = 72.0
        sut.targetWeight = 72.0

        await sut.save()

        XCTAssertEqual(mockService.lastUpdatedProfile?.height, 175)
        XCTAssertEqual(mockService.lastUpdatedProfile?.currentWeight ?? 0, 72.0, accuracy: 0.001)
    }

    func test_save_success_callsUserStoreSetters() async throws {
        await sut.save()

        XCTAssertEqual(mockUserStore.setTargetCaloriesCallCount, 1)
        XCTAssertEqual(mockUserStore.setBasalMetabolicRateCallCount, 1)
        XCTAssertEqual(mockUserStore.setProfileCallCount, 1)
        XCTAssertEqual(mockUserStore.lastSetTargetCalories, sut.targetCaloriesDaily)
    }

    func test_save_withAvatarData_uploadsPhotoFirst() async throws {
        let avatarData = Data("fake-image".utf8)
        sut.avatarData = avatarData
        mockPhotoService.uploadResult = .success("https://cdn.example.com/avatar.jpg")

        await sut.save()

        XCTAssertEqual(mockPhotoService.uploadCallCount, 1)
        XCTAssertEqual(mockPhotoService.lastUploadedData, avatarData)
        XCTAssertEqual(sut.avatarUrl, "https://cdn.example.com/avatar.jpg")
    }

    func test_save_withAvatarData_clearsAvatarDataAfterSuccess() async throws {
        sut.avatarData = Data("fake-image".utf8)
        mockPhotoService.uploadResult = .success("https://cdn.example.com/avatar.jpg")

        await sut.save()

        XCTAssertNil(sut.avatarData)
    }

    func test_save_withoutAvatarData_doesNotCallPhotoService() async throws {
        sut.avatarData = nil

        await sut.save()

        XCTAssertEqual(mockPhotoService.uploadCallCount, 0)
    }

    func test_save_uploadAvatarError_setsErrorState_keepingAvatarData() async throws {
        let avatarData = Data("fake-image".utf8)
        sut.avatarData = avatarData
        mockPhotoService.uploadResult = .failure(NetworkError.requestFailed(statusCode: 500, message: ""))

        await sut.save()

        if case .error = sut.screenState { } else { XCTFail("Expected .error state") }
        XCTAssertEqual(sut.avatarData, avatarData, "avatarData must not be cleared on error")
        XCTAssertEqual(mockService.updateProfileCallCount, 0)
    }

    func test_save_updateProfileError_setsErrorState() async throws {
        mockService.updateProfileResult = .failure(NetworkError.requestFailed(statusCode: 422, message: "Bad data"))

        await sut.save()

        if case .error(let msg) = sut.screenState {
            XCTAssertFalse(msg.isEmpty)
        } else {
            XCTFail("Expected .error state")
        }
    }

    func test_save_updateProfileError_doesNotClearAvatarData() async throws {
        let data = Data("img".utf8)
        sut.avatarData = data
        mockPhotoService.uploadResult = .success("https://example.com/avatar.jpg")
        mockService.updateProfileResult = .failure(NetworkError.requestFailed(statusCode: 500, message: ""))

        await sut.save()

        XCTAssertEqual(sut.avatarData, data)
    }

    func test_save_updateProfileError_isEditingStaysTrue() async throws {
        sut.isEditing = true
        mockService.updateProfileResult = .failure(NetworkError.requestFailed(statusCode: 500, message: ""))

        await sut.save()

        XCTAssertTrue(sut.isEditing)
    }

    func test_save_emptyHeight_setsErrorState() async throws {
        sut.height = 0

        await sut.save()

        if case .error = sut.screenState { } else { XCTFail("Expected .error state") }
        XCTAssertEqual(mockService.updateProfileCallCount, 0)
    }

    func test_save_emptyWeight_setsErrorState() async throws {
        sut.weight = 0.0

        await sut.save()

        if case .error = sut.screenState { } else { XCTFail("Expected .error state") }
        XCTAssertEqual(mockService.updateProfileCallCount, 0)
    }

    func test_save_emptyTargetWeight_setsErrorState() async throws {
        sut.targetWeight = 0.0

        await sut.save()

        if case .error = sut.screenState { } else { XCTFail("Expected .error state") }
        XCTAssertEqual(mockService.updateProfileCallCount, 0)
    }

    func test_save_targetWeightAboveCurrentForLoseWeight_setsErrorState() async throws {
        sut.weight = 70.0
        sut.goal = .loseWeight
        sut.targetWeight = 80.0

        await sut.save()

        if case .error = sut.screenState { } else { XCTFail("Expected .error state") }
        XCTAssertNotNil(sut.targetWeightError)
        XCTAssertEqual(mockService.updateProfileCallCount, 0)
    }

    func test_save_targetWeightBelowCurrentForGainMuscle_setsErrorState() async throws {
        sut.weight = 70.0
        sut.goal = .gainMuscle
        sut.targetWeight = 60.0

        await sut.save()

        if case .error = sut.screenState { } else { XCTFail("Expected .error state") }
        XCTAssertNotNil(sut.targetWeightError)
        XCTAssertEqual(mockService.updateProfileCallCount, 0)
    }

    func test_goal_didSet_maintain_setsTargetWeightToCurrentWeight() async throws {
        sut.weight = 75.0
        sut.targetWeight = 65.0
        sut.goal = .maintain

        XCTAssertEqual(sut.targetWeight, 75.0, accuracy: 0.001)
    }

    func test_goal_didSet_maintainWithEqualWeights_clearsTargetWeightError() async throws {
        sut.weight = 70.0
        sut.goal = .maintain
        XCTAssertNil(sut.targetWeightError)
    }

    func test_countSheetCalories_setsFormStateToCounting() async throws {
        sut.weight = 65.0
        sut.height = 170

        await sut.countSheetCalories()

        XCTAssertEqual(sut.caloriesFormState, .counting)
    }

    func test_countSheetCalories_harrisBenedict_female_28yo() async throws {
        sut.weight = 65.0
        sut.height = 170
        sut.lifestyle = .active

        await sut.countSheetCalories()

        let expectedBMR: Float = 1411.5
        let expectedExpenditure: Float = expectedBMR * Lifestyle.active.physicalActivityLevel

        XCTAssertEqual(sut.sheetBasalMetabolicRate, expectedBMR, accuracy: 0.5)
        XCTAssertEqual(sut.sheetDailyExpenditure, expectedExpenditure, accuracy: 0.5)
    }

    func test_countSheetCalories_harrisBenedict_male_28yo() async throws {
        mockHealthKit.fetchGenderResult = .success(.male)
        sut.weight = 80.0
        sut.height = 180
        sut.lifestyle = .active

        await sut.countSheetCalories()

        let expectedBMR: Float = 1790.0
        let expectedExpenditure: Float = expectedBMR * Lifestyle.active.physicalActivityLevel

        XCTAssertEqual(sut.sheetBasalMetabolicRate, expectedBMR, accuracy: 0.5)
        XCTAssertEqual(sut.sheetDailyExpenditure, expectedExpenditure, accuracy: 0.5)
    }

    func test_countSheetCalories_maintain_targetEqualsExpenditure() async throws {
        sut.weight = 65.0
        sut.height = 170
        sut.goal = .maintain

        await sut.countSheetCalories()

        XCTAssertEqual(sut.sheetTargetCalories, sut.sheetDailyExpenditure, accuracy: 0.001)
    }

    func test_countSheetCalories_loseWeight_reducesTargetBy10Percent() async throws {
        sut.weight = 70.0
        sut.targetWeight = 60.0
        sut.goal = .loseWeight
        sut.height = 170

        await sut.countSheetCalories()

        XCTAssertLessThan(sut.sheetTargetCalories, sut.sheetDailyExpenditure)
        XCTAssertEqual(sut.sheetTargetCalories / sut.sheetDailyExpenditure, 0.9, accuracy: 0.001)
    }

    func test_countSheetCalories_gainMuscle_increasesTargetBy10Percent() async throws {
        sut.weight = 65.0
        sut.targetWeight = 75.0
        sut.goal = .gainMuscle
        sut.height = 170

        await sut.countSheetCalories()

        XCTAssertGreaterThan(sut.sheetTargetCalories, sut.sheetDailyExpenditure)
        XCTAssertEqual(sut.sheetTargetCalories / sut.sheetDailyExpenditure, 1.1, accuracy: 0.001)
    }

    func test_countSheetCalories_missingHealthData_usesDefaultAge30_setsError() async throws {
        mockHealthKit.fetchGenderResult = .failure(URLError(.unknown))
        mockHealthKit.fetchDateOfBirthResult = .failure(URLError(.unknown))
        sut.weight = 70.0
        sut.height = 175
        sut.lifestyle = .sedentary

        await sut.countSheetCalories()

        XCTAssertNotNil(sut.sheetHealthIntegrationError)

        let expectedBMR: Float = 1648.75
        XCTAssertEqual(sut.sheetBasalMetabolicRate, expectedBMR, accuracy: 0.5)
    }

    func test_countSheetCalories_missingHealthData_stillCompletesCalculation() async throws {
        mockHealthKit.fetchGenderResult = .failure(URLError(.unknown))
        mockHealthKit.fetchDateOfBirthResult = .failure(URLError(.unknown))
        sut.weight = 70.0
        sut.height = 170

        await sut.countSheetCalories()

        XCTAssertGreaterThan(sut.sheetBasalMetabolicRate, 0)
        XCTAssertGreaterThan(sut.sheetDailyExpenditure, 0)
        XCTAssertEqual(sut.caloriesFormState, .counting)
    }

    func test_countSheetCalories_zeroHeight_doesNotCalculate() async throws {
        sut.height = 0
        sut.weight = 70.0

        await sut.countSheetCalories()

        XCTAssertEqual(sut.sheetBasalMetabolicRate, 0.0)
        XCTAssertNotEqual(sut.caloriesFormState, .counting)
    }

    func test_countSheetCalories_healthDataPresent_clearsHealthIntegrationError() async throws {
        sut.weight = 65.0
        sut.height = 170

        await sut.countSheetCalories()

        XCTAssertNil(sut.sheetHealthIntegrationError)
    }

    func test_applySheetCalories_updatesTargetCaloriesDaily() async throws {
        sut.sheetTargetCalories = 2100

        await sut.applySheetCalories()

        XCTAssertEqual(sut.targetCaloriesDaily, 2100)
    }

    func test_applySheetCalories_closesSheet() async throws {
        sut.showCaloriesSheet = true
        sut.sheetTargetCalories = 2000

        await sut.applySheetCalories()

        XCTAssertFalse(sut.showCaloriesSheet)
    }

    func test_applySheetCalories_setsFormStateToPreview() async throws {
        sut.caloriesFormState = .counting
        sut.sheetTargetCalories = 2000

        await sut.applySheetCalories()

        XCTAssertEqual(sut.caloriesFormState, .preview)
    }

    func test_applySheetCalories_callsUpdateProfile() async throws {
        sut.sheetTargetCalories = 1800

        await sut.applySheetCalories()

        XCTAssertEqual(mockService.updateProfileCallCount, 1)
        XCTAssertEqual(mockService.lastUpdatedProfile?.targetCaloriesDaily, 1800)
    }

    func test_applySheetCalories_callsUserStoreSetters() async throws {
        sut.sheetTargetCalories = 2200
        sut.sheetBasalMetabolicRate = 1600

        await sut.applySheetCalories()

        XCTAssertEqual(mockUserStore.setTargetCaloriesCallCount, 1)
        XCTAssertEqual(mockUserStore.lastSetTargetCalories, 2200)
        XCTAssertEqual(mockUserStore.setBasalMetabolicRateCallCount, 1)
        XCTAssertEqual(mockUserStore.lastSetBasalMetabolicRate, 1600)
        XCTAssertEqual(mockUserStore.setProfileCallCount, 1)
    }

    func test_applySheetCalories_updateProfileError_setsErrorState() async throws {
        sut.sheetTargetCalories = 2000
        mockService.updateProfileResult = .failure(NetworkError.requestFailed(statusCode: 500, message: ""))

        await sut.applySheetCalories()

        XCTAssertFalse(sut.showCaloriesSheet)
        if case .error = sut.screenState { } else { XCTFail("Expected .error state") }
    }

    func test_deleteProfile_success_setsLogoutSuccessEvent() async throws {
        await sut.deleteProfile()

        XCTAssertEqual(sut.event, .logoutSuccess)
        XCTAssertEqual(mockService.deleteProfileCallCount, 1)
    }

    func test_deleteProfile_error_setsErrorState() async throws {
        mockService.deleteProfileResult = .failure(NetworkError.requestFailed(statusCode: 500, message: ""))

        await sut.deleteProfile()

        if case .error(let msg) = sut.screenState {
            XCTAssertFalse(msg.isEmpty)
        } else {
            XCTFail("Expected .error state")
        }
        XCTAssertNotEqual(sut.event, .logoutSuccess)
    }

    func test_logout_setsLogoutSuccessEvent() async throws {
        sut.logout()

        XCTAssertEqual(sut.event, .logoutSuccess)
    }
}
