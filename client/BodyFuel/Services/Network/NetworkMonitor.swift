import Network
import Combine

@MainActor
final class NetworkMonitor: ObservableObject {
    static let shared = NetworkMonitor()

    @Published private(set) var isOnline: Bool = true

    private let monitor = NWPathMonitor()
    private let monitorQueue = DispatchQueue(label: "com.bodyfuel.networkmonitor", qos: .utility)

    private var isPathSatisfied: Bool = true

    private var isServerReachable: Bool = true

    private init() {
        isPathSatisfied = monitor.currentPath.status == .satisfied
        isOnline = isPathSatisfied

        if isOnline {
            Task { await OfflineSyncManager.shared.flush() }
        }

        monitor.pathUpdateHandler = { [weak self] path in
            let satisfied = path.status == .satisfied
            Task { @MainActor [weak self] in
                guard let self, satisfied != self.isPathSatisfied else { return }
                self.isPathSatisfied = satisfied
                if satisfied {
                    self.isServerReachable = true
                }
                self.updateIsOnline()
                print("[INFO] [NetworkMonitor]: path=\(satisfied ? "up" : "down"), serverOk=\(self.isServerReachable)")
                if self.isOnline {
                    await OfflineSyncManager.shared.flush()
                }
            }
        }
        monitor.start(queue: monitorQueue)
    }

    // MARK: - Server reachability

    func markServerUnreachable() {
        guard isServerReachable else { return }
        isServerReachable = false
        updateIsOnline()
        print("[INFO] [NetworkMonitor]: Server unreachable — forcing offline mode")
    }

    func markServerReachable() {
        guard !isServerReachable else { return }
        isServerReachable = true
        updateIsOnline()
        print("[INFO] [NetworkMonitor]: Server reachable again")
        if isOnline {
            Task { await OfflineSyncManager.shared.flush() }
        }
    }

    // MARK: - Private

    private func updateIsOnline() {
        isOnline = isPathSatisfied && isServerReachable
    }
}
