import Foundation
import Combine

@MainActor
final class HomeViewModel: ObservableObject {

    enum ScreenState {
        case loading
        case loaded
        case error(String)
    }

    @Published var state: ScreenState = .loading
    @Published var goals: GoalTargets?
    @Published var stats: DayStats?
    @Published var basalMetabolicRate: Int?
    @Published var workout: WorkoutModel?
    @Published var meals: [MealPreview] = []

    private let health: HealthServiceProtocol
    private let nutrition: NutritionServiceProtocol
    private let goalsService: GoalsServiceProtocol
    
    private let sharedWidgetStorage = SharedWidgetStorage.shared

    init(
        health: HealthServiceProtocol = MockHealthService(),
        nutrition: NutritionServiceProtocol = MockNutritionService(),
        goalsService: GoalsServiceProtocol = MockGoalsService()
    ) {
        self.health = health
        self.nutrition = nutrition
        self.goalsService = goalsService
    }

    func load() async {
        state = .loading
        do {
            async let steps = health.fetchTodaySteps()
            async let consumed = nutrition.fetchTodayConsumedCalories()
            async let burned = nutrition.fetchTodayBurnedCalories()
            async let workout = goalsService.fetchTodayWorkout()
            async let meals = nutrition.fetchTodayMeals()

            self.stats = DayStats(
                steps: try await steps,
                caloriesConsumed: try await consumed,
                caloriesBurned: try await burned
            )
            self.goals = GoalTargets(
                steps: 10000,
                calories: sharedWidgetStorage.getTargetCalories() ?? 0
            )
            
            self.workout = try await workout
            self.meals = try await meals
            self.basalMetabolicRate = sharedWidgetStorage.getBasalMetabolicRate()
            
            state = .loaded
        } catch {
            state = .error("Не удалось загрузить данные")
        }
    }
}
