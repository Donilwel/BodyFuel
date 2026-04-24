import Foundation
import Combine

@MainActor
final class StatsStore: ObservableObject {
    static let shared = StatsStore()

    @Published var weightHistory: [WeightEntryResponse] = []
    @Published var recommendations: [RecommendationResponse] = []
    @Published var isLoadingWeight = false
    @Published var isLoadingRecommendations = false
    @Published var isWeightDataStale = false
    @Published var isRecommendationsStale = false

    private let statsService: StatsServiceProtocol = StatsService.shared
    private let healthKitService: HealthKitServiceProtocol = HealthKitService.shared
    private let diskCache = DiskCache.shared
    private let mutationQueue = MutationQueue.shared
    private let sessionManager = UserSessionManager.shared

    private static let weightTTL: TimeInterval = 24 * 60 * 60
    private static let recsTTL: TimeInterval = 60 * 60

    private var weightKey: String { "stats_weight_\(sessionManager.currentUserId ?? "anon")" }
    private var recsKey: String   { "stats_recs_\(sessionManager.currentUserId ?? "anon")" }

    private var reconnectCancellable: AnyCancellable?

    private init() {
        reconnectCancellable = NetworkMonitor.shared.$isOnline
            .dropFirst()
            .filter { $0 }
            .receive(on: RunLoop.main)
            .sink { [weak self] _ in
                guard let self, self.isWeightDataStale else { return }
                Task { await self.loadWeightHistory(force: true) }
            }
    }

    // MARK: Weight

    func loadWeightHistory(force: Bool = false) async {
        let expired = diskCache.isExpired(key: weightKey, ttl: Self.weightTTL)
        if !force, !expired, !weightHistory.isEmpty { return }
        guard !isLoadingWeight else { return }

        isLoadingWeight = true
        defer { isLoadingWeight = false }

        if !force, !NetworkMonitor.shared.isOnline {
            loadWeightFromDisk()
            return
        }

        do {
            let result = try await statsService.fetchWeightHistory()
            weightHistory = applyPendingWeightMutations(to: result.sorted { $0.date < $1.date })
            isWeightDataStale = false
            diskCache.save(weightHistory, key: weightKey)
        } catch {
            if isAuthError(error) { return }
            loadWeightFromDisk()
        }
    }

    func addWeight(_ weight: Double) async throws {
        do {
            try await statsService.addWeight(weight)
            invalidateWeightCache()
            await loadWeightHistory(force: true)
        } catch {
            if isAuthError(error) { throw error }
            if isTransportError(error) {
                let optimistic = WeightEntryResponse(
                    id: UUID().uuidString,
                    weight: weight,
                    date: ISO8601DateFormatter().string(from: Date())
                )
                weightHistory.append(optimistic)
                weightHistory.sort { $0.date < $1.date }
                diskCache.save(weightHistory, key: weightKey)
                mutationQueue.enqueue(type: .addWeight, payload: AddWeightPayload(weight: weight))
            } else {
                throw error
            }
        }
    }

    // MARK: Recommendations

    func loadRecommendations(force: Bool = false) async {
        let expired = diskCache.isExpired(key: recsKey, ttl: Self.recsTTL)
        if !force, !expired, !recommendations.isEmpty { return }

        if !NetworkMonitor.shared.isOnline {
            loadRecommendationsFromDisk()
            return
        }

        isLoadingRecommendations = true
        defer { isLoadingRecommendations = false }

        do {
            let result = try await statsService.fetchRecommendations(page: 1, limit: 10)
            recommendations = result
            isRecommendationsStale = false
            diskCache.save(recommendations, key: recsKey)
        } catch {
            if isAuthError(error) { return }
            loadRecommendationsFromDisk()
        }
    }

    func refreshRecommendations() async {
        isLoadingRecommendations = true
        defer { isLoadingRecommendations = false }
        if let result = try? await statsService.refreshRecommendations() {
            recommendations = result
            isRecommendationsStale = false
            diskCache.save(recommendations, key: recsKey)
        }
    }

