import SwiftUI
import Combine

@MainActor
final class PasswordRecoveryViewModel: ObservableObject {
    enum Step {
        case enterLogin
        case enterCode
        case success
    }

    @Published var login: String = ""
    @Published var code: String = ""
    @Published var newPassword: String = ""
    @Published var passwordError: String? = nil

    @Published var step: Step = .enterLogin
    @Published var screenState: ScreenState = .idle

    private let authService: AuthServiceProtocol = AuthService.shared

    func next() async {
        do {
            screenState = .loading
            defer { screenState = .idle }
            
            try validate()

            switch step {
            case .enterLogin:
                try await authService.sendRecoveryCode(login: login)
                step = .enterCode

            case .enterCode:
                try await authService.confirmRecovery(
                    code: code,
                    newPassword: newPassword
                )
                step = .success

            case .success: break
            }

        } catch let error as AuthError {
            screenState = .error(error.errorDescription ?? "Заполните все поля")
        } catch {
            print("[ERROR] [PasswordRecoveryViewModel/submit]: \(error.localizedDescription)")
            screenState = .error("Попробуйте еще раз позже")
        }
    }
    
    func validateLive() {
        passwordError = Validator.passwordError(newPassword)
    }
    
    private func validate() throws {
        var hasErrors = false
        switch step {
        case .enterLogin:
            hasErrors = login.isEmpty
        case .enterCode:
            hasErrors = code.isEmpty || passwordError != nil
        case .success: break
        }
        
        guard !hasErrors else {
            throw AuthError.invalidData("Заполните все поля")
        }
    }
}
