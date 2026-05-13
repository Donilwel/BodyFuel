import SwiftUI

struct FeedbackSheet: View {
    let title: String

    @Environment(\.dismiss) private var dismiss
    @StateObject private var viewModel = FeedbackSheetViewModel()

    var body: some View {
        NavigationStack {
            ZStack {
                Color.clear
                    .glassEffect(.regular.tint(AppColors.primary.opacity(0.6)).interactive(), in: .rect)
                    .ignoresSafeArea()

                ScrollView {
                    Text(title)
                        .sheetTitle()

                    VStack(alignment: .leading, spacing: 20) {
                        Text("Ваш отзыв поможет нам улучшить работу ИИ")
                            .font(.subheadline)
                            .foregroundStyle(.white.opacity(0.7))

                        VStack(alignment: .leading, spacing: 8) {
                            Text("Сообщение")
                                .font(.caption)
                                .foregroundStyle(.white.opacity(0.6))

                            TextEditor(text: $viewModel.message)
                                .frame(minHeight: 120)
                                .padding(12)
                                .background(.ultraThinMaterial)
                                .cornerRadius(12)
                                .foregroundStyle(.white)
                                .scrollContentBackground(.hidden)

                            if let error = viewModel.messageError {
                                Text(error)
                                    .font(.caption)
                                    .foregroundStyle(.red.opacity(0.9))
                            }
                        }

                        VStack(alignment: .leading, spacing: 8) {
                            Text("Email для обратной связи (необязательно)")
                                .font(.caption)
                                .foregroundStyle(.white.opacity(0.6))

                            TextField("", text: $viewModel.email)
                                .keyboardType(.emailAddress)
                                .autocapitalization(.none)
                                .textContentType(.emailAddress)
                                .padding(12)
                                .background(.ultraThinMaterial)
                                .cornerRadius(12)
                                .foregroundStyle(.white)
                        }

                        PrimaryButton(title: "Отправить", isLoading: viewModel.isSending) {
                            Task { await viewModel.submit() }
                        }
                        .disabled(viewModel.isSending)
                    }
                    .padding()
                }
            }
        }
        .onChange(of: viewModel.isDismissed) { _, dismissed in
            if dismissed { dismiss() }
        }
    }
}
