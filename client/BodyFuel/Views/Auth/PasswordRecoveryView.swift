import SwiftUI

struct PasswordRecoveryView: View {
    @StateObject private var viewModel = PasswordRecoveryViewModel()
    @Environment(\.dismiss) private var dismiss

    @FocusState private var emailFocused: EmailField?
    @FocusState private var recoveryFocused: RecoveryField?

    private enum EmailField: Hashable {
        case email
    }
    private enum RecoveryField: Hashable {
        case smsCode
        case password
    }

    var body: some View {
        ZStack {
            AnimatedBackground()
                .ignoresSafeArea()

            VStack(spacing: 20) {
                if viewModel.step == .success {
                    VStack(spacing: 16) {
                        Text("Пароль успешно изменён")
                            .font(.title2.bold())
                            .foregroundColor(.white)
                        PrimaryButton(title: "Войти") {
                            dismiss()
                        }
                    }
                } else {
                    Text("Восстановление пароля")
                        .font(.title2.bold())
                        .foregroundColor(.white)

                    Group {
                        switch viewModel.step {
                        case .enterEmail:
                            CustomTextField(
                                title: "Email",
                                keyboardType: .emailAddress,
                                field: EmailField.email,
                                focusedField: $emailFocused,
                                text: $viewModel.email
                            )
                        case .enterCode:
                            CustomTextField(
                                title: "Код из письма",
                                keyboardType: .numberPad,
                                field: RecoveryField.smsCode,
                                focusedField: $recoveryFocused,
                                text: $viewModel.code
                            )
                            ValidatedField(error: viewModel.passwordError) {
                                PasswordField(
                                    title: "Новый пароль",
                                    field: RecoveryField.password,
                                    focusedField: $recoveryFocused,
                                    text: $viewModel.newPassword.onChange {
                                        viewModel.validateLive()
                                    }
                                )
                            }
                        case .success:
                            EmptyView()
                        }
                    }

                    PrimaryButton(
                        title: viewModel.step == .enterEmail ? "Отправить код" : "Сменить пароль",
                        isLoading: viewModel.screenState == .loading
                    ) {
                        Task { await viewModel.next() }
                    }
                }
            }
            .padding(24)
            .background(
                RoundedRectangle(cornerRadius: 28)
                    .fill(.ultraThinMaterial)
            )
            .padding(.horizontal, 20)
        }
        .alert("Ошибка", isPresented: .constant(isError)) {
            Button("OK") { viewModel.screenState = .idle }
        } message: {
            if case let .error(message) = viewModel.screenState {
                Text(message)
            }
        }
        .navigationBarTitleDisplayMode(.inline)
        .onChange(of: viewModel.step) {
            resetFocusStates()
        }
        .onTapGesture {
            resetFocusStates()
        }
    }

    private var isError: Bool {
        if case .error = viewModel.screenState { return true }
        return false
    }

    private func resetFocusStates() {
        emailFocused = nil
        recoveryFocused = nil
    }
}

#Preview {
    PasswordRecoveryView()
}
