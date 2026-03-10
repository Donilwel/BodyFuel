import WidgetKit

struct WorkoutProvider: TimelineProvider {
    func placeholder(in context: Context) -> WorkoutEntry {
        let workoutModel = WorkoutWidgetModel(
            name: "Full Body",
            duration: 45,
            calories: 320,
            location: "Gym",
            type: "Strength"
        )
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
            policy: .after(Date().addingTimeInterval(3600))
        )

        completion(timeline)
    }
}
