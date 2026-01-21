import SwiftUI

struct PasswordRecoveryView: View {
    @StateObject private var viewModel = PasswordRecoveryViewModel()

    var body: some View {
        ZStack {
            AppColors.backgroundGradient.ignoresSafeArea()

            VStack(spacing: 20) {
                if viewModel.step == .success {
                    Text("Пароль успешно восстановлен")
                        .font(.title2.bold())
                        .foregroundColor(.white)
                } else {
                    Text("Восстановление пароля")
                        .font(.title2.bold())
                        .foregroundColor(.white)
                    
                    Group {
                        switch viewModel.step {
                        case .enterLogin:
                            AuthTextField(title: "Логин", keyboardType: .default, text: $viewModel.login)
                        case .enterCode:
                            AuthTextField(title: "Код из СМС", keyboardType: .numberPad, text: $viewModel.code)
                            ValidatedField(error: viewModel.passwordError) {
                                PasswordField(
                                    title: "Новый пароль",
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
                        title: viewModel.step == .enterLogin ? "Отправить код" : "Сменить пароль",
                        isLoading: false
                    ) {
                        Task { await viewModel.next() }
                    }
                }
            }
            .padding(24)
            .background(
                RoundedRectangle(cornerRadius: 28)
                    .fill(.ultraThinMaterial)
                    .overlay(
                        RoundedRectangle(cornerRadius: 28)
                            .stroke(Color.white.opacity(0.2), lineWidth: 1)
                    )
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
    }
    
    private var isError: Bool {
        if case .error = viewModel.screenState { return true }
        return false
    }
}

#Preview {
    PasswordRecoveryView()
}

