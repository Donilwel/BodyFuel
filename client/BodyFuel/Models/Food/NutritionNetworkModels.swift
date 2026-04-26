import Foundation

// MARK: - Requests

struct CreateFoodEntryRequestBody: Encodable {
    let description: String
    let calories: Int
    let protein: Double
    let carbs: Double
    let fat: Double
    let mealType: String
    let photoURL: String?
    let date: Date

    enum CodingKeys: String, CodingKey {
        case description, calories, protein, carbs, fat, date
        case mealType  = "meal_type"
        case photoURL  = "photo_url"
    }
}

struct UpdateFoodEntryRequestBody: Encodable {
    let description: String?
    let calories: Int?
    let protein: Double?
    let carbs: Double?
    let fat: Double?
    let mealType: String?

    enum CodingKeys: String, CodingKey {
        case description, calories, protein, carbs, fat
        case mealType = "meal_type"
    }

    func encode(to encoder: Encoder) throws {
        var c = encoder.container(keyedBy: CodingKeys.self)
        try c.encodeIfPresent(description, forKey: .description)
        try c.encodeIfPresent(calories,    forKey: .calories)
        try c.encodeIfPresent(protein,     forKey: .protein)
        try c.encodeIfPresent(carbs,       forKey: .carbs)
        try c.encodeIfPresent(fat,         forKey: .fat)
        try c.encodeIfPresent(mealType,    forKey: .mealType)
    }
}

// MARK: - Responses

struct FoodEntryResponseBody: Codable {
    let id: String
    let description: String
    let calories: Int
    let protein: Double
    let carbs: Double
    let fat: Double
    let mealType: String
    let photoURL: String?
    let date: String
    let createdAt: String

    enum CodingKeys: String, CodingKey {
        case id, description, calories, protein, carbs, fat, date
        case mealType  = "meal_type"
        case photoURL  = "photo_url"
        case createdAt = "created_at"
    }
}

struct NutritionDiaryResponseBody: Decodable {
    let date: String
    let entries: [FoodEntryResponseBody]
    let totalCalories: Int
    let totalProtein: Double
    let totalCarbs: Double
    let totalFat: Double

    enum CodingKeys: String, CodingKey {
        case date, entries
        case totalCalories = "total_calories"
        case totalProtein  = "total_protein"
        case totalCarbs    = "total_carbs"
        case totalFat      = "total_fat"
    }
}

struct RecipeMacrosResponseBody: Decodable {
    let protein: Double
    let fat: Double
    let carbs: Double
}

struct RecipeIngredientResponseBody: Decodable {
    let name: String
    let grams: Double
}

struct RecipeResponseBody: Decodable {
    let id: String
    let name: String
    let description: String
    let ingredients: [RecipeIngredientResponseBody]
    let macros: RecipeMacrosResponseBody
    let preparationTime: Int

    enum CodingKeys: String, CodingKey {
        case id, name, description, ingredients, macros
        case preparationTime = "preparation_time"
    }
}

// MARK: - Nutrition Report

struct NutritionReportResponse: Codable {
    let from: String
    let to: String
    let days: Int
    let entries: [FoodEntryResponseBody]
    let totalCalories: Double
    let totalProtein: Double
    let totalCarbs: Double
    let totalFat: Double
    let avgCaloriesPerDay: Double

    enum CodingKeys: String, CodingKey {
        case from, to, days, entries
        case totalCalories     = "total_calories"
        case totalProtein      = "total_protein"
        case totalCarbs        = "total_carbs"
        case totalFat          = "total_fat"
        case avgCaloriesPerDay = "avg_calories_per_day"
    }
}
