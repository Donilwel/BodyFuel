import XCTest
@testable import BodyFuel

@MainActor
final class WorkoutViewModelTests: XCTestCase {

    var mockWorkoutService: MockWorkoutService!
    var mockHealthKitService: MockHealthKitService!
    var mockWidgetStorage: MockSharedWidgetStorage!
    var mockLiveActivity: MockLiveActivityService!

    var sut: WorkoutViewModel!

    override func setUp() async throws {
        try await super.setUp()
        mockWorkoutService = MockWorkoutService()
        mockHealthKitService = MockHealthKitService()
        mockWidgetStorage = MockSharedWidgetStorage()
        mockLiveActivity = MockLiveActivityService()
        makeSUT()
    }

    override func tearDown() async throws {
        sut = nil
        DiskCache.shared.remove(key: "workout_active_anon")
        DiskCache.shared.remove(key: "workout_active_testuser")
        DiskCache.shared.remove(key: "workout_active_testuser2")
        try await super.tearDown()
    }

    private func makeSUT() {
        sut = WorkoutViewModel(
            workoutService: mockWorkoutService,
            healthKitService: mockHealthKitService,
            sharedWidgetStorage: mockWidgetStorage,
            liveActivityService: mockLiveActivity
        )
    }

    private func setupWorkoutInProgress(exercise: Exercise = .stub(setCount: 3)) {
        sut.exercises = [exercise]
        sut.currentExerciseIndex = 0
        sut.currentSet = 1
        sut.phase = .waitingForStart
        sut.isWorkoutActive = true
    }

    private func drainTasks() async {
        await Task.yield()
        await Task.yield()
        await Task.yield()
    }
    
    // MARK: - load()

    func test_load_success_setsRecommendedWorkout() async {
        let workout = Workout.stub(title: "Test Workout")
        mockWorkoutService.generateWorkoutResult = .success(("w1", workout))

        await sut.load()

        XCTAssertEqual(sut.recommendedWorkout?.title, "Test Workout")
        XCTAssertFalse(sut.isWorkoutStale)
        XCTAssertEqual(sut.screenState, .loaded)
    }

    func test_load_success_savesWidgetModel() async {
        let workout = Workout.stub(title: "Morning Workout")
        mockWorkoutService.generateWorkoutResult = .success(("w1", workout))

        await sut.load()

        XCTAssertEqual(mockWidgetStorage.saveWorkoutCallCount, 1)
        XCTAssertEqual(mockWidgetStorage.lastSavedWorkoutName, "Morning Workout")
        XCTAssertFalse(mockWidgetStorage.savedWorkoutWasNil)
    }

    func test_load_success_resetsTodayWorkoutDone() async {
        mockWorkoutService.generateWorkoutResult = .success(("w1", .stub()))

        await sut.load()

        XCTAssertEqual(mockWidgetStorage.saveTodayWorkoutDoneCallCount, 1)
        XCTAssertEqual(mockWidgetStorage.lastSavedWorkoutDoneValue, false)
    }

    func test_load_whenAlreadyLoaded_doesNotFetchAgain() async {
        sut.recommendedWorkout = Workout.stub()

        await sut.load()

        XCTAssertEqual(mockWorkoutService.generateWorkoutCallCount, 0)
        XCTAssertEqual(mockWorkoutService.fetchWorkoutHistoryCallCount, 0)
    }

    func test_load_transportError_withCache_setsIsWorkoutStale() async {
        let cachedWorkout = Workout.stub(title: "Cached Workout")
        DiskCache.shared.save(cachedWorkout, key: "workout_active_anon")

        mockWorkoutService.fetchWorkoutHistoryResult = .failure(URLError(.notConnectedToInternet))

        await sut.load()

        XCTAssertTrue(sut.isWorkoutStale)
        XCTAssertEqual(sut.recommendedWorkout?.title, "Cached Workout")
        XCTAssertEqual(sut.screenState, .loaded)
    }

