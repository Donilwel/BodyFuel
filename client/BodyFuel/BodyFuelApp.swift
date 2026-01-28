import SwiftUI

@main
struct BodyFuelApp: App {
    init() {
        Task {
            try? await HealthKitService.shared.requestAuthorization()
        }
    }
    
    var body: some Scene {
        WindowGroup {
//            UserParametersView()
            AuthView()
                .preferredColorScheme(.light)
        }
    }
}
