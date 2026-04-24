import Foundation
import Combine

@MainActor
final class OfflineSyncManager: ObservableObject {
    static let shared = OfflineSyncManager()

    @Published private(set) var isSyncing = false

    private let nutritionService: NutritionServiceProtocol = NutritionService.shared
    private let statsService: StatsServiceProtocol = StatsService.shared
    private let queue = MutationQueue.shared
    private let decoder: JSONDecoder = {
        let d = JSONDecoder()
        d.dateDecodingStrategy = .iso8601
        return d
    }()

    private init() {}

    // MARK: - Flush

    func flush() async {
        let pending = queue.mutations
        guard !pending.isEmpty, !isSyncing else { return }

        isSyncing = true
        defer { isSyncing = false }

        print("[INFO] [OfflineSyncManager]: Flushing \(pending.count) mutations")

        for mutation in pending {
            do {
                try await execute(mutation)
                queue.remove(id: mutation.id)
            } catch let error as NetworkError {
                switch error {
                case .network:
                    print("[WARN] [OfflineSyncManager]: Network error, stopping flush")
                    return
                default:
                    queue.incrementRetry(id: mutation.id)
                    print("[WARN] [OfflineSyncManager]: Server error for \(mutation.type.rawValue): \(error)")
                }
            } catch {
                queue.incrementRetry(id: mutation.id)
            }
        }

        NutritionStore.shared.invalidate()
        StatsStore.shared.invalidateWeightCache()
        StatsStore.shared.invalidateRecommendationsCache()

        try? await NutritionStore.shared.load()
        await StatsStore.shared.loadWeightHistory(force: true)
    }

    // MARK: - Execute

    private func execute(_ mutation: QueuedMutation) async throws {
        switch mutation.type {
        case .addMeal:
            let p = try decoder.decode(AddMealPayload.self, from: mutation.payload)
            try await nutritionService.saveMeal(p.meal)

        case .deleteMeal:
            let p = try decoder.decode(DeleteMealPayload.self, from: mutation.payload)
            try await nutritionService.deleteFoodEntry(id: p.mealId)

        case .addWeight:
            let p = try decoder.decode(AddWeightPayload.self, from: mutation.payload)
            try await statsService.addWeight(p.weight)

        case .markRecommendationRead:
            let p = try decoder.decode(MarkRecommendationReadPayload.self, from: mutation.payload)
            try await statsService.markRecommendationRead(id: p.id)
        }
    }
}
