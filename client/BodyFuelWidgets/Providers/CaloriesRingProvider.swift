import WidgetKit

struct CaloriesRingProvider: TimelineProvider {
    private let sharedWidgetStorage = SharedWidgetStorage.shared
    
    func placeholder(in context: Context) -> CaloriesRingEntry {
        guard let target = sharedWidgetStorage.getTargetCalories(),
              let consumedToday = sharedWidgetStorage.getTodayConsumedCalories(),
              let burnedToday = sharedWidgetStorage.getTodayBurnedCalories(),
              let basalMetabolicRate = sharedWidgetStorage.getBasalMetabolicRate() else {
            return CaloriesRingEntry(
                date: .now,
                calories: nil
            )
        }
        
        let caloriesModel = CaloriesModel(
            target: target,
            consumedToday: consumedToday,
            burnedToday: burnedToday,
            basalMetabolicRate: basalMetabolicRate
        )
        return CaloriesRingEntry(
            date: .now,
            calories: caloriesModel
        )
    }

    func getSnapshot(
        in context: Context,
        completion: @escaping (CaloriesRingEntry) -> Void
    ) {
        let calories = placeholder(in: context)

        completion(placeholder(in: context))
    }

    func getTimeline(
        in context: Context,
        completion: @escaping (Timeline<CaloriesRingEntry>) -> Void
    ) {
        let entry = placeholder(in: context)

        let timeline = Timeline(
            entries: [entry],
            policy: .after(Date().addingTimeInterval(1800))
        )

        completion(timeline)
    }
}
