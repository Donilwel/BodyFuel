import SwiftUI
import Combine

@MainActor
final class AuthViewModel: ObservableObject {
    enum AuthMode {
        case login
        case register
    }
    
    enum AuthEvent {
        case idle
        case loginSuccess
        case registrationSuccess
    }
    
    @Published var mode: AuthMode = .login
    @Published var screenState: ScreenState = .idle
    @Published var event: AuthEvent?

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
                event = .loginSuccess
            case .register:
                let registerPayload = RegisterPayload(
                    username: login,
                    name: name,
                    surname: surname,
                    email: email,
                    phone: phone,
                    password: password
                )
                try await authService.register(user: registerPayload)
                try await authService.login(user: LoginPayload(username: registerPayload.username, password: registerPayload.password))
                event = .registrationSuccess
            }
        } catch let error as AuthError {
            screenState = .error(error.errorDescription ?? "Заполните все поля")
        } catch {
            print("[ERROR] [AuthViewModel/submit]: \(error.localizedDescription)")
            screenState = .error("Попробуйте еще раз позже")
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
