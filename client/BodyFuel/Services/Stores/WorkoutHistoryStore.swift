import Foundation
import Combine

// MARK: - Models

struct WorkoutHistoryItem: Identifiable, Codable {
    let id: String
    let title: String
    let level: String
    let status: String
    let totalCalories: Int
    let duration: Int64
    let date: Date
    let exercisesCount: Int
    let completedCount: Int
    let exercises: [WorkoutHistoryExercise]
}

struct WorkoutHistoryExercise: Identifiable, Codable {
    let exerciseID: String
    let name: String
    let sets: Int
    let reps: Int
    let calories: Int
    let status: String

    var id: String { exerciseID }
    var isCompleted: Bool { status == "completed" }
}

// MARK: - Store

@MainActor
final class WorkoutHistoryStore: ObservableObject {
    static let shared = WorkoutHistoryStore()

    @Published var workouts: [WorkoutHistoryItem] = []
    @Published var isLoading = false
    @Published var isStale = false

    private let diskCache = DiskCache.shared
    private let workoutService: WorkoutServiceProtocol = WorkoutService.shared
    private let iso = ISO8601DateFormatter()

    private var cacheKey: String {
        "workout_history_\(UserSessionManager.shared.currentUserId ?? "anon")"
    }

    private init() {}

    // MARK: - Public

    func load() async {
        if let cached = diskCache.load([WorkoutHistoryItem].self, key: cacheKey) {
            workouts = cached
        }
        if NetworkMonitor.shared.isOnline {
            Task { await fetchFromServer() }
        } else {
            isStale = !workouts.isEmpty
        }
    }

    func refresh() async {
        await fetchFromServer()
    }

    func invalidate() {
        diskCache.remove(key: cacheKey)
    }

    // MARK: - Computed

    var todayCompletedCount: Int {
        workouts.filter { $0.status == "workout_done" && Calendar.current.isDateInToday($0.date) }.count
    }

    var thisWeekCompletedCount: Int {
        guard let weekInterval = Calendar.current.dateInterval(of: .weekOfYear, for: Date()) else { return 0 }
        return workouts.filter { $0.status == "workout_done" && weekInterval.contains($0.date) }.count
    }

    // MARK: - Private

    private func fetchFromServer() async {
        isLoading = true
        defer { isLoading = false }

        do {
            let response = try await workoutService.fetchWorkoutHistory(limit: 100, offset: 0)
            let items = response.workouts.map { mapToHistoryItem($0) }
            workouts = items
            diskCache.save(items, key: cacheKey)
            isStale = false
            NetworkMonitor.shared.markServerReachable()
        } catch {
            if isTransportError(error) {
                NetworkMonitor.shared.markServerUnreachable()
                isStale = !workouts.isEmpty
            }
        }
    }

    private func mapToHistoryItem(_ body: WorkoutSummaryResponseBody) -> WorkoutHistoryItem {
        let date = iso.date(from: body.date) ?? Date()
        let exercises = body.exercises.map { ex in
            WorkoutHistoryExercise(
                exerciseID: ex.exerciseID,
                name: ex.name,
                sets: ex.sets,
                reps: ex.reps,
                calories: ex.calories,
                status: ex.status
            )
        }
        return WorkoutHistoryItem(
            id: body.id,
            title: mapWorkoutTitle(body.level),
            level: body.level,
            status: body.status,
            totalCalories: body.totalCalories,
            duration: body.duration,
            date: date,
            exercisesCount: body.exercisesCount,
            completedCount: body.completedCount,
            exercises: exercises
        )
    }

    private func mapWorkoutTitle(_ level: String) -> String {
        switch level {
        case "workout_light": return "Лёгкая тренировка"
        case "workout_middle": return "Средняя тренировка"
        case "workout_hard": return "Интенсивная тренировка"
        default: return "Тренировка"
        }
    }
}
