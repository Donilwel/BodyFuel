import SwiftUI

struct RootView: View {
    @EnvironmentObject var router: AppRouter

    var body: some View {
        switch router.currentFlow {
        case .auth:
            AuthView()

        case .profileSetup:
            UserParametersView()

        default:
            TabBarView()
        }
    }
}
