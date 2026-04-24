import Foundation
import Combine

@MainActor
final class HomeViewModel: ObservableObject {
    @Published var state: ScreenState = .loading
    @Published var goals: GoalTargets?
    @Published var stats: DayStats?
    @Published var basalMetabolicRate: Int?
    @Published var meals: [MealPreview] = []

    private let health: HealthKitServiceProtocol = HealthKitService.shared

    private var nutritionCancellable: AnyCancellable?
    private var userCancellable: AnyCancellable?
    private var summaryCancellable: AnyCancellable?
    private var burnedCancellable: AnyCancellable?
    private var stepsCancellable: AnyCancellable?

    init() {
        nutritionCancellable = NutritionStore.shared.$mealPreviews
            .receive(on: RunLoop.main)
            .sink { [weak self] previews in
                self?.meals = previews
            }

        userCancellable = UserStore.shared.$targetCalories
            .receive(on: RunLoop.main)
            .sink { [weak self] calories in
                self?.goals = GoalTargets(steps: 10000, calories: calories)
            }

        summaryCancellable = NutritionStore.shared.$dailySummary
            .dropFirst()
            .receive(on: RunLoop.main)
            .sink { [weak self] summary in
                guard let self, let summary, let existing = self.stats else { return }
                self.stats = DayStats(
                    steps: existing.steps,
                    caloriesConsumed: summary.consumed.calories,
                    caloriesBurned: UserStore.shared.caloriesBurned
                )
            }

        burnedCancellable = UserStore.shared.$caloriesBurned
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
        
        stepsCancellable = UserStore.shared.$todaySteps
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
    }

    func load() async {
        state = .loading
        do {
            try await NutritionStore.shared.load()
        } catch {
            if AppRouter.shared.handleIfUnauthorized(error) { return }
            let appError = ErrorMapper.map(error)
            state = .error(appError.errorDescription ?? "Не удалось загрузить данные питания")
            return
        }

        await UserStore.shared.load()

        await HealthKitService.shared.refreshDailyActivity()

        let steps = UserStore.shared.todaySteps
        let burned = UserStore.shared.caloriesBurned

        self.stats = DayStats(
            steps: steps,
            caloriesConsumed: NutritionStore.shared.dailySummary?.consumed.calories ?? 0,
            caloriesBurned: burned
        )
        self.goals = GoalTargets(
            steps: 10000,
            calories: UserStore.shared.targetCalories
        )
        self.basalMetabolicRate = UserStore.shared.basalMetabolicRate > 0
            ? UserStore.shared.basalMetabolicRate : nil

        state = .loaded
    }
}