    func test_load_transportError_noCache_setsErrorState() async {
        mockWorkoutService.fetchWorkoutHistoryResult = .failure(URLError(.notConnectedToInternet))

        await sut.load()

        guard case .error = sut.screenState else {
            return XCTFail("Expected .error state, got \(sut.screenState)")
        }
    }

    func test_load_withStoredWorkoutID_callsFetchWorkout() async {
        UserSessionManager.shared.login(
            userId: "testuser", accessToken: "tok", refreshToken: "rtok"
        )
        UserDefaults.standard.set("stored-id", forKey: "pending_workout_id_testuser")
        defer {
            UserDefaults.standard.removeObject(forKey: "pending_workout_id_testuser")
            UserSessionManager.shared.logout(userId: "testuser")
        }

        mockWorkoutService.fetchWorkoutResult = .success(("stored-id", .stub()))
        makeSUT()

        await sut.load()

        XCTAssertEqual(mockWorkoutService.fetchWorkoutCallCount, 1)
        XCTAssertEqual(mockWorkoutService.lastFetchWorkoutId, "stored-id")
        XCTAssertEqual(mockWorkoutService.generateWorkoutCallCount, 0)
    }

    func test_load_storedWorkoutID_nonTransportError_resetsAndGenerates() async {
        UserSessionManager.shared.login(
            userId: "testuser2", accessToken: "tok", refreshToken: "rtok"
        )
        UserDefaults.standard.set("bad-id", forKey: "pending_workout_id_testuser2")
        defer {
            UserDefaults.standard.removeObject(forKey: "pending_workout_id_testuser2")
            UserSessionManager.shared.logout(userId: "testuser2")
        }

        mockWorkoutService.fetchWorkoutResult = .failure(
            NetworkError.requestFailed(statusCode: 404, message: "Not found")
        )
        mockWorkoutService.generateWorkoutResult = .success(("new-id", .stub()))
        makeSUT()

        await sut.load()

        XCTAssertEqual(mockWorkoutService.generateWorkoutCallCount, 1)
        XCTAssertNotNil(sut.recommendedWorkout)
    }
    
    // MARK: - startWorkout()

    func test_startWorkout_setsIsWorkoutActiveTrue() {
        sut.recommendedWorkout = Workout.stub()
        mockHealthKitService.hasGrantedPermission = true

        sut.startWorkout()

        XCTAssertTrue(sut.isWorkoutActive)
    }

    func test_startWorkout_setsPhaseWaitingForStart() {
        sut.recommendedWorkout = Workout.stub()
        mockHealthKitService.hasGrantedPermission = true

        sut.startWorkout()

        XCTAssertEqual(sut.phase, .waitingForStart)
    }

    func test_startWorkout_setsIsPausedFalse() {
        sut.recommendedWorkout = Workout.stub()
        sut.isPaused = true
        mockHealthKitService.hasGrantedPermission = true

        sut.startWorkout()

        XCTAssertFalse(sut.isPaused)
    }

    func test_startWorkout_withoutPermission_requestsAuthorization() async {
        sut.recommendedWorkout = Workout.stub()
        mockHealthKitService.hasGrantedPermission = false

        sut.startWorkout()
        await drainTasks()

        XCTAssertEqual(mockHealthKitService.requestAuthorizationCallCount, 1)
    }

    func test_startWorkout_startsHealthKitWorkout() async {
        sut.recommendedWorkout = Workout.stub(type: .cardio)
        mockHealthKitService.hasGrantedPermission = true

        sut.startWorkout()
        await drainTasks()

        XCTAssertEqual(mockHealthKitService.startWorkoutCallCount, 1)
    }
    
    // MARK: - finishWorkout()

    func test_finishWorkout_completed_setsShowWorkoutSummaryTrue() {
        setupWorkoutInProgress()
        sut.exerciseStats = [ExerciseStats(exercise: .stub(), repCount: ["10", "8", "6"])]
        sut.currentSet = 2

        sut.skipWorkout()

        XCTAssertTrue(sut.showWorkoutSummary)
        XCTAssertFalse(sut.isWorkoutActive)
    }

