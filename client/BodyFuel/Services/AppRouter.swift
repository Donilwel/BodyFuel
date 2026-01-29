import Combine

final class AppRouter: ObservableObject {
    @Published var currentFlow: AppFlow = .auth
}
