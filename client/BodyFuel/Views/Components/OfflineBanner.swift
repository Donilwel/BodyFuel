import SwiftUI

struct OfflineBanner: View {
    @ObservedObject private var monitor = NetworkMonitor.shared
    @ObservedObject private var syncManager = OfflineSyncManager.shared
    @ObservedObject private var mutationQueue = MutationQueue.shared

    private var hasPending: Bool { mutationQueue.pendingCount > 0 }
    private var shouldShow: Bool { !monitor.isOnline || syncManager.isSyncing || hasPending }

    var body: some View {
        if shouldShow {
            bannerView
                .transition(.move(edge: .top).combined(with: .opacity))
                .animation(.spring(duration: 0.35), value: shouldShow)
        }
    }

    private var bannerView: some View {
        HStack(spacing: 8) {
            if syncManager.isSyncing {
                ProgressView()
                    .tint(.white)
                    .scaleEffect(0.85)
                Text("Синхронизация...")
                    .font(.caption.weight(.medium))
                    .foregroundStyle(.white)
            } else if !monitor.isOnline {
                Image(systemName: "wifi.slash")
                    .font(.caption.weight(.semibold))
                    .foregroundStyle(.white)
                if hasPending {
                    Text("Офлайн · \(mutationQueue.pendingCount) \(pendingLabel) ожидают отправки")
                        .font(.caption.weight(.medium))
                        .foregroundStyle(.white)
                } else {
                    Text("Офлайн-режим")
                        .font(.caption.weight(.medium))
                        .foregroundStyle(.white)
                }
            } else if hasPending {
                Image(systemName: "arrow.triangle.2.circlepath")
                    .font(.caption.weight(.semibold))
                    .foregroundStyle(.white)
                Text("\(mutationQueue.pendingCount) \(pendingLabel) ожидают отправки")
                    .font(.caption.weight(.medium))
                    .foregroundStyle(.white)
            }

            Spacer()
        }
        .padding(.bottom, 4)
        .padding(.horizontal, 16)
        .background(bannerColor)
    }

    private var bannerColor: Color {
        if syncManager.isSyncing { return .indigo }
        if !monitor.isOnline { return Color.black.opacity(0.75) }
        return Color.orange.opacity(0.85)
    }

    private var pendingLabel: String {
        let count = mutationQueue.pendingCount
        switch count % 10 {
        case 1 where count % 100 != 11: return "изменение"
        case 2...4 where !(11...14).contains(count % 100): return "изменения"
        default: return "изменений"
        }
    }
}
