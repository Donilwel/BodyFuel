import SwiftUI

struct MacroSummaryCard: View {
    let summary: NutritionDailySummary

    var body: some View {
        VStack(spacing: 16) {
            HStack(spacing: 20) {
                ZStack {
                    MacroRingView(summary: summary)
                        .frame(width: 140, height: 140)

                    VStack(spacing: 2) {
                        Text("\(summary.remainingCalories)")
                            .font(.title2.bold())
                            .foregroundColor(.white)
                        Text("ккал осталось")
                            .font(.caption)
                            .foregroundColor(.white.opacity(0.7))
                            .multilineTextAlignment(.center)
                    }
                }

                VStack(alignment: .leading, spacing: 10) {
                    MacroLegendRow(label: "Белки", value: summary.consumed.protein, goal: summary.goal.protein, color: .blue)
                    MacroLegendRow(label: "Жиры", value: summary.consumed.fat, goal: summary.goal.fat, color: .purple)
                    MacroLegendRow(label: "Углеводы", value: summary.consumed.carbs, goal: summary.goal.carbs, color: .indigo)
                }

                Spacer()
            }

            HStack {
                Spacer()
                VStack {
                    Text("\(summary.consumed.calories)")
                        .font(.subheadline.bold())
                        .foregroundColor(.white)
                    Text("потреблено")
                        .font(.caption)
                        .foregroundColor(.white.opacity(0.7))
                }
                Spacer()
                VStack {
                    Text("\(summary.goal.calories)")
                        .font(.subheadline.bold())
                        .foregroundColor(.white)
                    Text("цель")
                        .font(.caption)
                        .foregroundColor(.white.opacity(0.7))
                }
                Spacer()
            }
        }
        .padding()
        .background(.ultraThinMaterial)
        .cornerRadius(24)
    }
}

struct MacroLegendRow: View {
    let label: String
    let value: Double
    let goal: Double
    let color: Color

    var body: some View {
        HStack(spacing: 8) {
            Circle()
                .fill(color)
                .frame(width: 8, height: 8)
            VStack(alignment: .leading, spacing: 2) {
                Text(label)
                    .font(.caption)
                    .foregroundColor(.white.opacity(0.7))
                Text("\(Int(value)) / \(Int(goal)) г")
                    .font(.caption.bold())
                    .foregroundColor(.white)
            }
        }
    }
}
