import SwiftUI

struct AuthView: View {
    @EnvironmentObject var router: AppRouter
    @StateObject private var viewModel = AuthViewModel()
    
    @FocusState private var loginFocused: LoginField?
    @FocusState private var registerFocused: RegisterField?
    
    private enum LoginField: Hashable {
        case login
        case password
    }
    private enum RegisterField: Hashable {
        case login
        case name
        case surname
        case phone
        case email
        case password
        case passwordConfirmation
    }
    
    private var loginForm: some View {
        VStack(spacing: 16) {
            CustomTextField(
                title: "Логин",
                field: LoginField.login,
                focusedField: $loginFocused,
                text: $viewModel.login
            )
            PasswordField(
                title: "Пароль",
                field: LoginField.password,
                focusedField: $loginFocused,
                text: $viewModel.password
            )
        }
    }
    
    private var registerForm: some View {
        VStack(spacing: 16) {
            CustomTextField(
                title: "Логин",
                field: RegisterField.login,
                focusedField: $registerFocused,
                text: $viewModel.login
            )
            CustomTextField(
                title: "Имя",
                field: RegisterField.name,
                focusedField: $registerFocused,
                text: $viewModel.name
            )
            CustomTextField(
                title: "Фамилия",
                field: RegisterField.surname,
                focusedField: $registerFocused,
                text: $viewModel.surname
            )
            ValidatedField(error: viewModel.phoneError) {
                CustomTextField(
                    title: "Телефон",
                    keyboardType: .phonePad,
                    field: RegisterField.phone,
                    focusedField: $registerFocused,
                    text: $viewModel.phone.onChange {
                        viewModel.validateLive()
                    }
                )
            }
            ValidatedField(error: viewModel.emailError) {
                CustomTextField(
                    title: "Почта",
                    keyboardType: .emailAddress,
                    field: RegisterField.email,
                    focusedField: $registerFocused,
                    text: $viewModel.email.onChange {
                        viewModel.validateLive()
                    }
                )
            }
            ValidatedField(error: viewModel.passwordError) {
                PasswordField(
                    title: "Пароль",
                    field: RegisterField.password,
                    focusedField: $registerFocused,
                    text: $viewModel.password.onChange {
                        viewModel.validateLive()
                    }
                )
            }
            ValidatedField(error: viewModel.confirmPasswordError) {
                PasswordField(
                    title: "Повторите пароль",
                    field: RegisterField.passwordConfirmation,
                    focusedField: $registerFocused,
                    text: $viewModel.confirmPassword.onChange {
                        viewModel.validateLive()
                    }
                )
            }
        }
    }
    
    private var formContent: some View {
        VStack(spacing: 16) {
            Picker("", selection: $viewModel.mode) {
                Text("Вход").tag(AuthViewModel.AuthMode.login)
                Text("Регистрация").tag(AuthViewModel.AuthMode.register)
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
        NavigationStack() {
            ZStack(alignment: .center) {
                AnimatedBackground()

                ScrollView {
//                    Image("emblema")
//                        .clipShape(.rect(cornerRadius: 12))
//                    
//                    Spacer()
                    
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
            .onChange(of: viewModel.event) { event in
                switch event {
                case .loginSuccess:
                    router.currentFlow = .main
                case .registrationSuccess:
                    router.currentFlow = .onboarding
                default:
                    break
                }
                
                viewModel.event = .idle
            }
            .onChange(of: viewModel.mode) {
                resetFocusStates()
            }
            .onTapGesture {
                resetFocusStates()
            }
        }
    }

    private var isError: Bool {
        if case .error = viewModel.screenState { return true }
        return false
    }
    
    private func resetFocusStates() {
        loginFocused = nil
        registerFocused = nil
    }
}

#Preview {
    AuthView()
}
