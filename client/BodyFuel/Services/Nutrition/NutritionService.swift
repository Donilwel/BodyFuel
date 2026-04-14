import Foundation

protocol NutritionServiceProtocol {
    func fetchTodayConsumedCalories() async throws -> Int
    func fetchTodayBurnedCalories() async throws -> Int
    func fetchTodayMeals() async throws -> [MealPreview]
    func fetchDailySummary() async throws -> NutritionDailySummary
    func fetchMeals() async throws -> [Meal]
    func saveMeal(_ meal: Meal) async throws
    func addMealByText(description: String, mealType: MealType) async throws -> Meal
    func analyzeMealFromPhoto(_ imageData: Data, mealType: MealType) async throws -> Meal
    func generateRecipes() async throws -> [Recipe]
}

final class NutritionService: NutritionServiceProtocol {
    static let shared = NutritionService()

    private let sharedWidgetStorage = SharedWidgetStorage.shared

    private init() {}

    func fetchTodayConsumedCalories() async throws -> Int {
        sharedWidgetStorage.saveTodayConsumedCalories(1420)
        return 1420
    }

    func fetchTodayBurnedCalories() async throws -> Int {
        sharedWidgetStorage.saveTodayBurnedCalories(345)
        return 345
    }

    func fetchTodayMeals() async throws -> [MealPreview] {
        [
            MealPreview(title: "Завтрак", calories: 420),
            MealPreview(title: "Обед", calories: 650),
            MealPreview(title: "Перекус", calories: 350)
        ]
    }

    func fetchDailySummary() async throws -> NutritionDailySummary {
        try? await Task.sleep(nanoseconds: 300_000_000)
        let consumed = MacroNutrients(protein: 98, fat: 52, carbs: 180)
        let goal = MacroNutrients(protein: 150, fat: 70, carbs: 250)
        return NutritionDailySummary(consumed: consumed, goal: goal, burned: 345)
    }

    func fetchMeals() async throws -> [Meal] {
        try? await Task.sleep(nanoseconds: 300_000_000)
        let now = Date()
        let calendar = Calendar.current
        return [
            Meal(
                name: "Овсянка с бананом",
                mealType: .breakfast,
                macros: MacroNutrients(protein: 12, fat: 6, carbs: 58),
                time: calendar.date(bySettingHour: 8, minute: 30, second: 0, of: now) ?? now
            ),
            Meal(
                name: "Яйца (2 шт)",
                mealType: .breakfast,
                macros: MacroNutrients(protein: 14, fat: 10, carbs: 1),
                time: calendar.date(bySettingHour: 8, minute: 35, second: 0, of: now) ?? now
            ),
            Meal(
                name: "Куриная грудка с рисом",
                mealType: .lunch,
                macros: MacroNutrients(protein: 45, fat: 8, carbs: 60),
                time: calendar.date(bySettingHour: 13, minute: 0, second: 0, of: now) ?? now
            ),
            Meal(
                name: "Салат из свежих овощей",
                mealType: .lunch,
                macros: MacroNutrients(protein: 3, fat: 5, carbs: 12),
                time: calendar.date(bySettingHour: 13, minute: 5, second: 0, of: now) ?? now
            ),
            Meal(
                name: "Греческий йогурт",
                mealType: .snack,
                macros: MacroNutrients(protein: 15, fat: 3, carbs: 8),
                time: calendar.date(bySettingHour: 16, minute: 0, second: 0, of: now) ?? now
            ),
            Meal(
                name: "Орехи (30 г)",
                mealType: .snack,
                macros: MacroNutrients(protein: 6, fat: 18, carbs: 5),
                time: calendar.date(bySettingHour: 16, minute: 5, second: 0, of: now) ?? now
            )
        ]
    }

    func saveMeal(_ meal: Meal) async throws {
        try? await Task.sleep(nanoseconds: 200_000_000)
        // Mock: в продакшене — POST /nutrition/meals
    }

    func addMealByText(description: String, mealType: MealType) async throws -> Meal {
        try? await Task.sleep(nanoseconds: 700_000_000)
        // Mock: в продакшене — отправить описание на сервер для AI-анализа
        return Meal(
            name: description,
            mealType: mealType,
            macros: MacroNutrients(protein: 20, fat: 10, carbs: 30),
            time: Date()
        )
    }

    func analyzeMealFromPhoto(_ imageData: Data, mealType: MealType) async throws -> Meal {
        try? await Task.sleep(nanoseconds: 1_000_000_000)
        // Mock: в продакшене — загрузить фото на сервер, получить КБЖУ от AI
        return Meal(
            name: "Распознанное блюдо",
            mealType: mealType,
            macros: MacroNutrients(protein: 25, fat: 12, carbs: 35),
            time: Date()
        )
    }

    func generateRecipes() async throws -> [Recipe] {
        try? await Task.sleep(nanoseconds: 800_000_000)
        // Mock: в продакшене — отправить текущее КБЖУ и цель, получить рекомендации
        return [
            Recipe(
                name: "Тунец с авокадо",
                description: "Лёгкий ужин богатый белком и полезными жирами. Смешайте консервированный тунец с нарезанным авокадо, добавьте лимонный сок и специи.",
                macros: MacroNutrients(protein: 32, fat: 14, carbs: 4),
                preparationTime: 5
            ),
            Recipe(
                name: "Творог с ягодами",
                description: "Идеальный перекус. Обезжиренный творог с горстью свежих или замороженных ягод. Можно добавить немного мёда.",
                macros: MacroNutrients(protein: 18, fat: 2, carbs: 20),
                preparationTime: 3
            ),
            Recipe(
                name: "Куриный суп с овощами",
                description: "Питательный обед. Куриная грудка, брокколи, морковь и зелёный горошек в лёгком бульоне с зеленью.",
                macros: MacroNutrients(protein: 28, fat: 6, carbs: 18),
                preparationTime: 20
            )
        ]
    }
}
