import Foundation
import Combine

@MainActor
final class AlertService: ObservableObject {
    static let shared = AlertService()

    struct AlertItem: Identifiable {
        let id = UUID()
        let message: String
    }

    @Published var current: AlertItem?

    private init() {}

    func show(_ message: String) {
        current = AlertItem(message: message)
    }
}
