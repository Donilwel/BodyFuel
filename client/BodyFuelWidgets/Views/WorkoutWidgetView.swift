import SwiftUI
import WidgetKit

struct WorkoutWidgetView: View {
    @Environment(\.widgetRenderingMode)
    var renderingMode
    
    @Environment(\.widgetFamily)
    var widgetFamily
    
    let entry: WorkoutEntry

    var body: some View {
        Group {
            if let workout = entry.workout {
                VStack(alignment: .leading, spacing: 12) {
                    if widgetFamily != .systemSmall {
                        HStack {
                            Image(systemName: "figure.strengthtraining.traditional")
                            Text("Тренировка сегодня")
                        }
                        .font(.headline)
                        .foregroundColor(.white)
                        .widgetAccentable()
                        .symbolRenderingMode(.hierarchical)
                    }

                    Text(workout.name)
                        .font(.title3.bold())
                        .foregroundColor(.white)
                        .widgetAccentable()
                        .symbolRenderingMode(.hierarchical)

                    durationAndCalories(duration: workout.duration, calories: workout.calories)
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .leading)
            } else {
                HStack {
                    Image(systemName: "cloud.drizzle.fill")
                    Text("Нет информации о тренировках")
                }
                .symbolRenderingMode(.hierarchical)
            }
        }
        .containerBackground(for: .widget) {
            AppColors.gradient.opacity(0.9)
        }
        .widgetURL(URL(string: "bodyfuel://workouts"))
    }
    
    private func durationAndCalories(duration: Int, calories: Int) -> some View {
        Group {
            if widgetFamily == .systemSmall {
                VStack(alignment: .leading, spacing: 6) {
                    Text("\(duration) мин")
                        .foregroundColor(.white.opacity(0.7))
                        .widgetAccentable()
                        .symbolRenderingMode(.hierarchical)
                    
                    Text("~\(calories) ккал")
                        .foregroundColor(.white.opacity(0.7))
                        .widgetAccentable()
                        .symbolRenderingMode(.hierarchical)
                }
            } else {
                Text("\(duration) мин • ~\(calories) ккал")
                    .foregroundColor(.white.opacity(0.7))
                    .widgetAccentable()
                    .symbolRenderingMode(.hierarchical)
            }
        }
    }
}
