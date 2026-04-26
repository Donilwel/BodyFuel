import WidgetKit
import SwiftUI

struct CaloriesRingWidget: Widget {
    let kind = "CaloriesRingWidget"

    var body: some WidgetConfiguration {
        StaticConfiguration(
            kind: kind,
            provider: CaloriesRingProvider()
        ) { entry in
            CaloriesRingWidgetView(entry: entry)
        }
        .configurationDisplayName("Диаграмма плана калорий на сегодня")
        .description("Показывает круговую диаграмму потраченных и потребленных калорий за сегодня.")
        .supportedFamilies([
            .systemSmall,
            .systemMedium
        ])
    }
}
