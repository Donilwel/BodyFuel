import SwiftUI

struct RingDiagramView: View {
    let progress: Double
    
    @State private var head: Double = 0
    @State private var tail: Double = 0
    
    private let ringWidth: CGFloat = 10
    private let startAngle: Double = -90
    
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
        RingShape.progressToAngle(progress: head, startAngle: 0)
    }
    private var relativePercentageAngle: Double {
        absolutePercentageAngle + startAngle
    }
    
    var body: some View {
        GeometryReader { geometry in
            let radius = (min(geometry.size.width, geometry.size.height)) / 2
            ZStack {
                RingShape()
                    .stroke(style: StrokeStyle(lineWidth: ringWidth, lineCap: .round))
                    .fill(.white.opacity(0.15))
                
                RingShape(
                    startProgress: tail,
                    endProgress: head,
                    startAngle: startAngle
                )
                .stroke(ringGradient, style: StrokeStyle(lineWidth: ringWidth, lineCap: .round, lineJoin: .round))

                Circle()
                    .fill(lastGradientColor)
                    .frame(width: ringWidth, height: ringWidth)
                    .offset(y: -radius)
                    .rotationEffect(.degrees(head * 360))
            }
        }
        .padding(ringWidth / 2)
        .onAppear {
            animateSnake(to: progress)
        }
        .onChange(of: progress) { newValue in
            animateSnake(to: newValue)
        }
    }
    
    private func animateSnake(to value: Double) {
        let stepDuration = 1.0 * ceil(value)
        
        if value < 1 {
            withAnimation(.easeOut(duration: stepDuration)) {
                head = min(value, 1)
            }
            return
        }

        for i in 1...Int(ceil(value)) {
            withAnimation(.easeOut(duration: stepDuration)) {
                head = min(value, Double(i))
            }

            withAnimation(.easeOut(duration: stepDuration)) {
                tail = max(0, min(value - Double(i), 0))
            }
        }
    }
}
