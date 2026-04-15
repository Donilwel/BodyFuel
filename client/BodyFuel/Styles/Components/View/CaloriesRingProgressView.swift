import SwiftUI

struct CaloriesRingProgressView: View {
    let consumed: Int
    let goal: Int
    let burned: Int
    let basalMetabolicRate: Int

    @State private var showInfo = false

    private var total: Int {
        basalMetabolicRate + burned
    }

    private var progress: Double {
        Double(consumed) / Double(total)
    }

    var body: some View {
        ZStack(alignment: .topTrailing) {
            mainContent

            Button {
                withAnimation(.easeInOut(duration: 0.2)) {
                    showInfo.toggle()
                }
            } label: {
                Image(systemName: "exclamationmark.circle")
                    .font(.subheadline)
                    .foregroundColor(.white.opacity(0.55))
                    .padding(12)
            }

            if showInfo {
                infoOverlay
                    .transition(.opacity.combined(with: .scale(scale: 0.97, anchor: .topTrailing)))
            }
        }
        .padding()
        .background(.ultraThinMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 24))
    }

    private var mainContent: some View {
        VStack(alignment: .center, spacing: 12) {
            HStack {
                VStack {
                    Text("Потреблено")
                        .multilineTextAlignment(.center)
                        .font(.subheadline)
                        .foregroundColor(.white.opacity(0.8))

                    Text(consumed.description)
                        .font(.title3.bold())
                        .foregroundColor(.white.opacity(0.8))
                }
                .frame(width: 100, height: 100)

                Spacer()
                ZStack {
                    RingDiagramView(progress: progress)
                        .frame(width: 140, height: 140)

                    VStack {
                        Text("Осталось")
                            .font(.subheadline)
                            .opacity(0.7)
                        Text("\(total - consumed) ккал")
                            .font(.title3.bold())
                            .lineLimit(1)
                        Text("из \(total)")
                            .font(.subheadline)
                            .opacity(0.7)
                            .lineLimit(1)
                    }
                    .foregroundColor(.white)
                }

                Spacer()

                VStack {
                    Text("Сожжено")
                        .multilineTextAlignment(.center)
                        .font(.subheadline)
                        .foregroundColor(.white.opacity(0.8))

                    Text(burned.description)
                        .font(.title3.bold())
                        .foregroundColor(.white.opacity(0.8))
                }
                .frame(width: 100, height: 100)
            }

            Text("Цель — \(goal.description) ккал")
                .multilineTextAlignment(.center)
                .font(.subheadline)
                .foregroundColor(.white.opacity(0.8))
        }
    }

    private var infoOverlay: some View {
        ZStack(alignment: .topTrailing) {
            RoundedRectangle(cornerRadius: 20)
                .fill(.ultraThinMaterial)

            VStack(alignment: .leading, spacing: 14) {
                Text("Как считается диаграмма")
                    .font(.headline)
                    .foregroundColor(.white)

                infoRow(
                    icon: "flame.fill",
                    color: .orange,
                    title: "Базовый обмен",
                    value: "\(basalMetabolicRate) ккал",
                    detail: "Калории, которые твое тело тратит в состоянии покоя"
                )

                infoRow(
                    icon: "figure.run",
                    color: .green,
                    title: "Активность",
                    value: "+ \(burned) ккал",
                    detail: "Сожжено на тренировках и движении"
                )

                Divider().background(.white.opacity(0.3))

                infoRow(
                    icon: "circle.dashed",
                    color: .white,
                    title: "Максимум кольца",
                    value: "\(total) ккал",
                    detail: "Базовый обмен + активность — столько калорий ты потратил сегодня"
                )

                infoRow(
                    icon: "target",
                    color: .blue.opacity(0.8),
                    title: "Цель",
                    value: "\(goal) ккал",
                    detail: "Норма, которую ты установил. Можно поменять в настройках"
                )
            }
            .padding(16)

            Button {
                withAnimation(.easeInOut(duration: 0.2)) {
                    showInfo = false
                }
            } label: {
                Image(systemName: "xmark")
                    .font(.caption.bold())
                    .foregroundColor(.white.opacity(0.6))
                    .padding(12)
            }
        }
    }

    private func infoRow(icon: String, color: Color, title: String, value: String, detail: String) -> some View {
        HStack(alignment: .top, spacing: 10) {
            Image(systemName: icon)
                .foregroundColor(color)
                .frame(width: 20)

            VStack(alignment: .leading, spacing: 2) {
                HStack {
                    Text(title)
                        .font(.subheadline.weight(.semibold))
                        .foregroundColor(.white)
                    Spacer()
                    Text(value)
                        .font(.subheadline.bold())
                        .foregroundColor(.white)
                }
                Text(detail)
                    .font(.caption)
                    .foregroundColor(.white.opacity(0.6))
                    .fixedSize(horizontal: false, vertical: true)
            }
        }
    }
}

#Preview {
    ZStack {
        AnimatedBackground()
        
        CaloriesRingProgressView(consumed: 500, goal: 1500, burned: 400, basalMetabolicRate: 1500)
    }
}
