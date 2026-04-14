import Foundation

struct MacroNutrients: Equatable {
    let protein: Double
    let fat: Double
    let carbs: Double

    var calories: Int {
        Int(protein * 4 + fat * 9 + carbs * 4)
    }

    static let zero = MacroNutrients(protein: 0, fat: 0, carbs: 0)
}

struct NutritionDailySummary {
    let consumed: MacroNutrients
    let goal: MacroNutrients
    let burned: Int

    var remainingCalories: Int {
        goal.calories - consumed.calories + burned
    }

    var proteinProgress: Double {
        guard goal.protein > 0 else { return 0 }
        return min(consumed.protein / goal.protein, 1)
    }

    var fatProgress: Double {
        guard goal.fat > 0 else { return 0 }
        return min(consumed.fat / goal.fat, 1)
    }

    var carbsProgress: Double {
        guard goal.carbs > 0 else { return 0 }
        return min(consumed.carbs / goal.carbs, 1)
    }

    var totalMacrosCalories: Int {
        consumed.calories
    }
}
