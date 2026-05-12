import Foundation
@testable import BodyFuel

final class MockFeedbackService: FeedbackServiceProtocol {

    var sendFeedbackCallCount = 0
    var lastMessage: String?
    var lastEmail: String?
    var sendFeedbackResult: Result<Void, Error> = .success(())

    func sendFeedback(message: String, email: String?) async throws {
        sendFeedbackCallCount += 1
        lastMessage = message
        lastEmail = email
        _ = try sendFeedbackResult.get()
    }
}