    func test_finishWorkout_savesNilToWidget_synchronouslyBeforeAPICall() {
        setupWorkoutInProgress()
        sut.exerciseStats = [ExerciseStats(exercise: .stub(), repCount: ["10"])]
        sut.currentSet = 2

        sut.skipWorkout()

        XCTAssertTrue(mockWidgetStorage.savedWorkoutWasNil)
        XCTAssertEqual(mockWorkoutService.updateWorkoutCallCount, 0)
    }

    func test_finishWorkout_completed_savesTodayWorkoutDoneTrue() {
        setupWorkoutInProgress()
        sut.exerciseStats = [ExerciseStats(exercise: .stub(), repCount: ["10"])]
        sut.currentSet = 2

        sut.skipWorkout()

        XCTAssertEqual(mockWidgetStorage.lastSavedWorkoutDoneValue, true)
    }

    func test_finishWorkout_failed_doesNotSaveTodayWorkoutDone() {
        setupWorkoutInProgress()
        sut.exerciseStats = []
        sut.currentSetRepCount = []
        sut.currentSet = 2

        sut.skipWorkout()

        XCTAssertTrue(sut.showWorkoutSummary)
        XCTAssertNil(mockWidgetStorage.lastSavedWorkoutDoneValue)
    }

    func test_finishWorkout_completed_callsUpdateWorkoutWithCompletedStatus() async {
        setupWorkoutInProgress()
        sut.currentWorkoutID = "workout-123"
        sut.exerciseStats = [ExerciseStats(exercise: .stub(), repCount: ["10"])]
        sut.currentSet = 2

        sut.skipWorkout()
        await drainTasks()

        XCTAssertEqual(mockWorkoutService.updateWorkoutCallCount, 1)
        XCTAssertEqual(mockWorkoutService.lastUpdateStatus, .completed)
        XCTAssertEqual(mockWorkoutService.lastUpdateWorkoutId, "workout-123")
    }

    func test_finishWorkout_failed_callsUpdateWorkoutWithFailedStatus() async {
        setupWorkoutInProgress()
        sut.currentWorkoutID = "workout-abc"
        sut.exerciseStats = []
        sut.currentSetRepCount = []
        sut.currentSet = 2

        sut.skipWorkout()
        await drainTasks()

        XCTAssertEqual(mockWorkoutService.updateWorkoutCallCount, 1)
        XCTAssertEqual(mockWorkoutService.lastUpdateStatus, .failed)
    }
    
    // MARK: - togglePause()

    func test_togglePause_setsPausedTrue() {
        sut.isPaused = false

        sut.togglePause()

        XCTAssertTrue(sut.isPaused)
    }

    func test_togglePause_callsHealthKitPause() {
        sut.isPaused = false

        sut.togglePause()

        XCTAssertEqual(mockHealthKitService.pauseWorkoutCallCount, 1)
    }

    func test_togglePause_secondCall_setsPausedFalse() {
        sut.isPaused = true

        sut.togglePause()

        XCTAssertFalse(sut.isPaused)
    }

    func test_togglePause_secondCall_callsHealthKitResume() {
        sut.isPaused = true

        sut.togglePause()

        XCTAssertEqual(mockHealthKitService.resumeWorkoutCallCount, 1)
    }

    func test_togglePause_inWaitingForStart_resumePreservesPhase() {
        sut.isPaused = true
        sut.phase = .waitingForStart
        sut.timeRemaining = 30

        sut.togglePause()

        XCTAssertFalse(sut.isPaused)
        XCTAssertEqual(sut.phase, .waitingForStart)
        XCTAssertEqual(sut.timeRemaining, 30)
        XCTAssertEqual(mockHealthKitService.resumeWorkoutCallCount, 1)
    }

    func test_togglePause_inExercise_resumePreservesPhase() {
        sut.isPaused = true
        sut.phase = .exercise
        sut.exercises = [Exercise.stub()]
        sut.currentExerciseIndex = 0

        sut.togglePause()

        XCTAssertFalse(sut.isPaused)
        XCTAssertEqual(sut.phase, .exercise)
    }
    
