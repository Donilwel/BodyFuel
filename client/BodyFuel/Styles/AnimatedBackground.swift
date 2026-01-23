import SwiftUI

struct AnimatedBackground: View {
    @State private var animateGradient = false

    var body: some View {
        LinearGradient(
            colors: [
                animateGradient ? .indigo : .blue,
                animateGradient ? .blue : .indigo,
            ],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
        .hueRotation(.degrees(animateGradient ? 25 : 0))
        .ignoresSafeArea()
        .onAppear {
            withAnimation(.easeInOut(duration: 7.0).repeatForever(autoreverses: true)) {
                animateGradient.toggle()
            }
        }
    }
}

#Preview {
    AnimatedBackground()
}
