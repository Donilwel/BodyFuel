import Foundation
import Combine
import UIKit

@MainActor
final class FeedbackSheetViewModel: ObservableObject {
    @Published var message = ""
    @Published var email = ""
    @Published var isSending = false
    @Published var messageError: String? = nil
    @Published var isDismissed = false

    private let feedbackService: FeedbackServiceProtocol

    init(feedbackService: FeedbackServiceProtocol = FeedbackService.shared) {
        self.feedbackService = feedbackService
    }

    func submit() async {
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
            isDismissed = true
            ToastService.shared.show("Отзыв отправится при подключении к сети")
            return
        }

        do {
            try await feedbackService.sendFeedback(message: trimmedMessage, email: emailValue)
            HapticService.notification(.success)
            isSending = false
            isDismissed = true
            ToastService.shared.show("Спасибо! Ваш отзыв отправлен")
        } catch {
            if isTransportError(error) {
                MutationQueue.shared.enqueue(
                    type: .sendFeedback,
                    payload: SendFeedbackPayload(message: trimmedMessage, email: emailValue)
                )
                isSending = false
                isDismissed = true
                ToastService.shared.show("Отзыв отправится при подключении к сети")
            } else {
                HapticService.notification(.error)
                isSending = false
                messageError = "Не удалось отправить отзыв, попробуйте позже"
            }
        }
    }
}
