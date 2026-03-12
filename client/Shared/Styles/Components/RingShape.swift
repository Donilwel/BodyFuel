import SwiftUI

struct RingShape: Shape {
    var startProgress: Double = 0
    var endProgress: Double = 1
    var startAngle: Double = -90
    
    static func progressToAngle(progress: Double, startAngle: Double) -> Double {
        (progress * 360) + startAngle
    }

    var animatableData: AnimatablePair<Double, Double> {
        get { AnimatablePair(startProgress, endProgress) }
        set {
            startProgress = newValue.first
            endProgress = newValue.second
        }
    }

    func path(in rect: CGRect) -> Path {
        let radius = min(rect.width, rect.height) / 2
        let center = CGPoint(x: rect.midX, y: rect.midY)

        let start = Angle(degrees: startAngle + startProgress * 360)
        let end = Angle(degrees: startAngle + endProgress * 360)

        var path = Path()
        path.addArc(
            center: center,
            radius: radius,
            startAngle: start,
            endAngle: end,
            clockwise: false
        )

        return path
    }
}
