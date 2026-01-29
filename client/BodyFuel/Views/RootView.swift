import SwiftUI

struct RootView: View {
    @EnvironmentObject var router: AppRouter

    var body: some View {
        switch router.currentFlow {
        case .auth:
            AuthView()

        case .onboarding:
            UserParametersView()

        case .main:
            MainScreenView()
        }
    }
}
