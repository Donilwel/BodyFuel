import SwiftUI
import WidgetKit

struct CaloriesDiagramView: View {
    let consumed: Int
    let burned: Int
    let basalMetabolicRate: Int
    
    private var total: Int {
        basalMetabolicRate + burned
    }
    
    private var progress: Double {
        Double(consumed) / Double(total)
    }
    
    private let startAngle: Double = -90
    private let lineWidth: CGFloat = 8
    
    private let foregroundColors: [Color] = [.indigo, .blue]
    private var lastGradientColor: Color {
        foregroundColors.last ?? .black
    }
    
    private var gradientStartAngle: Double {
        progress >= 1 ? relativePercentageAngle - 360 : startAngle
    }
    
    private var ringGradient: AngularGradient {
        AngularGradient(
        gradient: Gradient(colors: foregroundColors),
        center: .center,
        startAngle: Angle(degrees: gradientStartAngle),
        endAngle: Angle(degrees: relativePercentageAngle)
        )
    }
    
    private var absolutePercentageAngle: Double {
        RingShape.progressToAngle(progress: progress, startAngle: 0)
    }
    private var relativePercentageAngle: Double {
        absolutePercentageAngle + startAngle
    }
    
    var body: some View {
        GeometryReader { geometry in
            let radius = (min(geometry.size.width, geometry.size.height)) / 2
            ZStack {
                RingShape()
                    .stroke(Color.white.opacity(0.15), lineWidth: lineWidth)
                
                RingShape(
                    endProgress: progress,
                    startAngle: -90
                )
                .stroke(ringGradient, style: StrokeStyle(lineWidth: lineWidth, lineCap: .round))
                
                Circle()
                    .fill(lastGradientColor)
                    .frame(width: lineWidth, height: lineWidth)
                    .offset(y: -radius)
                    .rotationEffect(.degrees(progress * 360))
                
                VStack {
                    Text("Осталось")
                        .font(.caption)
                        .opacity(0.7)
                    Text("\(total - consumed) ккал")
                        .font(.title3.bold())
                        .lineLimit(1)
                    Text("из \(total)")
                        .font(.caption)
                        .opacity(0.7)
                        .lineLimit(1)
                }
                .foregroundColor(.white)
            }
        }
    }
}