    // MARK: - startExercise()

    func test_startExercise_changesPhaseToExercise() {
        sut.exercises = [Exercise.stub(duration: 60)]
        sut.currentExerciseIndex = 0
        sut.phase = .waitingForStart

        sut.startExercise()

        XCTAssertEqual(sut.phase, .exercise)
    }

    func test_startExercise_setsTimeRemaining() {
        let exercise = Exercise.stub(duration: 45)
        sut.exercises = [exercise]
        sut.currentExerciseIndex = 0

        sut.startExercise()

        XCTAssertEqual(sut.timeRemaining, 45)
        XCTAssertEqual(sut.elapsedTime, 0)
    }
    
    // MARK: - moveToNextPhase()

    func test_moveToNextPhase_fromRestBetweenSets_incrementsCurrentSet() {
        sut.exercises = [Exercise.stub(setCount: 3, rest: 60)]
        sut.currentExerciseIndex = 0
        sut.phase = .restBetweenSets
        sut.currentSet = 1

        sut.moveToNextPhase()

        XCTAssertEqual(sut.currentSet, 2)
        XCTAssertEqual(sut.phase, .waitingForStart)
    }

    func test_moveToNextPhase_fromRestBetweenExercises_advancesToNextExercise() {
        let exercises = [Exercise.stub(name: "Ex1"), Exercise.stub(name: "Ex2", type: .upperBody)]
        sut.exercises = exercises
        sut.currentExerciseIndex = 0
        sut.phase = .restBetweenExercises

        sut.moveToNextPhase()

        XCTAssertEqual(sut.currentExerciseIndex, 1)
        XCTAssertEqual(sut.phase, .waitingForStart)
    }

    func test_moveToNextPhase_fromRestBetweenExercises_onLastExercise_finishesWorkout() {
        sut.exercises = [Exercise.stub()]
        sut.currentExerciseIndex = 0
        sut.phase = .restBetweenExercises

        sut.moveToNextPhase()

        XCTAssertTrue(sut.showWorkoutSummary)
        XCTAssertFalse(sut.isWorkoutActive)
    }

    func test_moveToNextPhase_fromExercise_cardio_transitionsToRestBetweenSets() {
        let cardio = Exercise.cardioStub(duration: 60, setCount: 3, rest: 45)
        sut.exercises = [cardio]
        sut.currentExerciseIndex = 0
        sut.currentSet = 1
        sut.phase = .exercise
        sut.timeRemaining = 15

        sut.moveToNextPhase()

        XCTAssertEqual(sut.phase, .restBetweenSets)
        XCTAssertEqual(sut.currentSetRepCount.count, 1)
    }

    func test_moveToNextPhase_fromExercise_nonCardio_validRep_transitionsToRest() {
        let exercise = Exercise.stub(setCount: 3, rest: 60)
        sut.exercises = [exercise]
        sut.currentExerciseIndex = 0
        sut.currentSet = 1
        sut.phase = .exercise
        sut.currentExerciseRepCount = "12"

        sut.moveToNextPhase()

        XCTAssertEqual(sut.phase, .restBetweenSets)
    }

    func test_moveToNextPhase_fromExercise_nonCardio_emptyRep_showsError() {
        let exercise = Exercise.stub(setCount: 3)
        sut.exercises = [exercise]
        sut.currentExerciseIndex = 0
        sut.phase = .exercise
        sut.currentExerciseRepCount = ""

        sut.moveToNextPhase()

        XCTAssertFalse(sut.currentExerciseRepCountError.isEmpty)
        XCTAssertEqual(sut.phase, .exercise)
    }
    
    // MARK: - skipExercise()

    func test_skipExercise_withTwoCompletedSets_keepsCompletedAndPadsZero() {
        let exercise = Exercise.stub(setCount: 3)
        sut.exercises = [exercise]
        sut.currentExerciseIndex = 0
        sut.currentSetRepCount = ["10", "8"]

        sut.skipExercise()

        XCTAssertEqual(sut.exerciseStats.first?.repCount, ["10", "8", "0"])
    }

