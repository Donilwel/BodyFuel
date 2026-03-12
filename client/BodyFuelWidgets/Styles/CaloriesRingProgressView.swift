import SwiftUI
import WidgetKit

struct CaloriesRingProgressView: View {
    let consumed: Int
    let goal: Int
    let burned: Int
    let basalMetabolicRate: Int
    
    private var total: Int {
        basalMetabolicRate + burned
    }
    
    private var progress: Double {
        min(Double(consumed) / Double(total), 1)
    }
    
    var body: some View {
        VStack(alignment: .center, spacing: 12) {
            HStack {
                VStack {
                    Text("Потреблено")
                        .multilineTextAlignment(.center)
                        .font(.subheadline)
                        .foregroundColor(.white.opacity(0.8))
                        .widgetAccentable()
                        .symbolRenderingMode(.hierarchical)
                    
                    Text(consumed.description)
                        .font(.title3.bold())
                        .foregroundColor(.white.opacity(0.8))
                        .widgetAccentable()
                        .symbolRenderingMode(.hierarchical)
                }
                .frame(width: 100, height: 100)
                
                Spacer()
                
                CaloriesDiagramView(
                    consumed: consumed,
                    burned: burned,
                    basalMetabolicRate: basalMetabolicRate
                )
                .frame(width: 110, height: 110)
            
                Spacer()
                
                VStack {
                    Text("Сожжено")
                        .multilineTextAlignment(.center)
                        .font(.subheadline)
                        .foregroundColor(.white.opacity(0.8))
                        .widgetAccentable()
                        .symbolRenderingMode(.hierarchical)
                    
                    Text(burned.description)
                        .font(.title3.bold())
                        .foregroundColor(.white.opacity(0.8))
                        .widgetAccentable()
                        .symbolRenderingMode(.hierarchical)
                }
                .frame(width: 100, height: 100)
            }
            
            Text("Цель - \(goal.description) ккал")
                .multilineTextAlignment(.center)
                .lineLimit(1)
                .font(.subheadline)
                .foregroundColor(.white.opacity(0.8))
                .widgetAccentable()
                .symbolRenderingMode(.hierarchical)
        }
        .padding()
        .clipShape(RoundedRectangle(cornerRadius: 24))
    }
}
