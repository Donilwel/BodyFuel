import Foundation

final class SharedWidgetStorage {
    static let shared = SharedWidgetStorage()

    private let defaults = UserDefaults(
        suiteName: "group.com.bodyfuel.shared"
    )
    
    func saveTodayBurnedCalories(_ calories: Int) {
        defaults?.set(calories, forKey: "todayBurnedCalories")
    }

    func getTodayBurnedCalories() -> Int? {
        defaults?.integer(forKey: "todayBurnedCalories")
    }
    
    func saveTodayConsumedCalories(_ calories: Int) {
        defaults?.set(calories, forKey: "todayConsumedCalories")
    }
    
    func getTodayConsumedCalories() -> Int? {
        defaults?.integer(forKey: "todayConsumedCalories")
    }
    
    func saveTargetCalories(_ calories: Int) {
        defaults?.set(calories, forKey: "targetCalories")
    }
    
    func getTargetCalories() -> Int? {
        defaults?.integer(forKey: "targetCalories")
    }
    
    func saveBasalMetabolicRate(_ bmr: Int) {
        defaults?.set(bmr, forKey: "basalMetabolicRate")
    }
    
    func getBasalMetabolicRate() -> Int? {
        defaults?.integer(forKey: "basalMetabolicRate")
    }

    func saveWorkout(_ workout: WorkoutModel?) {
        if let workout {
            let data = try? JSONEncoder().encode(workout)
            defaults?.set(data, forKey: "todayWorkout")
        } else {
            defaults?.removeObject(forKey: "todayWorkout")
        }
    }

    func getWorkout() -> WorkoutModel? {
        guard let data = defaults?.data(forKey: "todayWorkout") else {
            return nil
        }

        return try? JSONDecoder().decode(
            WorkoutModel.self,
            from: data
        )
    }
}
