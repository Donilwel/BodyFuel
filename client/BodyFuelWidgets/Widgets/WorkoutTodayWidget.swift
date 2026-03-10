import WidgetKit
import SwiftUI

struct WorkoutTodayWidget: Widget {
    let kind = "WorkoutTodayWidget"

    var body: some WidgetConfiguration {
        StaticConfiguration(
            kind: kind,
            provider: WorkoutProvider()
        ) { entry in
            WorkoutWidgetView(entry: entry)
        }
        .configurationDisplayName("Тренировка сегодня")
        .description("Показывает план тренировки на сегодня.")
        .supportedFamilies([
            .systemSmall,
            .systemMedium
        ])
    }
}
