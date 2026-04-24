import Foundation

struct Meal: Identifiable, Codable {
    let id: UUID
    let name: String
    let mealType: MealType
    let macros: MacroNutrients
    let time: Date
    let photoURL: String?

    init(
        id: UUID = UUID(),
        name: String,
        mealType: MealType,
        macros: MacroNutrients,
        time: Date = Date(),
        photoURL: String? = nil
    ) {
        self.id = id
        self.name = name
        self.mealType = mealType
        self.macros = macros
        self.time = time
        self.photoURL = photoURL
    }
}

struct RecipeIngredient {
    let name: String
    let grams: Double
}

struct Recipe: Identifiable {
    let id: UUID
    let name: String
    let description: String
    let macros: MacroNutrients
    let ingredients: [RecipeIngredient]
    let preparationTime: Int

    init(
        id: UUID = UUID(),
        name: String,
        description: String,
        macros: MacroNutrients,
        ingredients: [RecipeIngredient] = [],
        preparationTime: Int
    ) {
        self.id = id
        self.name = name
        self.description = description
        self.macros = macros
        self.ingredients = ingredients
        self.preparationTime = preparationTime
    }
}
