import Foundation
import Combine

// MARK: - Mutation Types

enum MutationType: String, Codable {
    case addMeal
    case deleteMeal
    case addWeight
    case markRecommendationRead
}

struct QueuedMutation: Codable, Identifiable {
    let id: UUID
    let type: MutationType
    let payload: Data
    let timestamp: Date
    var retryCount: Int

    static let maxRetries = 3
}

// MARK: - Payloads

struct AddMealPayload: Codable {
    let meal: Meal
}

struct DeleteMealPayload: Codable {
    let mealId: String
}

struct AddWeightPayload: Codable {
    let weight: Double
}

struct MarkRecommendationReadPayload: Codable {
    let id: String
}

// MARK: - Queue

@MainActor
final class MutationQueue: ObservableObject {
    static let shared = MutationQueue()

    @Published private(set) var mutations: [QueuedMutation] = []

    var pendingCount: Int { mutations.count }

    private let diskCache = DiskCache.shared
    private let sessionManager = UserSessionManager.shared
    private let encoder: JSONEncoder = {
        let e = JSONEncoder()
        e.dateEncodingStrategy = .iso8601
        return e
    }()

    private var cacheKey: String {
        "mutation_queue_\(sessionManager.currentUserId ?? "anon")"
    }

    private init() {
        reload()
    }

    // MARK: - Lifecycle

    func reload() {
        mutations = diskCache.load([QueuedMutation].self, key: cacheKey) ?? []
        print("[INFO] [MutationQueue]: Loaded \(mutations.count) pending mutations")
    }

    func clear() {
        mutations = []
        diskCache.remove(key: cacheKey)
    }

    // MARK: - Enqueue

    func enqueue(type: MutationType, payload: some Encodable) {
        guard let data = try? encoder.encode(payload) else { return }
        let mutation = QueuedMutation(
            id: UUID(),
            type: type,
            payload: data,
            timestamp: Date(),
            retryCount: 0
        )
        mutations.append(mutation)
        persist()
        print("[INFO] [MutationQueue]: Enqueued \(type.rawValue), total: \(mutations.count)")
    }

    // MARK: - Queue management

    func remove(id: UUID) {
        mutations.removeAll { $0.id == id }
        persist()
    }

    func incrementRetry(id: UUID) {
        guard let idx = mutations.firstIndex(where: { $0.id == id }) else { return }
        mutations[idx].retryCount += 1
        if mutations[idx].retryCount >= QueuedMutation.maxRetries {
            print("[WARN] [MutationQueue]: Dropping \(mutations[idx].type.rawValue) after max retries")
            mutations.remove(at: idx)
        }
        persist()
    }

    // MARK: - Private

    private func persist() {
        diskCache.save(mutations, key: cacheKey)
    }
}
