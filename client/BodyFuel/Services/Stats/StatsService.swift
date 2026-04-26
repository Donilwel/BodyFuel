import Foundation

protocol StatsServiceProtocol {
    func fetchWeightHistory() async throws -> [WeightEntryResponse]
    func addWeight(_ weight: Double) async throws
    func fetchCaloriesHistory(from: Date, to: Date) async throws -> [CaloriesHistoryEntryResponse]
    func fetchNutritionReport(from: Date, to: Date) async throws -> NutritionReportResponse
    func fetchRecommendations(page: Int, limit: Int) async throws -> [RecommendationResponse]
    func refreshRecommendations() async throws -> [RecommendationResponse]
    func markRecommendationRead(id: String) async throws
}

final class StatsService: StatsServiceProtocol {
    static let shared = StatsService()

    private let networkClient = NetworkClient.shared
    private init() {}

    private static let isoFormatter: ISO8601DateFormatter = {
        let f = ISO8601DateFormatter()
        f.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return f
    }()

    private static let dateFormatter: DateFormatter = {
        let f = DateFormatter()
        f.dateFormat = "yyyy-MM-dd'T'HH:mm:ssZ"
        return f
    }()

    // MARK: Weight

    func fetchWeightHistory() async throws -> [WeightEntryResponse] {
        guard let url = URL(string: API.baseURLString + API.weightHistory) else {
            throw NetworkError.invalidURL
        }
        let response: [WeightEntryResponse] = try await networkClient.request(url: url, method: .get)
        print("[INFO] [StatsService/fetchWeightHistory]: \(response.count) entries")
        return response
    }

    func addWeight(_ weight: Double) async throws {
        guard let url = URL(string: API.baseURLString + API.weight) else {
            throw NetworkError.invalidURL
        }
        let _: DefaultDecodable = try await networkClient.request(
            url: url,
            method: .post,
            requestBody: AddWeightRequestBody(weight: weight)
        )
        print("[INFO] [StatsService/addWeight]: Added \(weight) kg")
    }

    // MARK: Calories History

    func fetchCaloriesHistory(from startDate: Date, to endDate: Date) async throws -> [CaloriesHistoryEntryResponse] {
        var components = URLComponents(string: API.baseURLString + API.caloriesHistory)
        components?.queryItems = [
            URLQueryItem(name: "start_date", value: Self.isoFormatter.string(from: startDate)),
            URLQueryItem(name: "end_date",   value: Self.isoFormatter.string(from: endDate))
        ]
        guard let url = components?.url else { throw NetworkError.invalidURL }
        let response: [CaloriesHistoryEntryResponse] = try await networkClient.request(url: url, method: .get)
        print("[INFO] [StatsService/fetchCaloriesHistory]: \(response.count) entries")
        return response
    }

    // MARK: Nutrition Report

    func fetchNutritionReport(from startDate: Date, to endDate: Date) async throws -> NutritionReportResponse {
        var components = URLComponents(string: API.baseURLString + API.Nutrition.report)
        let df = DateFormatter()
        df.dateFormat = "yyyy-MM-dd"
        components?.queryItems = [
            URLQueryItem(name: "from", value: df.string(from: startDate)),
            URLQueryItem(name: "to",   value: df.string(from: endDate))
        ]
        guard let url = components?.url else { throw NetworkError.invalidURL }
        let response: NutritionReportResponse = try await networkClient.request(url: url, method: .get)
        print("[INFO] [StatsService/fetchNutritionReport]: \(response.entries.count) entries, \(response.days) days")
        return response
    }

    // MARK: Recommendations

    func fetchRecommendations(page: Int = 1, limit: Int = 10) async throws -> [RecommendationResponse] {
        var components = URLComponents(string: API.baseURLString + API.recommendations)
        components?.queryItems = [
            URLQueryItem(name: "page",  value: "\(page)"),
            URLQueryItem(name: "limit", value: "\(limit)")
        ]
        guard let url = components?.url else { throw NetworkError.invalidURL }
        let response: [RecommendationResponse] = try await networkClient.request(url: url, method: .get)
        print("[INFO] [StatsService/fetchRecommendations]: \(response.count) recommendations")
        return response
    }

    func refreshRecommendations() async throws -> [RecommendationResponse] {
        guard let url = URL(string: API.baseURLString + API.recommendationsRefresh) else {
            throw NetworkError.invalidURL
        }
        let response: [RecommendationResponse] = try await networkClient.request(url: url, method: .post)
        print("[INFO] [StatsService/refreshRecommendations]: \(response.count) new recommendations")
        return response
    }

    func markRecommendationRead(id: String) async throws {
        guard let url = URL(string: API.baseURLString + API.recommendations + "/\(id)/read") else {
            throw NetworkError.invalidURL
        }
        let _: DefaultDecodable = try await networkClient.request(url: url, method: .patch)
    }
}
