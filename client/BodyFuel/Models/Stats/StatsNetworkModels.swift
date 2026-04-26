import Foundation

// MARK: - Weight

struct WeightEntryResponse: Codable, Identifiable {
    let id: String
    let weight: Double
    let date: String

    private enum CodingKeys: String, CodingKey {
        case id, weight, date
    }
}

struct AddWeightRequestBody: Encodable {
    let weight: Double
}

// MARK: - Calories History

struct CaloriesHistoryEntryResponse: Codable, Identifiable {
    let id: String
    let calories: Int
    let description: String?
    let date: String

    private enum CodingKeys: String, CodingKey {
        case id, calories, description, date
    }
}

// MARK: - Recommendations

struct RecommendationResponse: Codable, Identifiable {
    let id: String
    let type: String
    let description: String
    let priority: Int
    let isRead: Bool
    let generatedAt: String

    private enum CodingKeys: String, CodingKey {
        case id, type, description, priority
        case isRead = "is_read"
        case generatedAt = "generated_at"
    }
}

// MARK: - HealthKit helper

struct DailySteps: Identifiable {
    let id = UUID()
    let date: Date
    let count: Int
}

// MARK: - Chart

struct ChartDataPoint: Identifiable, Equatable {
    let id = UUID()
    let date: Date
    let value: Double
}

// MARK: - Meal Breakdown

enum MealTypeLabel: String, CaseIterable {
    case breakfast = "breakfast"
    case lunch     = "lunch"
    case dinner    = "dinner"
    case snack     = "snack"

    var localizedName: String {
        switch self {
        case .breakfast: return "Завтрак"
        case .lunch:     return "Обед"
        case .dinner:    return "Ужин"
        case .snack:     return "Перекусы"
        }
    }
}

struct MealBreakdownItem: Identifiable {
    let id = UUID()
    let mealType: MealTypeLabel
    let calories: Int
    let percent: Int
}
