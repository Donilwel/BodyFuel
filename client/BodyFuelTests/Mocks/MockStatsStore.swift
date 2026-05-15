import Foundation
@testable import BodyFuel

@MainActor
final class MockStatsStore: StatsStoreProtocol {

    // MARK: - State

    var weightHistory: [WeightEntryResponse] = []
    var recommendations: [RecommendationResponse] = []
    var isLoadingRecommendations: Bool = false

    // MARK: - Configurable responses

    var addWeightResult: Result<Void, Error> = .success(())
    var fetchNutritionReportResult: NutritionReportResponse? = nil
    var fetchDailyStepsResult: [DailySteps] = []

    // MARK: - Call tracking

    var loadWeightHistoryCallCount = 0
    var loadRecommendationsCallCount = 0
    var markAllReadCallCount = 0
    var addWeightCallCount = 0
    var refreshRecommendationsCallCount = 0
    var fetchNutritionReportCallCount = 0
    var fetchDailyStepsCallCount = 0

    var lastAddedWeight: Double?

    // MARK: - Protocol

    func loadWeightHistory() async {
        loadWeightHistoryCallCount += 1
    }

    func loadRecommendations() async {
        loadRecommendationsCallCount += 1
    }

    func markAllRead() {
        markAllReadCallCount += 1
    }

    func addWeight(_ weight: Double) async throws {
        addWeightCallCount += 1
        lastAddedWeight = weight
        try addWeightResult.get()
    }

    func refreshRecommendations() async {
        refreshRecommendationsCallCount += 1
    }

    func fetchNutritionReport(from: Date, to: Date) async -> NutritionReportResponse? {
        fetchNutritionReportCallCount += 1
        return fetchNutritionReportResult
    }

    func fetchDailySteps(from: Date, to: Date) async -> [DailySteps] {
        fetchDailyStepsCallCount += 1
        return fetchDailyStepsResult
    }
}
