import SwiftUI
import Foundation

struct RootView: View {
    @EnvironmentObject var router: AppRouter

    var body: some View {
        switch router.rootRoute {
        case .auth:
            AuthView()

        case .parametersSetup:
            UserParametersView()

        case .main:
            TabBarView()
        }
    }
}
