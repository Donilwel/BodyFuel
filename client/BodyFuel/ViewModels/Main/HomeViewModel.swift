import Foundation
import Combine

@MainActor
final class HomeViewModel: ObservableObject {
    @Published var state: ScreenState = .loading
    @Published var goals: GoalTargets?
    @Published var stats: DayStats?
    @Published var basalMetabolicRate: Int?
    @Published var meals: [MealPreview] = []
    @Published var hasTodayWorkout: Bool = false
    @Published var hasWeeklyGoalMet: Bool = false

    private let nutritionStore: NutritionStoreProtocol
    private let userStore: UserStoreProtocol
    private let workoutHistoryStore: WorkoutHistoryStoreProtocol
    private let health: HealthKitServiceProtocol

    private var nutritionCancellable: AnyCancellable?
    private var userCancellable: AnyCancellable?
    private var summaryCancellable: AnyCancellable?
    private var burnedCancellable: AnyCancellable?
    private var stepsCancellable: AnyCancellable?
    private var historyCancellable: AnyCancellable?

    init() {
        self.nutritionStore = NutritionStore.shared
        self.userStore = UserStore.shared
        self.workoutHistoryStore = WorkoutHistoryStore.shared
        self.health = HealthKitService.shared
        setupSubscriptions()
    }

    init(
        nutritionStore: NutritionStoreProtocol,
        userStore: UserStoreProtocol,
        workoutHistoryStore: WorkoutHistoryStoreProtocol,
        health: HealthKitServiceProtocol
    ) {
        self.nutritionStore = nutritionStore
        self.userStore = userStore
        self.workoutHistoryStore = workoutHistoryStore
        self.health = health
        setupSubscriptions()
    }

    private func setupSubscriptions() {
        nutritionCancellable = nutritionStore.mealPreviewsPublisher
            .receive(on: RunLoop.main)
            .sink { [weak self] previews in
                self?.meals = previews
            }

        userCancellable = userStore.targetCaloriesPublisher
            .receive(on: RunLoop.main)
            .sink { [weak self] calories in
                self?.goals = GoalTargets(steps: 10000, calories: calories)
            }

        summaryCancellable = nutritionStore.dailySummaryPublisher
            .dropFirst()
            .receive(on: RunLoop.main)
            .sink { [weak self] summary in
                guard let self, let summary, let existing = self.stats else { return }
                self.stats = DayStats(
                    steps: existing.steps,
                    caloriesConsumed: summary.consumed.calories,
                    caloriesBurned: self.userStore.caloriesBurned
                )
            }

        burnedCancellable = userStore.caloriesBurnedPublisher
            .dropFirst()
            .receive(on: RunLoop.main)
            .sink { [weak self] burned in
                guard let self, let existing = self.stats else { return }
                self.stats = DayStats(
                    steps: existing.steps,
                    caloriesConsumed: existing.caloriesConsumed,
                    caloriesBurned: burned
                )
            }

        stepsCancellable = userStore.todayStepsPublisher
            .dropFirst()
            .receive(on: RunLoop.main)
            .sink { [weak self] steps in
                guard let self, let existing = self.stats else { return }
                self.stats = DayStats(
                    steps: steps,
                    caloriesConsumed: existing.caloriesConsumed,
                    caloriesBurned: existing.caloriesBurned
                )
            }

        historyCancellable = workoutHistoryStore.workoutsPublisher
            .receive(on: RunLoop.main)
            .sink { [weak self] _ in
                self?.updateWorkoutGoalStatus()
            }
    }

    private func updateWorkoutGoalStatus() {
        hasTodayWorkout = workoutHistoryStore.todayCompletedCount > 0
        let target = userStore.profile?.targetWorkoutsWeekly ?? 0
        hasWeeklyGoalMet = target > 0 && workoutHistoryStore.thisWeekCompletedCount >= target
    }

    func load() async {
        state = .loading
        do {
            try await nutritionStore.load()
        } catch {
            if AppRouter.shared.handleIfUnauthorized(error) { return }
            let appError = ErrorMapper.map(error)
            state = .error(appError.errorDescription ?? "Не удалось загрузить данные питания")
            return
        }

        await userStore.load()
        await workoutHistoryStore.load()
        updateWorkoutGoalStatus()

        await health.refreshDailyActivity()

        let steps = userStore.todaySteps
        let burned = userStore.caloriesBurned

        self.stats = DayStats(
            steps: steps,
            caloriesConsumed: nutritionStore.dailySummary?.consumed.calories ?? 0,
            caloriesBurned: burned
        )
        self.goals = GoalTargets(
            steps: 10000,
            calories: userStore.targetCalories
        )
        self.basalMetabolicRate = userStore.basalMetabolicRate > 0
            ? userStore.basalMetabolicRate : nil

        state = .loaded
    }
}
