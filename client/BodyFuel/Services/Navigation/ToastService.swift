import Foundation
import Combine

@MainActor
final class ToastService: ObservableObject {
    static let shared = ToastService()

    @Published var toast: String?

    private var toastTask: Task<Void, Never>?

    private init() {}

    func show(_ message: String) {
        toastTask?.cancel()
        toast = message
        toastTask = Task { @MainActor [weak self] in
            try? await Task.sleep(for: .seconds(3))
            guard !(Task.isCancelled) else { return }
            self?.toast = nil
        }
    }

    func dismiss() {
        toastTask?.cancel()
        toast = nil
    }
}
