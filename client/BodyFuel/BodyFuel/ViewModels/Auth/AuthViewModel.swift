import SwiftUI
import Combine

@MainActor
final class AuthViewModel: ObservableObject {
    @Published var mode: AuthMode = .login
    @Published var screenState: AuthScreenState = .idle

    @Published var login = ""
    @Published var password = ""
    @Published var passwordError: String? = nil
    @Published var confirmPassword = ""
    @Published var confirmPasswordError: String? = nil

    @Published var name = ""
    @Published var surname = ""
    @Published var phone = ""
    @Published var phoneError: String? = nil
    @Published var email = ""
    @Published var emailError: String? = nil

    private let authService: AuthServiceProtocol = AuthService.shared

    func submit() async {
        validateLive()
        
        do {
            try validate()
            screenState = .loading
            defer { screenState = .idle }

            switch mode {
            case .login:
                let payload = LoginPayload(
                    username: login,
                    password: password
                )
                try await authService.login(user: payload)
            case .register:
                let payload = RegisterPayload(
                    username: login,
                    name: name,
                    surname: surname,
                    email: email,
                    phone: phone,
                    password: password
                )
                try await authService.register(user: payload)
            }
        } catch {
            screenState = .error(error.localizedDescription)
        }
    }
    
    func validateLive() {
        passwordError = Validator.passwordError(password)
        confirmPasswordError = password != confirmPassword ? "Пароли не совпадают" : nil

        if mode == .register {
            phoneError = Validator.phoneError(phone)
            emailError = Validator.emailError(email)
        }
    }

    private func validate() throws {
        var hasErrors = false
        switch mode {
        case .login:
            hasErrors = [login, password].contains { $0.isEmpty }
        case .register:
            hasErrors = [login, name, surname].contains { $0.isEmpty } ||
            [passwordError, confirmPasswordError, phoneError, emailError].contains { $0 != nil }
        }
        
        guard !hasErrors else {
            throw AuthError.invalidData("Заполните все поля")
        }
    }
}
