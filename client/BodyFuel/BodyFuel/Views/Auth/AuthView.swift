import SwiftUI

struct AuthView: View {
    @StateObject private var viewModel = AuthViewModel()
    
    var loginForm: some View {
        VStack(spacing: 16) {
            AuthTextField(title: "Логин", keyboardType: .default, text: $viewModel.login)
            PasswordField(title: "Пароль", text: $viewModel.password)
        }
    }
    
    var registerForm: some View {
        VStack(spacing: 16) {
            AuthTextField(title: "Логин", keyboardType: .default, text: $viewModel.login)
            AuthTextField(title: "Имя", keyboardType: .default, text: $viewModel.name)
            AuthTextField(title: "Фамилия", keyboardType: .default, text: $viewModel.surname)
            ValidatedField(error: viewModel.phoneError) {
                AuthTextField(
                    title: "Телефон",
                    keyboardType: .phonePad,
                    text: $viewModel.phone.onChange {
                        viewModel.validateLive()
                    }
                )
            }
            ValidatedField(error: viewModel.emailError) {
                AuthTextField(
                    title: "Почта",
                    keyboardType: .emailAddress,
                    text: $viewModel.email.onChange {
                        viewModel.validateLive()
                    }
                )
            }
            ValidatedField(error: viewModel.passwordError) {
                PasswordField(
                    title: "Пароль",
                    text: $viewModel.password.onChange {
                        viewModel.validateLive()
                    }
                )
            }
            ValidatedField(error: viewModel.confirmPasswordError) {
                PasswordField(
                    title: "Повторите пароль",
                    text: $viewModel.confirmPassword.onChange {
                        viewModel.validateLive()
                    }
                )
            }
        }
    }
    
    var formContent: some View {
        VStack(spacing: 16) {
            Picker("", selection: $viewModel.mode) {
                Text("Вход").tag(AuthMode.login)
                Text("Регистрация").tag(AuthMode.register)
            }
            .pickerStyle(.segmented)
            .padding(.bottom, 8)
            
            switch viewModel.mode {
            case .login:
                loginForm
                    .transition(.push(from: .leading).combined(with: .blurReplace))
            case .register:
                registerForm
                    .transition(.push(from: .trailing).combined(with: .blurReplace))
            }
            
            PrimaryButton(
                title: viewModel.mode == .login ? "Войти" : "Зарегистрироваться",
                isLoading: viewModel.screenState == .loading
            ) {
                Task { await viewModel.submit() }
            }
            
            if viewModel.mode == .login {
                NavigationLink("Забыли пароль?") {
                    PasswordRecoveryView()
                }
                .foregroundColor(.white)
            }
        }
        .padding(24)
        .background(
            RoundedRectangle(cornerRadius: 28)
                .fill(.ultraThinMaterial)
        )
        .padding(.horizontal, 20)
        .padding(.top, 40)
        .animation(.easeInOut(duration: 0.5), value: viewModel.mode)
    }

    var body: some View {
        NavigationStack {
            ZStack {
                AppColors.backgroundGradient.ignoresSafeArea()

                ScrollView {
                    Image("emblema")
                        .clipShape(.rect(cornerRadius: 12))
                    
                    Spacer()
                    
                    formContent
                }
            }
            .alert("Что-то пошло не так", isPresented: .constant(isError)) {
                Button("OK") { viewModel.screenState = .idle }
            } message: {
                if case let .error(message) = viewModel.screenState {
                    Text(message)
                }
            }
        }
    }

    private var isError: Bool {
        if case .error = viewModel.screenState { return true }
        return false
    }
}

#Preview {
    AuthView()
}
