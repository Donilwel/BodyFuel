import XCTest
@testable import BodyFuel

@MainActor
final class StatsViewModelTests: XCTestCase {

    var mockStore: MockStatsStore!
    var mockHealth: MockHealthKitService!
    var mockUserStore: MockUserStore!
    var sut: StatsViewModel!

    override func setUp() async throws {
        mockStore = MockStatsStore()
        mockHealth = MockHealthKitService()
        mockUserStore = MockUserStore()
        makeSUT()
    }

    override func tearDown() async throws {
        sut = nil
        mockStore = nil
        mockHealth = nil
        mockUserStore = nil
    }

    private func makeSUT() {
        sut = StatsViewModel(
            store: mockStore,
            healthService: mockHealth,
            userStore: mockUserStore
        )
    }

    // MARK: - load()

    func test_load_callsHealthKitRefresh() async throws {
        await sut.load()

        XCTAssertEqual(mockHealth.refreshDailyActivityCallCount, 1)
    }

    func test_load_callsLoadWeightHistory() async throws {
        await sut.load()

        XCTAssertEqual(mockStore.loadWeightHistoryCallCount, 1)
    }

    func test_load_callsLoadRecommendations() async throws {
        await sut.load()

        XCTAssertEqual(mockStore.loadRecommendationsCallCount, 1)
    }

    func test_load_callsMarkAllRead() async throws {
        await sut.load()

        XCTAssertEqual(mockStore.markAllReadCallCount, 1)
    }

    func test_load_setsIsLoadingChartFalse_afterCompletion() async throws {
        await sut.load()

        XCTAssertFalse(sut.isLoadingChart)
    }

    func test_load_callsReloadChart_whichSetsIsLoadingChartFalse() async throws {
        await sut.load()

        XCTAssertFalse(sut.isLoadingChart)
    }

    func test_load_defaultMetricIsWeight_populatesChartPointsFromHistory() async throws {
        let now = Date()
        let iso = ISO8601DateFormatter()
        iso.formatOptions = [.withInternetDateTime]
        mockStore.weightHistory = [
            WeightEntryResponse(id: "1", weight: 70.0, date: iso.string(from: now))
        ]

        await sut.load()

        XCTAssertEqual(sut.chartPoints.count, 1)
        XCTAssertEqual(sut.chartPoints.first?.value, 70.0)
    }

    func test_load_weightMetric_weightPeriodChange_nilWhenLessThanTwoPoints() async throws {
        let iso = ISO8601DateFormatter()
        iso.formatOptions = [.withInternetDateTime]
        mockStore.weightHistory = [
            WeightEntryResponse(id: "1", weight: 70.0, date: iso.string(from: Date()))
        ]

        await sut.load()

        XCTAssertNil(sut.weightPeriodChange)
    }

    func test_load_weightMetric_weightPeriodChange_computedWhenTwoOrMorePoints() async throws {
        let iso = ISO8601DateFormatter()
        iso.formatOptions = [.withInternetDateTime]
        let day1 = Calendar.current.date(byAdding: .day, value: -3, to: Date())!
        let day2 = Date()
        mockStore.weightHistory = [
            WeightEntryResponse(id: "1", weight: 70.0, date: iso.string(from: day1)),
            WeightEntryResponse(id: "2", weight: 72.0, date: iso.string(from: day2))
        ]

        await sut.load()

        XCTAssertEqual(sut.weightPeriodChange ?? 0, 2.0, accuracy: 0.001)
    }

    func test_load_emptyWeightHistory_chartPointsIsEmpty() async throws {
        mockStore.weightHistory = []

        await sut.load()

        XCTAssertTrue(sut.chartPoints.isEmpty)
    }

    // MARK: - reloadChart() on period change

    func test_reloadChart_onPeriodChange_fetchesWeightDataAgain() async throws {
        sut.selectedMetric = .weight

        await sut.reloadChart()
        let callsAfterFirst = mockStore.fetchNutritionReportCallCount

        sut.selectedPeriod = .month
        await sut.reloadChart()

        XCTAssertEqual(callsAfterFirst, 0)
        XCTAssertEqual(mockStore.fetchNutritionReportCallCount, 0)
    }

    func test_reloadChart_caloriesMetric_callsFetchNutritionReport() async throws {
        sut.selectedMetric = .calories
        mockStore.fetchNutritionReportResult = .stub()

        await sut.reloadChart()

        XCTAssertEqual(mockStore.fetchNutritionReportCallCount, 1)
    }

