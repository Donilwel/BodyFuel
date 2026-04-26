import SwiftUI

private struct AppAlertModifier: ViewModifier {
    @ObservedObject private var service = AlertService.shared

    func body(content: Content) -> some View {
        content
            .alert(
                "Ошибка",
                isPresented: Binding(
                    get: { service.current != nil },
                    set: { if !$0 { service.current = nil } }
                ),
                presenting: service.current
            ) { _ in
                Button("OK") { service.current = nil }
            } message: { item in
                Text(item.message)
            }
    }
}

extension View {
    func appAlert() -> some View {
        modifier(AppAlertModifier())
    }

    func cardStyle() -> some View {
        self
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding()
            .background(.ultraThinMaterial)
            .cornerRadius(24)
    }

    func screenLoading(_ isLoading: Bool) -> some View {
        self.overlay {
            if isLoading {
                ZStack {
                    Color.clear
                        .background(.black.opacity(0.1))
                        .ignoresSafeArea()
                    ProgressView()
                        .tint(.white)
                        .scaleEffect(1.5)
                }
                .transition(.opacity.animation(.easeOut(duration: 0.25)))
            }
        }
    }
}
