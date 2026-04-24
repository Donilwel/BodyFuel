import Foundation
import Combine
import WidgetKit

@MainActor
final class NutritionStore: ObservableObject {
    static let shared = NutritionStore()

    @Published var meals: [Meal] = []
    @Published var dailySummary: NutritionDailySummary?
    @Published var mealPreviews: [MealPreview] = []
    @Published var isLoading = false
    @Published var isDataStale = false

    private var sessionCacheLoaded = false
    private var reconnectCancellable: AnyCancellable?

    private let nutritionService: NutritionServiceProtocol = NutritionService.shared
    private let sharedWidgetStorage = SharedWidgetStorage.shared
    private let diskCache = DiskCache.shared
    private let mutationQueue = MutationQueue.shared
    private let sessionManager = UserSessionManager.shared

    private var mealsKey: String {
        "nutrition_meals_\(sessionManager.currentUserId ?? "anon")"
    }

    private init() {
        reconnectCancellable = NetworkMonitor.shared.$isOnline
            .dropFirst()
            .filter { $0 }
            .receive(on: RunLoop.main)
            .sink { [weak self] _ in
                guard let self, self.isDataStale else { return }
                self.sessionCacheLoaded = false
                Task { try? await self.load() }
            }
    }

    // MARK: - Load

    func load() async throws {
        guard !sessionCacheLoaded, !isLoading else { return }

        isLoading = true
        defer { isLoading = false }

        if !NetworkMonitor.shared.isOnline {
            loadFromDisk()
            return
        }

        do {
            try await loadFromServer()
        } catch {
            if isAuthError(error) { throw error }
            if isTransportError(error) {
                NetworkMonitor.shared.markServerUnreachable()
            }
            loadFromDisk()
        }
    }

    func forceReload() async {
        isLoading = true
        defer { isLoading = false }
        do {
            try await loadFromServer()
        } catch {
            if !isAuthError(error) { loadFromDisk() }
        }
    }

    private func loadFromServer() async throws {
        async let summaryReq = nutritionService.fetchDailySummary()
        async let mealsReq = nutritionService.fetchMeals()
        let (summary, fetchedMeals) = try await (summaryReq, mealsReq)

        NetworkMonitor.shared.markServerReachable()

        let todayMeals = deduplicatedByID(fetchedMeals.filter { Calendar.current.isDateInToday($0.time) })

        self.dailySummary = summary
        self.meals = todayMeals
        self.mealPreviews = buildMealPreviews(from: todayMeals)
        self.sessionCacheLoaded = true
        self.isDataStale = false

        diskCache.save(todayMeals, key: mealsKey)
        sharedWidgetStorage.saveTodayConsumedCalories(summary.consumed.calories)
        print("[INFO] [NutritionStore]: Loaded \(todayMeals.count) meals from server")
    }

    private func loadFromDisk() {
        guard let cached = diskCache.load([Meal].self, key: mealsKey),
              !diskCache.isFromDifferentDay(key: mealsKey) else {
            print("[INFO] [NutritionStore]: No valid disk cache")
            return
        }
        let combined = deduplicatedByID(applyPendingMutations(to: cached))
        self.meals = combined
        self.dailySummary = deriveSummary(from: combined)
        self.mealPreviews = buildMealPreviews(from: combined)
        self.isDataStale = true
        sharedWidgetStorage.saveTodayConsumedCalories(dailySummary?.consumed.calories ?? 0)
        print("[INFO] [NutritionStore]: Loaded \(combined.count) meals from disk (stale)")
    }

    // MARK: - Mutations

    func addMeal(_ meal: Meal) async throws {
        guard !meals.contains(where: { $0.id == meal.id }) else { return }

        meals.append(meal)
        mealPreviews = buildMealPreviews(from: meals)
        dailySummary = deriveSummary(from: meals)
        persistMealsToDisk()
        sharedWidgetStorage.saveTodayConsumedCalories(dailySummary?.consumed.calories ?? 0)
        WidgetCenter.shared.reloadAllTimelines()

        if !NetworkMonitor.shared.isOnline {
            mutationQueue.enqueue(type: .addMeal, payload: AddMealPayload(meal: meal))
            ToastService.shared.show("Добавлено. Синхронизируется при подключении.")
            return
        }

        Task {
            do {
                try await nutritionService.saveMeal(meal)
                NetworkMonitor.shared.markServerReachable()
                await refreshSummaryFromServer()
            } catch {
                if isAuthError(error) {
                    _ = AppRouter.shared.handleIfUnauthorized(error)
                    return
                }
                if isTransportError(error) { NetworkMonitor.shared.markServerUnreachable() }
                mutationQueue.enqueue(type: .addMeal, payload: AddMealPayload(meal: meal))
                ToastService.shared.show("Добавлено. Синхронизируется при подключении.")
            }
        }
    }

