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
    @Published var workout: WorkoutPreview?
    @Published var meals: [MealPreview] = []

    private let health: HealthServiceProtocol
    private let nutrition: NutritionServiceProtocol
    private let goalsService: GoalsServiceProtocol

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
            async let eaten = nutrition.fetchTodayCalories()
            async let goals = goalsService.fetchGoals()
            async let workout = goalsService.fetchTodayWorkout()
            async let meals = nutrition.fetchTodayMeals()

            self.stats = DayStats(
                steps: try await steps,
                caloriesConsumed: try await eaten
            )
            self.goals = try await goals
            self.workout = try await workout
            self.meals = try await meals

            state = .loaded
        } catch {
            state = .error("Не удалось загрузить данные")
        }
    }
}
