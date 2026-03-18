import Combine
import SwiftUI
import Foundation

final class AppRouter: ObservableObject {
    static let shared = AppRouter()
    
    private var workoutViewModel: WorkoutViewModel?
    
    @Published var selectedTab: TabRoute = .home
    @Published var rootRoute: RootRoute = .auth
    @Published var currentUser: User?
    
    private let sessionManager = UserSessionManager.shared
    
    private init() {
        updateRoute()
        loadCurrentUser()
    }
    
    func configure(workoutViewModel: WorkoutViewModel) {
        self.workoutViewModel = workoutViewModel
    }
    
    func updateRoute() {
        if sessionManager.currentUserId == nil {
            rootRoute = .auth

        } else if !sessionManager.hasCompletedParametersSetup {
            rootRoute = .parametersSetup

        } else {
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
            workoutViewModel.shouldStartFromDeepLink = true

        case .food:
            selectedTab = .food

        case .calories:
            selectedTab = .home
        }
    }
}
