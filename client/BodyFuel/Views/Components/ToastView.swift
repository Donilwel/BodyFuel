import SwiftUI

struct ToastView: View {
    @ObservedObject private var toastService = ToastService.shared

    var body: some View {
        if let message = toastService.toast {
            Text(message)
                .font(.subheadline.weight(.medium))
                .foregroundStyle(.white)
                .multilineTextAlignment(.center)
                .padding(.horizontal, 18)
                .padding(.vertical, 11)
                .background(.ultraThinMaterial)
                .clipShape(Capsule())
                .padding(.bottom, 96)
                .transition(.move(edge: .bottom).combined(with: .opacity))
                .animation(.spring(duration: 0.35), value: message)
                .allowsHitTesting(false)
        }
    }
}
