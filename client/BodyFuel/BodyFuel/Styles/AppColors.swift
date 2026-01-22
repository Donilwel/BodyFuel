import SwiftUI

enum AppColors {
    static let backgroundGradient = LinearGradient(
        colors: [Color.blue.opacity(0.9), Color.indigo.opacity(0.9)],
        startPoint: .topTrailing,
        endPoint: .bottom
    )
    static let background =  Color.blue
    static let primary = Color.indigo
}

#Preview {
    AuthView()
}