    func test_skipExercise_withNoCompletedSets_allZeros() {
        let exercise = Exercise.stub(setCount: 3)
        sut.exercises = [exercise]
        sut.currentExerciseIndex = 0
        sut.currentSetRepCount = []

        sut.skipExercise()

        XCTAssertEqual(sut.exerciseStats.first?.repCount, ["0", "0", "0"])
    }

    func test_skipExercise_resetsCurrentSetState() {
        let exercise = Exercise.stub(setCount: 2)
        sut.exercises = [exercise]
        sut.currentSetRepCount = ["5"]
        sut.currentSet = 2

        sut.skipExercise()

        XCTAssertEqual(sut.currentSetRepCount, [])
        XCTAssertEqual(sut.currentSet, 1)
    }

    func test_skipExercise_transitionsToRestBetweenExercises() {
        let exercise = Exercise.stub(setCount: 2)
        sut.exercises = [exercise, Exercise.stub(name: "Ex2", type: .upperBody)]
        sut.currentExerciseIndex = 0
        sut.currentSetRepCount = []

        sut.skipExercise()

        XCTAssertEqual(sut.phase, .restBetweenExercises)
    }
    
    // MARK: - skipWorkout()

    func test_skipWorkout_atStart_deactivatesWorkoutWithoutSummary() {
        setupWorkoutInProgress()

        sut.skipWorkout()

        XCTAssertFalse(sut.isWorkoutActive)
        XCTAssertNil(sut.recommendedWorkout)
        XCTAssertFalse(sut.showWorkoutSummary)
    }

    func test_skipWorkout_atStart_callsDiscardAndUpdateFailed() async {
        setupWorkoutInProgress()
        sut.currentWorkoutID = "workout-cancel"

        sut.skipWorkout()
        await drainTasks()

        XCTAssertEqual(mockHealthKitService.discardWorkoutCallCount, 1)
        XCTAssertEqual(mockWorkoutService.updateWorkoutCallCount, 1)
        XCTAssertEqual(mockWorkoutService.lastUpdateStatus, .failed)
    }

    func test_skipWorkout_atStart_doesNotClearWidget() {
        setupWorkoutInProgress()

        sut.skipWorkout()

        XCTAssertEqual(mockWidgetStorage.saveWorkoutCallCount, 0)
    }

    func test_skipWorkout_afterSomeReps_finishesAsCompleted() {
        setupWorkoutInProgress()
        sut.exerciseStats = [ExerciseStats(exercise: .stub(), repCount: ["10", "8"])]
        sut.currentSet = 2

        sut.skipWorkout()

        XCTAssertTrue(sut.showWorkoutSummary)
        XCTAssertEqual(mockWidgetStorage.lastSavedWorkoutDoneValue, true)
    }

    func test_skipWorkout_noRepsAnywhere_finishesAsFailed() {
        setupWorkoutInProgress()
        sut.exerciseStats = []
        sut.currentSetRepCount = []
        sut.currentSet = 2

        sut.skipWorkout()

        XCTAssertTrue(sut.showWorkoutSummary)
        XCTAssertNil(mockWidgetStorage.lastSavedWorkoutDoneValue)
        XCTAssertTrue(mockWidgetStorage.savedWorkoutWasNil)
    }
    
    // MARK: - workoutProgress

    func test_workoutProgress_noCompletedSets_isZero() {
        sut.exercises = [Exercise.stub(setCount: 3)]
        sut.currentExerciseIndex = 0
        sut.exerciseStats = []
        sut.currentSetRepCount = []

        XCTAssertEqual(sut.workoutProgress, 0)
    }

    func test_workoutProgress_afterFirstExerciseCompleted_isCorrect() {
        let ex1 = Exercise.stub(name: "Ex1", setCount: 2)
        let ex2 = Exercise.stub(name: "Ex2", setCount: 2)
        sut.exercises = [ex1, ex2]
        sut.currentExerciseIndex = 1
        sut.exerciseStats = [ExerciseStats(exercise: ex1, repCount: ["10", "8"])]
        sut.currentSetRepCount = []

        XCTAssertEqual(sut.workoutProgress, 0.5, accuracy: 0.001)
    }

