import Foundation
@testable import BodyFuel

final class MockStatsService: StatsServiceProtocol {

    // MARK: - Call tracking

    var fetchWeightHistoryCallCount = 0
    var addWeightCallCount = 0
    var fetchCaloriesHistoryCallCount = 0
    var fetchNutritionReportCallCount = 0
    var fetchRecommendationsCallCount = 0
    var refreshRecommendationsCallCount = 0
    var markRecommendationReadCallCount = 0

    var lastAddedWeight: Double?
    var lastMarkedReadId: String?

    // MARK: - Configurable responses

    var fetchWeightHistoryResult: Result<[WeightEntryResponse], Error> = .success([])
    var addWeightResult: Result<Void, Error> = .success(())
    var fetchCaloriesHistoryResult: Result<[CaloriesHistoryEntryResponse], Error> = .success([])
    var fetchNutritionReportResult: Result<NutritionReportResponse, Error> = .success(.stub())
    var fetchRecommendationsResult: Result<[RecommendationResponse], Error> = .success([])
    var refreshRecommendationsResult: Result<[RecommendationResponse], Error> = .success([])
    var markRecommendationReadResult: Result<Void, Error> = .success(())

    // MARK: - Protocol

    func fetchWeightHistory() async throws -> [WeightEntryResponse] {
        fetchWeightHistoryCallCount += 1
        return try fetchWeightHistoryResult.get()
    }

    func addWeight(_ weight: Double) async throws {
        addWeightCallCount += 1
        lastAddedWeight = weight
        _ = try addWeightResult.get()
    }

    func fetchCaloriesHistory(from: Date, to: Date) async throws -> [CaloriesHistoryEntryResponse] {
        fetchCaloriesHistoryCallCount += 1
        return try fetchCaloriesHistoryResult.get()
    }

    func fetchNutritionReport(from: Date, to: Date) async throws -> NutritionReportResponse {
        fetchNutritionReportCallCount += 1
        return try fetchNutritionReportResult.get()
    }

    func fetchRecommendations(page: Int, limit: Int) async throws -> [RecommendationResponse] {
        fetchRecommendationsCallCount += 1
        return try fetchRecommendationsResult.get()
    }

    func refreshRecommendations() async throws -> [RecommendationResponse] {
        refreshRecommendationsCallCount += 1
        return try refreshRecommendationsResult.get()
    }

    func markRecommendationRead(id: String) async throws {
        markRecommendationReadCallCount += 1
        lastMarkedReadId = id
        _ = try markRecommendationReadResult.get()
    }
}
