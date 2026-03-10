import SwiftUI

@main
struct BodyFuelApp: App {
    @StateObject private var router = AppRouter.shared
    
    init() {
        Task {
            try? await HealthKitService.shared.requestAuthorization()
        }
    }
    
    var body: some Scene {
        WindowGroup {
            RootView()
                .environmentObject(router)
                .preferredColorScheme(.light)
                .onOpenURL { url in
                    router.handleDeepLink(url)
                }
        }
    }
}
