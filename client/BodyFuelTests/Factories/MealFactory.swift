import Foundation
@testable import BodyFuel

extension MacroNutrients {
    static func stub(
        protein: Double = 30,
        fat: Double = 10,
        carbs: Double = 50
    ) -> MacroNutrients {
        MacroNutrients(protein: protein, fat: fat, carbs: carbs)
    }
}

extension Meal {
    static func stub(
        id: UUID = UUID(),
        name: String = "Куриная грудка с рисом",
        mealType: MealType = .lunch,
        macros: MacroNutrients = .stub(),
        time: Date = Date(),
        photoURL: String? = nil
    ) -> Meal {
        Meal(
            id: id,
            name: name,
            mealType: mealType,
            macros: macros,
            time: time,
            photoURL: photoURL
        )
    }
}

extension Recipe {
    static func stub(
        id: UUID = UUID(),
        name: String = "Куриный суп",
        description: String = "Лёгкий и питательный суп",
        macros: MacroNutrients = .stub(),
        ingredients: [RecipeIngredient] = [
            RecipeIngredient(name: "Курица", grams: 200),
            RecipeIngredient(name: "Морковь", grams: 50)
        ],
        preparationTime: Int = 30
    ) -> Recipe {
        Recipe(
            id: id,
            name: name,
            description: description,
            macros: macros,
            ingredients: ingredients,
            preparationTime: preparationTime
        )
    }
}

extension NutritionDailySummary {
    static func stub(
        consumed: MacroNutrients = .stub(protein: 80, fat: 30, carbs: 150),
        goal: MacroNutrients = .stub(protein: 120, fat: 50, carbs: 200),
        burned: Int = 300
    ) -> NutritionDailySummary {
        NutritionDailySummary(consumed: consumed, goal: goal, burned: burned)
    }
}