    func deleteMeal(_ meal: Meal) async {
        meals.removeAll { $0.id == meal.id }
        mealPreviews = buildMealPreviews(from: meals)
        dailySummary = deriveSummary(from: meals)
        persistMealsToDisk()
        sharedWidgetStorage.saveTodayConsumedCalories(dailySummary?.consumed.calories ?? 0)
        WidgetCenter.shared.reloadAllTimelines()

        if !NetworkMonitor.shared.isOnline {
            mutationQueue.enqueue(type: .deleteMeal, payload: DeleteMealPayload(mealId: meal.id.uuidString))
            return
        }

        Task {
            do {
                try await nutritionService.deleteFoodEntry(id: meal.id.uuidString)
                NetworkMonitor.shared.markServerReachable()
                await refreshSummaryFromServer()
            } catch {
                if isAuthError(error) { _ = AppRouter.shared.handleIfUnauthorized(error); return }
                if isTransportError(error) {
                    NetworkMonitor.shared.markServerUnreachable()
                    mutationQueue.enqueue(type: .deleteMeal, payload: DeleteMealPayload(mealId: meal.id.uuidString))
                }
            }
        }
    }

    // MARK: - Cache control

    func invalidate() { sessionCacheLoaded = false }

    func reset() {
        meals = []
        dailySummary = nil
        mealPreviews = []
        isDataStale = false
        sessionCacheLoaded = false
        diskCache.remove(key: mealsKey)
    }

    // MARK: - Private helpers

    private func refreshSummaryFromServer() async {
        if let summary = try? await nutritionService.fetchDailySummary() {
            dailySummary = summary
            sharedWidgetStorage.saveTodayConsumedCalories(summary.consumed.calories)
        }
    }

    private func persistMealsToDisk() {
        diskCache.save(meals, key: mealsKey)
    }

    private func deriveSummary(from meals: [Meal]) -> NutritionDailySummary {
        let consumed = MacroNutrients(
            protein: meals.reduce(0) { $0 + $1.macros.protein },
            fat: meals.reduce(0) { $0 + $1.macros.fat },
            carbs: meals.reduce(0) { $0 + $1.macros.carbs }
        )
        let targetCalories = sharedWidgetStorage.getTargetCalories() ?? 2000
        let burned = sharedWidgetStorage.getTodayBurnedCalories() ?? 0
        let goal = MacroNutrients(
            protein: Double(targetCalories) * 0.30 / 4,
            fat: Double(targetCalories) * 0.30 / 9,
            carbs: Double(targetCalories) * 0.40 / 4
        )
        return NutritionDailySummary(consumed: consumed, goal: goal, burned: burned)
    }

    private func applyPendingMutations(to base: [Meal]) -> [Meal] {
        var result = base
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        for mutation in mutationQueue.mutations {
            switch mutation.type {
            case .addMeal:
                if let p = try? decoder.decode(AddMealPayload.self, from: mutation.payload),
                   Calendar.current.isDateInToday(p.meal.time) {
                    result.append(p.meal)
                }
            case .deleteMeal:
                if let p = try? decoder.decode(DeleteMealPayload.self, from: mutation.payload),
                   let id = UUID(uuidString: p.mealId) {
                    result.removeAll { $0.id == id }
                }
            default: break
            }
        }
        return result
    }

    private func buildMealPreviews(from meals: [Meal]) -> [MealPreview] {
        MealType.allCases.compactMap { type in
            let group = meals.filter { $0.mealType == type }
            guard !group.isEmpty else { return nil }
            let total = group.reduce(0) { $0 + $1.macros.calories }
            return MealPreview(title: type.displayName, calories: total)
        }
    }

    private func deduplicatedByID(_ meals: [Meal]) -> [Meal] {
        var seen = Set<UUID>()
        return meals.filter { seen.insert($0.id).inserted }
    }
}
