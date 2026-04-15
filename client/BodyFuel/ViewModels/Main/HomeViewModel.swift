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
    @Published var meals: [MealPreview] = []

    private let health: HealthKitServiceProtocol = HealthKitService.shared
    private let nutrition: NutritionServiceProtocol = NutritionService.shared
    
    private let sharedWidgetStorage = SharedWidgetStorage.shared

    func load() async {
        state = .loading
        do {
            async let steps = health.fetchTodaySteps()
            async let consumed = nutrition.fetchTodayConsumedCalories()
            async let burned = nutrition.fetchTodayBurnedCalories()
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
            
            self.meals = try await meals
            self.basalMetabolicRate = sharedWidgetStorage.getBasalMetabolicRate()
            
            state = .loaded
        } catch {
            if AppRouter.shared.handleIfUnauthorized(error) { return }
            let appError = ErrorMapper.map(error)
            state = .error(appError.errorDescription ?? "Не удалось загрузить данные")
        }
    }
}
