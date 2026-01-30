import Combine
import Foundation

final class AppRouter: ObservableObject {
    @Published var currentFlow: AppFlow?
    
    private let tokenStorage = TokenStorage.shared
    
    init() {
        currentFlow = .auth
//        if tokenStorage.token == nil {
//            currentFlow = .auth
//        } else if UserDefaults.standard.hasCompletedProfileSetup == false {
//            currentFlow = .profileSetup
//        } else {
//            currentFlow = .main
//        }
    }
}
