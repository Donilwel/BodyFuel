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
        defaults?.object(forKey: "todayBurnedCalories") as? Int
    }

    func saveTodayConsumedCalories(_ calories: Int) {
        defaults?.set(calories, forKey: "todayConsumedCalories")
    }

    func getTodayConsumedCalories() -> Int? {
        defaults?.object(forKey: "todayConsumedCalories") as? Int
    }

    func saveTargetCalories(_ calories: Int) {
        defaults?.set(calories, forKey: "targetCalories")
    }

    func getTargetCalories() -> Int? {
        guard let v = defaults?.object(forKey: "targetCalories") as? Int, v > 0 else { return nil }
        return v
    }

    func saveBasalMetabolicRate(_ bmr: Int) {
        defaults?.set(bmr, forKey: "basalMetabolicRate")
    }

    func getBasalMetabolicRate() -> Int? {
        guard let v = defaults?.object(forKey: "basalMetabolicRate") as? Int, v > 0 else { return nil }
        return v
    }

    func saveTodaySteps(_ steps: Int) {
        defaults?.set(steps, forKey: "todaySteps")
    }

    func getTodaySteps() -> Int? {
        defaults?.object(forKey: "todaySteps") as? Int
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

    func clearAll() {
        defaults?.removeObject(forKey: "todayBurnedCalories")
        defaults?.removeObject(forKey: "todayConsumedCalories")
        defaults?.removeObject(forKey: "targetCalories")
        defaults?.removeObject(forKey: "basalMetabolicRate")
        defaults?.removeObject(forKey: "todaySteps")
        defaults?.removeObject(forKey: "todayWorkout")
    }
}
