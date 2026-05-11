import Foundation
@testable import BodyFuel

extension UserProfile {
    static func stub(
        height: Int = 170,
        photo: String = "",
        goal: MainGoal = .maintain,
        lifestyle: Lifestyle = .active,
        fitnessLevel: FitnessLevel = .intermediate,
        currentWeight: Double = 70,
        targetWeight: Double = 65,
        targetCaloriesDaily: Int = 2000,
        targetWorkoutsWeekly: Int = 3
    ) -> UserProfile {
        UserProfile(
            height: height,
            photo: photo,
            goal: goal,
            lifestyle: lifestyle,
            fitnessLevel: fitnessLevel,
            currentWeight: currentWeight,
            targetWeight: targetWeight,
            targetCaloriesDaily: targetCaloriesDaily,
            targetWorkoutsWeekly: targetWorkoutsWeekly
        )
    }
}

extension NutritionReportResponse {
    static func stub(
        from: String = "2026-01-01",
        to: String = "2026-01-07",
        days: Int = 7,
        entries: [FoodEntryResponseBody] = [],
        totalCalories: Double = 14000,
        totalProtein: Double = 560,
        totalCarbs: Double = 1050,
        totalFat: Double = 350,
        avgCaloriesPerDay: Double = 2000
    ) -> NutritionReportResponse {
        NutritionReportResponse(
            from: from,
            to: to,
            days: days,
            entries: entries,
            totalCalories: totalCalories,
            totalProtein: totalProtein,
            totalCarbs: totalCarbs,
            totalFat: totalFat,
            avgCaloriesPerDay: avgCaloriesPerDay
        )
    }
}
