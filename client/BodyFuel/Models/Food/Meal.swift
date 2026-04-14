import Foundation

struct Meal: Identifiable {
    let id: UUID
    let name: String
    let mealType: MealType
    let macros: MacroNutrients
    let time: Date

    init(
        id: UUID = UUID(),
        name: String,
        mealType: MealType,
        macros: MacroNutrients,
        time: Date = Date()
    ) {
        self.id = id
        self.name = name
        self.mealType = mealType
        self.macros = macros
        self.time = time
    }
}

struct Recipe: Identifiable {
    let id: UUID
    let name: String
    let description: String
    let macros: MacroNutrients
    let preparationTime: Int

    init(
        id: UUID = UUID(),
        name: String,
        description: String,
        macros: MacroNutrients,
        preparationTime: Int
    ) {
        self.id = id
        self.name = name
        self.description = description
        self.macros = macros
        self.preparationTime = preparationTime
    }
}
