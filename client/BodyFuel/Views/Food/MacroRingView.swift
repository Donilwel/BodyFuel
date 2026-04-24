import SwiftUI

struct MacroRingView: View {
    let summary: NutritionDailySummary

    private let ringWidth: CGFloat = 12

    var body: some View {
        GeometryReader { geometry in
            let size = min(geometry.size.width, geometry.size.height)
            let radius = size / 2

            ZStack {
                Circle()
                    .stroke(Color.white.opacity(0.15), lineWidth: ringWidth)

                MacroArcShape(startFraction: 0, endFraction: summary.proteinProgress * 0.333)
                    .stroke(Color.blue, style: StrokeStyle(lineWidth: ringWidth, lineCap: .round))

                MacroArcShape(startFraction: 0.333, endFraction: 0.333 + summary.fatProgress * 0.333)
                    .stroke(Color.purple, style: StrokeStyle(lineWidth: ringWidth, lineCap: .round))

                MacroArcShape(startFraction: 0.667, endFraction: 0.667 + summary.carbsProgress * 0.333)
                    .stroke(Color.indigo, style: StrokeStyle(lineWidth: ringWidth, lineCap: .round))
            }
            .padding(ringWidth / 2)
        }
    }
}

struct MacroArcShape: Shape {
    var startFraction: Double
    var endFraction: Double

    func path(in rect: CGRect) -> Path {
        var path = Path()
        let center = CGPoint(x: rect.midX, y: rect.midY)
        let radius = min(rect.width, rect.height) / 2
        let startAngle = Angle(degrees: -90 + startFraction * 360)
        let endAngle = Angle(degrees: -90 + endFraction * 360)
        path.addArc(center: center, radius: radius, startAngle: startAngle, endAngle: endAngle, clockwise: false)
        return path
    }
}
