import Foundation

struct FoodProduct: Identifiable {
    let id: UUID
    let name: String
    let brand: String?
    let per100g: MacroNutrients
    let code: String?

    init(id: UUID = UUID(), name: String, brand: String?, per100g: MacroNutrients, code: String? = nil) {
        self.id = id
        self.name = name
        self.brand = brand
        self.per100g = per100g
        self.code = code
    }

    func macrosFor(grams: Double) -> MacroNutrients {
        MacroNutrients(
            protein: per100g.protein * grams / 100,
            fat: per100g.fat * grams / 100,
            carbs: per100g.carbs * grams / 100
        )
    }
}