    func markAllRead() {
        let unread = recommendations.filter { !$0.isRead }
        guard !unread.isEmpty else { return }
        recommendations = recommendations.map { rec in
            guard !rec.isRead else { return rec }
            return RecommendationResponse(
                id: rec.id, type: rec.type, description: rec.description,
                priority: rec.priority, isRead: true, generatedAt: rec.generatedAt
            )
        }
        diskCache.save(recommendations, key: recsKey)
        for rec in unread {
            Task { try? await self.statsService.markRecommendationRead(id: rec.id) }
        }
    }

    func markRead(id: String) async {
        updateReadStatus(id: id, isRead: true)
        do {
            try await statsService.markRecommendationRead(id: id)
        } catch {
            if !isAuthError(error), isTransportError(error) {
                mutationQueue.enqueue(
                    type: .markRecommendationRead,
                    payload: MarkRecommendationReadPayload(id: id)
                )
            }
        }
        diskCache.save(recommendations, key: recsKey)
    }

    // MARK: Calories / Nutrition Report

    func fetchCaloriesHistory(from: Date, to: Date) async -> [CaloriesHistoryEntryResponse] {
        return (try? await statsService.fetchCaloriesHistory(from: from, to: to)) ?? []
    }

    func fetchNutritionReport(from: Date, to: Date) async -> NutritionReportResponse? {
        let key = nutritionReportKey(from: from, to: to)

        if !NetworkMonitor.shared.isOnline {
            let cached = diskCache.load(NutritionReportResponse.self, key: key)
            print("[INFO] [StatsStore]: Nutrition report from disk (offline), found=\(cached != nil)")
            return cached
        }

        if let result = try? await statsService.fetchNutritionReport(from: from, to: to) {
            diskCache.save(result, key: key)
            return result
        }

        return diskCache.load(NutritionReportResponse.self, key: key)
    }

    private func nutritionReportKey(from: Date, to: Date) -> String {
        let fmt = DateFormatter()
        fmt.dateFormat = "yyyy-MM-dd"
        return "stats_nutrition_\(sessionManager.currentUserId ?? "anon")_\(fmt.string(from: from))_\(fmt.string(from: to))"
    }

    // MARK: Steps (HealthKit)

    func fetchDailySteps(from: Date, to: Date) async -> [DailySteps] {
        return await healthKitService.fetchDailySteps(from: from, to: to)
    }

    // MARK: Cache control

    func invalidateWeightCache() {
        diskCache.remove(key: weightKey)
    }

    func invalidateRecommendationsCache() {
        diskCache.remove(key: recsKey)
    }

    func reset() {
        weightHistory = []
        recommendations = []
        isWeightDataStale = false
        isRecommendationsStale = false
        diskCache.remove(key: weightKey)
        diskCache.remove(key: recsKey)
    }

    // MARK: Private

    private func loadWeightFromDisk() {
        if let cached = diskCache.load([WeightEntryResponse].self, key: weightKey) {
            weightHistory = applyPendingWeightMutations(to: cached)
            isWeightDataStale = true
            print("[INFO] [StatsStore]: Loaded \(cached.count) weight entries from disk (stale)")
        }
    }

    private func loadRecommendationsFromDisk() {
        if let cached = diskCache.load([RecommendationResponse].self, key: recsKey) {
            recommendations = cached
            isRecommendationsStale = true
            print("[INFO] [StatsStore]: Loaded \(cached.count) recommendations from disk (stale)")
        }
    }

    private func applyPendingWeightMutations(to base: [WeightEntryResponse]) -> [WeightEntryResponse] {
        var result = base
        let decoder = JSONDecoder()
        for mutation in mutationQueue.mutations where mutation.type == .addWeight {
            if let p = try? decoder.decode(AddWeightPayload.self, from: mutation.payload) {
                let entry = WeightEntryResponse(
                    id: mutation.id.uuidString,
                    weight: p.weight,
                    date: ISO8601DateFormatter().string(from: mutation.timestamp)
                )
                result.append(entry)
            }
        }
        return result.sorted { $0.date < $1.date }
    }

    private func updateReadStatus(id: String, isRead: Bool) {
        guard let idx = recommendations.firstIndex(where: { $0.id == id }) else { return }
        let old = recommendations[idx]
        recommendations[idx] = RecommendationResponse(
            id: old.id, type: old.type, description: old.description,
            priority: old.priority, isRead: isRead, generatedAt: old.generatedAt
        )
    }
}
