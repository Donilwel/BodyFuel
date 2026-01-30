import Combine

final class AppRouter: ObservableObject {
    @Published var currentFlow: AppFlow?
    
    private let tokenStorage = TokenStorage.shared
    
    init() {
        currentFlow = .auth
//        currentFlow = tokenStorage.token == nil ? .auth : .main
    }
}