    func test_reloadChart_caloriesMetric_populatesReportTotals() async throws {
        sut.selectedMetric = .calories
        mockStore.fetchNutritionReportResult = .stub(totalCalories: 9800, avgCaloriesPerDay: 1400)

        await sut.reloadChart()

        XCTAssertEqual(sut.reportPeriodTotalCalories, 9800)
        XCTAssertEqual(sut.reportPeriodAvgCalories, 1400)
    }

    func test_reloadChart_caloriesMetric_nilReport_clearsChartPoints() async throws {
        sut.selectedMetric = .calories
        mockStore.fetchNutritionReportResult = nil

        await sut.reloadChart()

        XCTAssertTrue(sut.chartPoints.isEmpty)
        XCTAssertEqual(sut.reportPeriodTotalCalories, 0)
        XCTAssertEqual(sut.reportPeriodAvgCalories, 0)
    }

    func test_reloadChart_stepsMetric_callsFetchDailySteps() async throws {
        sut.selectedMetric = .steps

        await sut.reloadChart()

        XCTAssertEqual(mockStore.fetchDailyStepsCallCount, 1)
    }

    func test_reloadChart_stepsMetric_populatesChartPointsAndAverage() async throws {
        sut.selectedMetric = .steps
        let now = Date()
        mockStore.fetchDailyStepsResult = [
            DailySteps(date: Calendar.current.date(byAdding: .day, value: -1, to: now)!, count: 8000),
            DailySteps(date: now, count: 12000)
        ]

        await sut.reloadChart()

        XCTAssertEqual(sut.chartPoints.count, 2)
        XCTAssertEqual(sut.stepsAverage, 10000)
    }

    func test_reloadChart_stepsMetric_emptySteps_stepsAverageIsZero() async throws {
        sut.selectedMetric = .steps
        mockStore.fetchDailyStepsResult = []

        await sut.reloadChart()

        XCTAssertEqual(sut.stepsAverage, 0)
    }

    func test_reloadChart_switchingMetric_clearsPreviousData() async throws {
        sut.selectedMetric = .calories
        mockStore.fetchNutritionReportResult = .stub(totalCalories: 5000)
        await sut.reloadChart()
        XCTAssertEqual(sut.reportPeriodTotalCalories, 5000)

        sut.selectedMetric = .weight
        await sut.reloadChart()

        XCTAssertEqual(sut.reportPeriodTotalCalories, 0)
    }

    // MARK: - addWeight()

    func test_addWeight_callsStoreAddWeight() async throws {
        try await sut.addWeight(75.5)

        XCTAssertEqual(mockStore.addWeightCallCount, 1)
        XCTAssertEqual(mockStore.lastAddedWeight, 75.5)
    }

    func test_addWeight_setsShowWeightInputFalse() async throws {
        sut.showWeightInput = true

        try await sut.addWeight(80.0)

        XCTAssertFalse(sut.showWeightInput)
    }

    func test_addWeight_callsReloadChart_afterSuccess() async throws {
        sut.selectedMetric = .calories
        mockStore.fetchNutritionReportResult = .stub()

        try await sut.addWeight(70.0)

        XCTAssertEqual(mockStore.fetchNutritionReportCallCount, 1)
    }

    func test_addWeight_throwsWhenStoreFails() async throws {
        mockStore.addWeightResult = .failure(NetworkError.requestFailed(statusCode: 500, message: ""))

        do {
            try await sut.addWeight(70.0)
            XCTFail("Expected error to be thrown")
        } catch {
            XCTAssertNotNil(error)
        }
    }

    func test_addWeight_showWeightInput_remainsTrueOnFailure() async throws {
        sut.showWeightInput = true
        mockStore.addWeightResult = .failure(NetworkError.requestFailed(statusCode: 500, message: ""))

        try? await sut.addWeight(70.0)

        XCTAssertTrue(sut.showWeightInput)
    }

    // MARK: - weightGoalAssessment

    func test_weightGoalAssessment_nilWhenNoPeriodChange() {
        sut.weightPeriodChange = nil

        XCTAssertNil(sut.weightGoalAssessment)
    }

    func test_weightGoalAssessment_loseWeight_positiveProgress() {
        mockUserStore.setProfileValue(.stub(goal: .loseWeight))
        sut.weightPeriodChange = -0.5

        XCTAssertEqual(sut.weightGoalAssessment, "Отличный прогресс — вы движетесь к цели!")
    }

    func test_weightGoalAssessment_loseWeight_noProgress() {
        mockUserStore.setProfileValue(.stub(goal: .loseWeight))
        sut.weightPeriodChange = 0.3

        XCTAssertEqual(sut.weightGoalAssessment, "Продолжайте держать дефицит калорий")
    }

