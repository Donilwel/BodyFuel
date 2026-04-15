import SwiftUI

@main
struct BodyFuelApp: App {
    @StateObject private var router = AppRouter.shared
    @StateObject var workoutViewModel = WorkoutViewModel()
    
    init() {
        Task {
            await HealthKitService.shared.requestAuthorization()
        }
    }
    
    var body: some Scene {
        WindowGroup {
            RootView()
                .environmentObject(router)
                .environmentObject(workoutViewModel)
                .preferredColorScheme(.light)
                .onOpenURL { url in
                    router.handleDeepLink(url)
                }
                .onAppear {
                    router.configure(workoutViewModel: workoutViewModel)
                }
        }
    }
}

#Preview {
    RootView()
        .environmentObject(AppRouter.shared)
        .environmentObject(WorkoutViewModel())
}
