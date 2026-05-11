import SwiftUI

struct FeedbackSheet: View {
    let title: String

    @Environment(\.dismiss) private var dismiss

    @State private var message = ""
    @State private var email = ""
    @State private var isSending = false
    @State private var messageError: String? = nil

    var body: some View {
        NavigationStack {
            ZStack {
                AnimatedBackground()
                    .ignoresSafeArea()

                ScrollView {
                    VStack(alignment: .leading, spacing: 20) {
                        Text("Ваш отзыв поможет нам улучшить работу ИИ")
                            .font(.subheadline)
                            .foregroundStyle(.white.opacity(0.7))

                        VStack(alignment: .leading, spacing: 8) {
                            Text("Сообщение")
                                .font(.caption)
                                .foregroundStyle(.white.opacity(0.6))

                            TextEditor(text: $message)
                                .frame(minHeight: 120)
                                .padding(12)
                                .background(.ultraThinMaterial)
                                .cornerRadius(12)
                                .foregroundStyle(.white)
                                .scrollContentBackground(.hidden)

                            if let error = messageError {
                                Text(error)
                                    .font(.caption)
                                    .foregroundStyle(.red.opacity(0.9))
                            }
                        }

                        VStack(alignment: .leading, spacing: 8) {
                            Text("Email для обратной связи (необязательно)")
                                .font(.caption)
                                .foregroundStyle(.white.opacity(0.6))

                            TextField("", text: $email)
                                .keyboardType(.emailAddress)
                                .autocapitalization(.none)
                                .textContentType(.emailAddress)
                                .padding(12)
                                .background(.ultraThinMaterial)
                                .cornerRadius(12)
                                .foregroundStyle(.white)
                        }

                        PrimaryButton(title: "Отправить", isLoading: isSending) {
                            submit()
                        }
                        .disabled(isSending)
                    }
                    .padding()
                }
            }
            .navigationTitle(title)
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button("Закрыть") { dismiss() }
                        .foregroundStyle(.white)
                }
            }
        }
    }

    private func submit() {
        let trimmedMessage = message.trimmingCharacters(in: .whitespaces)
        guard trimmedMessage.count >= 10 else {
            messageError = "Сообщение должно быть не менее 10 символов"
            HapticService.notification(.warning)
            return
        }
        messageError = nil
        isSending = true

        let trimmedEmail = email.trimmingCharacters(in: .whitespaces)
        let emailValue: String? = trimmedEmail.isEmpty ? nil : trimmedEmail

        if !NetworkMonitor.shared.isOnline {
            MutationQueue.shared.enqueue(
                type: .sendFeedback,
                payload: SendFeedbackPayload(message: trimmedMessage, email: emailValue)
            )
            isSending = false
            dismiss()
            ToastService.shared.show("Отзыв отправится при подключении к сети")
            return
        }

        Task { @MainActor in
            do {
                try await FeedbackService.shared.sendFeedback(message: trimmedMessage, email: emailValue)
                HapticService.notification(.success)
                isSending = false
                dismiss()
                ToastService.shared.show("Спасибо! Ваш отзыв отправлен")
            } catch {
                if isTransportError(error) {
                    MutationQueue.shared.enqueue(
                        type: .sendFeedback,
                        payload: SendFeedbackPayload(message: trimmedMessage, email: emailValue)
                    )
                    isSending = false
                    dismiss()
                    ToastService.shared.show("Отзыв отправится при подключении к сети")
                } else {
                    HapticService.notification(.error)
                    isSending = false
                    messageError = "Не удалось отправить отзыв, попробуйте позже"
                }
            }
        }
    }
}
