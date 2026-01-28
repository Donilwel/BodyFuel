import SwiftUI

@main
struct BodyFuelApp: App {
    @StateObject private var router = AppRouter()
    
    init() {
        Task {
            try? await HealthKitService.shared.requestAuthorization()
        }
    }
    
    var body: some Scene {
        WindowGroup {
//            UserParametersView()
            RootView()
                .environmentObject(router)
                .preferredColorScheme(.light)
        }
    }
}
