import Combine
import SwiftUI
import Foundation

final class AppRouter: ObservableObject {
    static let shared = AppRouter()

    private var workoutViewModel: WorkoutViewModel?

    @Published var selectedTab: TabRoute = .home
    @Published var rootRoute: RootRoute = .auth
    @Published var currentUser: User?
    @Published var pendingAddMeal = false

    private let sessionManager = UserSessionManager.shared

    private init() {
        updateRoute()
        loadCurrentUser()
    }

    func configure(workoutViewModel: WorkoutViewModel) {
        self.workoutViewModel = workoutViewModel
    }

    func logout() {
        ToastService.shared.dismiss()
        sessionManager.logout()
        SharedWidgetStorage.shared.clearAll()
        DiskCache.shared.removeAll()
        NutritionStore.shared.reset()
        UserStore.shared.reset()
        StatsStore.shared.reset()
        MutationQueue.shared.clear()
        selectedTab = .home
        pendingAddMeal = false
        rootRoute = .auth
    }

    func handleIfUnauthorized(_ error: Error) -> Bool {
        guard ErrorMapper.map(error) == .unauthorized else { return false }
        logout()
        return true
    }

    func updateRoute() {
        if sessionManager.currentUserId == nil {
            rootRoute = .auth

        } else if !sessionManager.hasCompletedParametersSetup {
            rootRoute = .parametersSetup

        } else {
            MutationQueue.shared.reload()
            rootRoute = .main
        }
    }
    
    func loadCurrentUser() {
        guard let userId = sessionManager.currentUserId else {
            currentUser = nil
            return
        }
    }
    
    func handleDeepLink(_ url: URL) {
        guard let deepLink = DeepLink(url: url),
              let workoutViewModel else { return }

        if sessionManager.currentUserId == nil {
            rootRoute = .auth
            return
        }

        if !sessionManager.hasCompletedParametersSetup {
            rootRoute = .parametersSetup
            return
        }

        rootRoute = .main

        switch deepLink {
        case .workouts:
            selectedTab = .home
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.15) { [weak workoutViewModel] in
                workoutViewModel?.shouldStartFromDeepLink = true
            }

        case .food:
            selectedTab = .food

        case .calories:
            selectedTab = .home
        }
    }
}