    func test_weightGoalAssessment_loseWeight_borderlineChange() {
        mockUserStore.setProfileValue(.stub(goal: .loseWeight))
        sut.weightPeriodChange = -0.1

        XCTAssertEqual(sut.weightGoalAssessment, "Продолжайте держать дефицит калорий")
    }

    func test_weightGoalAssessment_gainMuscle_positiveProgress() {
        mockUserStore.setProfileValue(.stub(goal: .gainMuscle))
        sut.weightPeriodChange = 0.5

        XCTAssertEqual(sut.weightGoalAssessment, "Набор идёт по плану — так держать!")
    }

    func test_weightGoalAssessment_gainMuscle_noProgress() {
        mockUserStore.setProfileValue(.stub(goal: .gainMuscle))
        sut.weightPeriodChange = -0.2

        XCTAssertEqual(sut.weightGoalAssessment, "Попробуйте добавить калорий в рацион")
    }

    func test_weightGoalAssessment_gainMuscle_borderlineChange() {
        mockUserStore.setProfileValue(.stub(goal: .gainMuscle))
        sut.weightPeriodChange = 0.1

        XCTAssertEqual(sut.weightGoalAssessment, "Попробуйте добавить калорий в рацион")
    }

    func test_weightGoalAssessment_maintain_stableWeight() {
        mockUserStore.setProfileValue(.stub(goal: .maintain))
        sut.weightPeriodChange = 0.3

        XCTAssertEqual(sut.weightGoalAssessment, "Вы отлично держите вес!")
    }

    func test_weightGoalAssessment_maintain_exactlyAtBoundary() {
        mockUserStore.setProfileValue(.stub(goal: .maintain))
        sut.weightPeriodChange = 0.5

        XCTAssertEqual(sut.weightGoalAssessment, "Вы отлично держите вес!")
    }

    func test_weightGoalAssessment_maintain_weightDecreasing() {
        mockUserStore.setProfileValue(.stub(goal: .maintain))
        sut.weightPeriodChange = -0.8

        XCTAssertEqual(sut.weightGoalAssessment, "Вес снижается — следите за рационом")
    }

    func test_weightGoalAssessment_maintain_weightIncreasing() {
        mockUserStore.setProfileValue(.stub(goal: .maintain))
        sut.weightPeriodChange = 0.8

        XCTAssertEqual(sut.weightGoalAssessment, "Вес растёт — следите за рационом")
    }

    func test_weightGoalAssessment_nilProfile_defaultsToMaintain_stableWeight() {
        mockUserStore.setProfileValue(nil)
        sut.weightPeriodChange = 0.2

        XCTAssertEqual(sut.weightGoalAssessment, "Вы отлично держите вес!")
    }

    // MARK: - isWeightChangeGood

    func test_isWeightChangeGood_trueWhenNoPeriodChange() {
        sut.weightPeriodChange = nil

        XCTAssertTrue(sut.isWeightChangeGood)
    }

    func test_isWeightChangeGood_loseWeight_trueWhenNegative() {
        mockUserStore.setProfileValue(.stub(goal: .loseWeight))
        sut.weightPeriodChange = -1.0

        XCTAssertTrue(sut.isWeightChangeGood)
    }

    func test_isWeightChangeGood_loseWeight_falseWhenPositive() {
        mockUserStore.setProfileValue(.stub(goal: .loseWeight))
        sut.weightPeriodChange = 0.5

        XCTAssertFalse(sut.isWeightChangeGood)
    }

    func test_isWeightChangeGood_gainMuscle_trueWhenPositive() {
        mockUserStore.setProfileValue(.stub(goal: .gainMuscle))
        sut.weightPeriodChange = 0.5

        XCTAssertTrue(sut.isWeightChangeGood)
    }

    func test_isWeightChangeGood_gainMuscle_falseWhenNegative() {
        mockUserStore.setProfileValue(.stub(goal: .gainMuscle))
        sut.weightPeriodChange = -0.5

        XCTAssertFalse(sut.isWeightChangeGood)
    }

    func test_isWeightChangeGood_maintain_trueWithinBounds() {
        mockUserStore.setProfileValue(.stub(goal: .maintain))
        sut.weightPeriodChange = 0.5

        XCTAssertTrue(sut.isWeightChangeGood)
    }

    func test_isWeightChangeGood_maintain_falseOutsideBounds() {
        mockUserStore.setProfileValue(.stub(goal: .maintain))
        sut.weightPeriodChange = 0.6

        XCTAssertFalse(sut.isWeightChangeGood)
    }

    // MARK: - exportCSV()

