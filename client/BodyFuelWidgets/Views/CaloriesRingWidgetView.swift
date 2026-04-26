import SwiftUI
import WidgetKit

struct CaloriesRingWidgetView: View {
    @Environment(\.widgetRenderingMode)
    var renderingMode
    
    @Environment(\.widgetFamily)
    var widgetFamily
    
    let entry: CaloriesRingEntry
    
    var body: some View {
        VStack {
            if let calories = entry.calories {
                if widgetFamily == .systemMedium {
                    CaloriesRingProgressView(
                        consumed: calories.consumedToday,
                        goal: calories.target,
                        burned: calories.burnedToday,
                        basalMetabolicRate: calories.basalMetabolicRate
                    )
                } else if widgetFamily == .systemSmall {
                    VStack {
                        CaloriesDiagramView(
                            consumed: calories.consumedToday,
                            burned: calories.burnedToday,
                            basalMetabolicRate: calories.basalMetabolicRate
                        )
                        
                        Text("Цель - \(calories.target.description) ккал")
                            .multilineTextAlignment(.center)
                            .font(.subheadline)
                            .foregroundColor(.white.opacity(0.8))
                            .widgetAccentable()
                            .symbolRenderingMode(.hierarchical)
                    }
                }
            } else {
                HStack {
                    Image(systemName: "cloud.drizzle.fill")
                    Text("Нет информации о калориях")
                }
                .widgetAccentable()
                .symbolRenderingMode(.hierarchical)
            }
        }
        .containerBackground(for: .widget) {
            AppColors.widgetBackground
        }
        .widgetURL(URL(string: "bodyfuel://calories"))
    }
}
