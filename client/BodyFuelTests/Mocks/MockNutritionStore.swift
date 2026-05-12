import Foundation
import Combine
@testable import BodyFuel

@MainActor
final class MockNutritionStore: NutritionStoreProtocol {

    // MARK: - Publishers

    private let mealsSubject = CurrentValueSubject<[Meal], Never>([])
    private let summarySubject = CurrentValueSubject<NutritionDailySummary?, Never>(nil)

    var mealsPublisher: AnyPublisher<[Meal], Never> { mealsSubject.eraseToAnyPublisher() }
    var dailySummaryPublisher: AnyPublisher<NutritionDailySummary?, Never> { summarySubject.eraseToAnyPublisher() }

    // MARK: - Configurable state

    var loadResult: Result<Void, Error> = .success(())
    var addMealResult: Result<Void, Error> = .success(())

    // MARK: - Call tracking

    var loadCallCount = 0
    var addMealCallCount = 0
    var deleteMealCallCount = 0

    var lastAddedMeal: Meal?
    var lastDeletedMeal: Meal?

    // MARK: - Helpers to drive publisher

    func setMeals(_ meals: [Meal]) {
        mealsSubject.send(meals)
    }

    func setSummary(_ summary: NutritionDailySummary?) {
        summarySubject.send(summary)
    }

    // MARK: - Protocol

    func load() async throws {
        loadCallCount += 1
        _ = try loadResult.get()
    }

    func addMeal(_ meal: Meal) async throws {
        addMealCallCount += 1
        lastAddedMeal = meal
        _ = try addMealResult.get()
        // Emit updated meals list
        mealsSubject.send(mealsSubject.value + [meal])
    }

    func deleteMeal(_ meal: Meal) async {
        deleteMealCallCount += 1
        lastDeletedMeal = meal
        mealsSubject.send(mealsSubject.value.filter { $0.id != meal.id })
    }
}