    func test_exportCSV_weightMetric_setsExportFileURL() {
        sut.selectedMetric = .weight
        sut.chartPoints = []

        sut.exportCSV()

        XCTAssertNotNil(sut.exportFileURL)
    }

    func test_exportCSV_caloriesMetric_setsExportFileURL() {
        sut.selectedMetric = .calories
        sut.chartPoints = []

        sut.exportCSV()

        XCTAssertNotNil(sut.exportFileURL)
    }

    func test_exportCSV_stepsMetric_setsExportFileURL() {
        sut.selectedMetric = .steps
        sut.chartPoints = []

        sut.exportCSV()

        XCTAssertNotNil(sut.exportFileURL)
    }

    func test_exportCSV_createsFileAtURL() throws {
        let date = Calendar.current.date(byAdding: .day, value: -1, to: Date())!
        sut.selectedMetric = .weight
        sut.chartPoints = [ChartDataPoint(date: date, value: 70.5)]

        sut.exportCSV()

        let url = try XCTUnwrap(sut.exportFileURL)
        XCTAssertTrue(FileManager.default.fileExists(atPath: url.path))
    }

    func test_exportCSV_fileContainsHeader_forWeightMetric() throws {
        sut.selectedMetric = .weight
        sut.chartPoints = []

        sut.exportCSV()

        let url = try XCTUnwrap(sut.exportFileURL)
        let content = try String(contentsOf: url, encoding: .utf8)
        XCTAssertTrue(content.hasPrefix("Дата,Вес (кг)"))
    }

    func test_exportCSV_fileContainsHeader_forCaloriesMetric() throws {
        sut.selectedMetric = .calories
        sut.chartPoints = []

        sut.exportCSV()

        let url = try XCTUnwrap(sut.exportFileURL)
        let content = try String(contentsOf: url, encoding: .utf8)
        XCTAssertTrue(content.hasPrefix("Дата,Калории (ккал)"))
    }

    func test_exportCSV_fileContainsHeader_forStepsMetric() throws {
        sut.selectedMetric = .steps
        sut.chartPoints = []

        sut.exportCSV()

        let url = try XCTUnwrap(sut.exportFileURL)
        let content = try String(contentsOf: url, encoding: .utf8)
        XCTAssertTrue(content.hasPrefix("Дата,Шаги"))
    }

    func test_exportCSV_fileContainsDataRow() throws {
        let date = Calendar.current.date(byAdding: .day, value: -1, to: Date())!
        sut.selectedMetric = .weight
        sut.chartPoints = [ChartDataPoint(date: date, value: 72.3)]

        sut.exportCSV()

        let url = try XCTUnwrap(sut.exportFileURL)
        let content = try String(contentsOf: url, encoding: .utf8)
        XCTAssertTrue(content.contains("72.3"))
    }

    func test_exportCSV_fileNameContainsMetricAndPeriod() throws {
        sut.selectedMetric = .weight
        sut.selectedPeriod = .month
        sut.chartPoints = []

        sut.exportCSV()

        let url = try XCTUnwrap(sut.exportFileURL)
        XCTAssertTrue(url.lastPathComponent.contains("вес"))
        XCTAssertTrue(url.lastPathComponent.contains("месяц"))
    }

    // MARK: - weightHistory / recommendations passthrough

    func test_weightHistory_reflectsStoreValue() {
        let iso = ISO8601DateFormatter()
        iso.formatOptions = [.withInternetDateTime]
        mockStore.weightHistory = [
            WeightEntryResponse(id: "1", weight: 68.0, date: iso.string(from: Date()))
        ]

        XCTAssertEqual(sut.weightHistory.count, 1)
        XCTAssertEqual(sut.weightHistory.first?.weight, 68.0)
    }

    func test_recommendations_reflectsStoreValue() {
        mockStore.recommendations = [
            RecommendationResponse(
                id: "r1", type: "diet", description: "Eat more protein",
                priority: 1, isRead: false, generatedAt: "2026-01-01T00:00:00Z"
            )
        ]

        XCTAssertEqual(sut.recommendations.count, 1)
    }

    func test_isLoadingRecommendations_reflectsStoreValue() {
        mockStore.isLoadingRecommendations = true

        XCTAssertTrue(sut.isLoadingRecommendations)
    }

    // MARK: - refreshRecommendations()

    func test_refreshRecommendations_callsStoreRefresh() async throws {
        await sut.refreshRecommendations()

        XCTAssertEqual(mockStore.refreshRecommendationsCallCount, 1)
    }

    func test_refreshRecommendations_callsMarkAllRead() async throws {
        await sut.refreshRecommendations()

        XCTAssertEqual(mockStore.markAllReadCallCount, 1)
    }
}
