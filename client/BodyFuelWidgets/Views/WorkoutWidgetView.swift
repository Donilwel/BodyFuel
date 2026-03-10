import SwiftUI
import WidgetKit

struct WorkoutWidgetView: View {
    @Environment(\.widgetRenderingMode)
    var renderingMode
    
    let entry: WorkoutEntry

    var body: some View {
        Group {
            if let workout = entry.workout {
                VStack(alignment: .leading, spacing: 12) {
                    HStack {
                        Image(systemName: "figure.strengthtraining.traditional")
                        Text("Тренировка сегодня")
                    }
                    .font(.headline)
                    .foregroundColor(.white)
                    .widgetAccentable()
                    .symbolRenderingMode(.hierarchical)

                    Text(workout.name)
                        .font(.title3.bold())
                        .foregroundColor(.white)
                        .widgetAccentable()
                        .symbolRenderingMode(.hierarchical)

                    Text("\(workout.duration) мин • ~\(workout.calories) ккал")
                        .foregroundColor(.white.opacity(0.7))
                        .widgetAccentable()
                        .symbolRenderingMode(.hierarchical)
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .leading)
            }
        }
        .containerBackground(for: .widget) {
            Color.widgetBackground
        }
        .widgetURL(URL(string: "bodyfuel://workouts"))
    }
}
