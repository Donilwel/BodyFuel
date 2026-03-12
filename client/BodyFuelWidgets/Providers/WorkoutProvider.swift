import WidgetKit

struct WorkoutProvider: TimelineProvider {
    private let sharedWidgetStorage = SharedWidgetStorage.shared
    
    func placeholder(in context: Context) -> WorkoutEntry {
        let workoutModel = sharedWidgetStorage.getWorkout()
        return WorkoutEntry(
            date: .now,
            workout: workoutModel
        )
    }

    func getSnapshot(
        in context: Context,
        completion: @escaping (WorkoutEntry) -> Void
    ) {
        completion(placeholder(in: context))
    }

    func getTimeline(
        in context: Context,
        completion: @escaping (Timeline<WorkoutEntry>) -> Void
    ) {
        let entry = placeholder(in: context)

        let timeline = Timeline(
            entries: [entry],
            policy: .after(Date().addingTimeInterval(1800))
        )

        completion(timeline)
    }
}