    func test_workoutProgress_inProgressSetsCountedForCurrentExercise() {
        let exercise = Exercise.stub(name: "Ex1", setCount: 4)
        sut.exercises = [exercise]
        sut.currentExerciseIndex = 0
        sut.exerciseStats = []
        sut.currentSetRepCount = ["10", "8"]

        XCTAssertEqual(sut.workoutProgress, 0.5, accuracy: 0.001)
    }
    
    // MARK: - isLastSet

    func test_isLastSet_trueOnLastSetOfLastExercise() {
        let exercise = Exercise.stub(setCount: 3)
        sut.exercises = [exercise]
        sut.currentExerciseIndex = 0
        sut.currentSet = 3

        XCTAssertTrue(sut.isLastSet)
    }

    func test_isLastSet_falseWhenNotOnLastSet() {
        let exercise = Exercise.stub(setCount: 3)
        sut.exercises = [exercise]
        sut.currentExerciseIndex = 0
        sut.currentSet = 2

        XCTAssertFalse(sut.isLastSet)
    }

    func test_isLastSet_falseWhenNotOnLastExercise() {
        sut.exercises = [Exercise.stub(setCount: 3), Exercise.stub(name: "Ex2", setCount: 2)]
        sut.currentExerciseIndex = 0
        sut.currentSet = 3

        XCTAssertFalse(sut.isLastSet)
    }
    
    // MARK: - generateWithFilters()

    func test_generateWithFilters_savesWidgetModelAndResetsDone() async {
        mockWorkoutService.generateWorkoutResult = .success(("w2", .stub(title: "Filtered")))

        await sut.generateWithFilters(place: .home, type: .cardio, level: .light)

        XCTAssertEqual(mockWidgetStorage.lastSavedWorkoutName, "Filtered")
        XCTAssertEqual(mockWidgetStorage.lastSavedWorkoutDoneValue, false)
    }

    func test_generateWithFilters_updatesOldWorkoutAsFailed() async {
        sut.currentWorkoutID = "old-id"
        mockWorkoutService.generateWorkoutResult = .success(("new-id", .stub()))

        await sut.generateWithFilters(place: nil, type: nil, level: nil)

        XCTAssertEqual(mockWorkoutService.updateWorkoutCallCount, 1)
        XCTAssertEqual(mockWorkoutService.lastUpdateWorkoutId, "old-id")
        XCTAssertEqual(mockWorkoutService.lastUpdateStatus, .failed)
    }
    
    // MARK: - startWorkoutFromDeepLink()

    func test_startWorkoutFromDeepLink_fetchesWorkoutById() async {
        mockWorkoutService.fetchWorkoutResult = .success(("dl-id", .stub()))

        sut.startWorkoutFromDeepLink(id: "dl-id")
        await drainTasks()

        XCTAssertEqual(mockWorkoutService.fetchWorkoutCallCount, 1)
        XCTAssertEqual(mockWorkoutService.lastFetchWorkoutId, "dl-id")
    }

    func test_startWorkoutFromDeepLink_startsWorkout() async {
        mockWorkoutService.fetchWorkoutResult = .success(("dl-id", .stub()))

        sut.startWorkoutFromDeepLink(id: "dl-id")
        await drainTasks()

        XCTAssertTrue(sut.isWorkoutActive)
        XCTAssertEqual(sut.phase, .waitingForStart)
    }

    func test_startWorkoutFromDeepLink_onFetchError_resetsScreenState() async {
        mockWorkoutService.fetchWorkoutResult = .failure(
            NetworkError.requestFailed(statusCode: 500, message: "Server error")
        )

        sut.startWorkoutFromDeepLink(id: "bad-id")
        await drainTasks()

        XCTAssertFalse(sut.isWorkoutActive)
        XCTAssertEqual(sut.screenState, .loaded)
    }
}
