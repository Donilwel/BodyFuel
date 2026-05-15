import XCTest
@testable import BodyFuel

@MainActor
final class FeedbackSheetViewModelTests: XCTestCase {

    var mockService: MockFeedbackService!
    var sut: FeedbackSheetViewModel!
    
    private let decoder: JSONDecoder = {
        let d = JSONDecoder()
        d.dateDecodingStrategy = .iso8601
        return d
    }()

    override func setUp() async throws {
        mockService = MockFeedbackService()
        sut = FeedbackSheetViewModel(feedbackService: mockService)
        MutationQueue.shared.clear()
        NetworkMonitor.shared.markServerReachable()
    }

    override func tearDown() async throws {
        MutationQueue.shared.clear()
        NetworkMonitor.shared.markServerReachable()
        sut = nil
        mockService = nil
    }

    // MARK: - Validation: message too short

    func test_submit_emptyMessage_setsMessageError() async {
        sut.message = ""

        await sut.submit()

        XCTAssertNotNil(sut.messageError)
        XCTAssertEqual(mockService.sendFeedbackCallCount, 0)
    }

    func test_submit_ninecharsMessage_setsMessageError() async {
        sut.message = "123456789"

        await sut.submit()

        XCTAssertNotNil(sut.messageError)
        XCTAssertEqual(mockService.sendFeedbackCallCount, 0)
    }

    func test_submit_whitespaceOnlyMessage_setsMessageError() async {
        sut.message = "          " 

        await sut.submit()

        XCTAssertNotNil(sut.messageError)
        XCTAssertEqual(mockService.sendFeedbackCallCount, 0)
    }

    func test_submit_paddedMessageUnder10_setsMessageError() async {
        sut.message = "  hello  "

        await sut.submit()

        XCTAssertNotNil(sut.messageError)
        XCTAssertEqual(mockService.sendFeedbackCallCount, 0)
    }

    func test_submit_shortMessage_doesNotDismiss() async {
        sut.message = "short"

        await sut.submit()

        XCTAssertFalse(sut.isDismissed)
    }

    func test_submit_shortMessage_doesNotEnqueueMutation() async {
        sut.message = "short"

        await sut.submit()

        XCTAssertTrue(MutationQueue.shared.mutations.isEmpty)
    }

    func test_submit_shortMessage_isSendingRemainsFlase() async {
        sut.message = "too short"

        await sut.submit()

        XCTAssertFalse(sut.isSending)
    }

    // MARK: - Validation: exactly 10 chars passes

    func test_submit_exactlyTenChars_clearsMessageError() async {
        sut.messageError = "some prior error"
        sut.message = "1234567890"

        await sut.submit()

        XCTAssertNil(sut.messageError)
        XCTAssertEqual(mockService.sendFeedbackCallCount, 1)
    }

    // MARK: - Online success

    func test_submit_online_success_callsService() async {
        sut.message = "This is a valid feedback message"
        sut.email = "user@example.com"

        await sut.submit()

        XCTAssertEqual(mockService.sendFeedbackCallCount, 1)
        XCTAssertEqual(mockService.lastMessage, "This is a valid feedback message")
    }

    func test_submit_online_success_passesEmailToService() async {
        sut.message = "This is a valid feedback message"
        sut.email = "user@example.com"

        await sut.submit()

        XCTAssertEqual(mockService.lastEmail, "user@example.com")
    }

    func test_submit_online_emptyEmail_passesNilToService() async {
        sut.message = "This is a valid feedback message"
        sut.email = ""

        await sut.submit()

        XCTAssertNil(mockService.lastEmail)
    }

    func test_submit_online_whitespaceEmail_passesNilToService() async {
        sut.message = "This is a valid feedback message"
        sut.email = "   "

        await sut.submit()

        XCTAssertNil(mockService.lastEmail)
    }

    func test_submit_online_trimmesMessageBeforeSending() async {
        sut.message = "  Valid message here  "

        await sut.submit()

        XCTAssertEqual(mockService.lastMessage, "Valid message here")
    }

    func test_submit_online_success_setsDismissed() async {
        sut.message = "This is a valid feedback message"

        await sut.submit()

        XCTAssertTrue(sut.isDismissed)
    }

    func test_submit_online_success_showsSuccessToast() async {
        sut.message = "This is a valid feedback message"

        await sut.submit()

        XCTAssertEqual(ToastService.shared.toast, "Спасибо! Ваш отзыв отправлен")
    }

    func test_submit_online_success_isSendingFalse() async {
        sut.message = "This is a valid feedback message"

        await sut.submit()

        XCTAssertFalse(sut.isSending)
    }

    func test_submit_online_success_noMutationEnqueued() async {
        sut.message = "This is a valid feedback message"

        await sut.submit()

        XCTAssertTrue(MutationQueue.shared.mutations.isEmpty)
    }

    func test_submit_online_success_messageErrorNil() async {
        sut.messageError = "previous error"
        sut.message = "This is a valid feedback message"

        await sut.submit()

        XCTAssertNil(sut.messageError)
    }

    // MARK: - Offline path

    func test_submit_offline_doesNotCallService() async {
        NetworkMonitor.shared.markServerUnreachable()
        sut.message = "This is a valid feedback message"

        await sut.submit()

        XCTAssertEqual(mockService.sendFeedbackCallCount, 0)
    }

    func test_submit_offline_enqueuesMutation() async {
        NetworkMonitor.shared.markServerUnreachable()
        sut.message = "This is a valid feedback message"

        await sut.submit()

        XCTAssertEqual(MutationQueue.shared.mutations.count, 1)
        XCTAssertEqual(MutationQueue.shared.mutations.first?.type, .sendFeedback)
    }

    func test_submit_offline_setsDismissed() async {
        NetworkMonitor.shared.markServerUnreachable()
        sut.message = "This is a valid feedback message"

        await sut.submit()

        XCTAssertTrue(sut.isDismissed)
    }

    func test_submit_offline_showsOfflineToast() async {
        NetworkMonitor.shared.markServerUnreachable()
        sut.message = "This is a valid feedback message"

        await sut.submit()

        XCTAssertEqual(ToastService.shared.toast, "Отзыв отправится при подключении к сети")
    }

    func test_submit_offline_isSendingFalse() async {
        NetworkMonitor.shared.markServerUnreachable()
        sut.message = "This is a valid feedback message"

        await sut.submit()

        XCTAssertFalse(sut.isSending)
    }

    func test_submit_offline_messageErrorNil() async {
        NetworkMonitor.shared.markServerUnreachable()
        sut.message = "This is a valid feedback message"
        sut.email = "u@e.com"

        await sut.submit()

        XCTAssertNil(sut.messageError)
    }

    func test_submit_offline_passesEmailInMutationPayload() async {
        NetworkMonitor.shared.markServerUnreachable()
        sut.message = "This is a valid feedback message"
        sut.email = "u@example.com"

        await sut.submit()

        let data = MutationQueue.shared.mutations.first?.payload
        let payload = data.flatMap { try? JSONDecoder().decode(SendFeedbackPayload.self, from: $0) }
        XCTAssertEqual(payload?.email, "u@example.com")
    }

    func test_submit_offline_emptyEmail_nilInPayload() async {
        NetworkMonitor.shared.markServerUnreachable()
        sut.message = "This is a valid feedback message"
        sut.email = ""

        await sut.submit()

        let data = MutationQueue.shared.mutations.first?.payload
        let payload = data.flatMap { try? JSONDecoder().decode(SendFeedbackPayload.self, from: $0) }
        XCTAssertNil(payload?.email)
    }

    // MARK: - Transport error

    func test_submit_transportError_enqueuesMutation() async {
        mockService.sendFeedbackResult = .failure(
            NetworkError.network(URLError(.notConnectedToInternet))
        )
        sut.message = "This is a valid feedback message"

        await sut.submit()

        XCTAssertEqual(MutationQueue.shared.mutations.count, 1)
        XCTAssertEqual(MutationQueue.shared.mutations.first?.type, .sendFeedback)
    }

    func test_submit_transportError_setsDismissed() async {
        mockService.sendFeedbackResult = .failure(
            NetworkError.network(URLError(.networkConnectionLost))
        )
        sut.message = "This is a valid feedback message"

        await sut.submit()

        XCTAssertTrue(sut.isDismissed)
    }

    func test_submit_transportError_showsOfflineToast() async {
        mockService.sendFeedbackResult = .failure(
            NetworkError.network(URLError(.timedOut))
        )
        sut.message = "This is a valid feedback message"

        await sut.submit()

        XCTAssertEqual(ToastService.shared.toast, "Отзыв отправится при подключении к сети")
    }

    func test_submit_transportError_doesNotSetMessageError() async {
        mockService.sendFeedbackResult = .failure(
            NetworkError.network(URLError(.notConnectedToInternet))
        )
        sut.message = "This is a valid feedback message"

        await sut.submit()

        XCTAssertNil(sut.messageError)
    }

    func test_submit_transportError_isSendingFalse() async {
        mockService.sendFeedbackResult = .failure(
            NetworkError.network(URLError(.notConnectedToInternet))
        )
        sut.message = "This is a valid feedback message"

        await sut.submit()

        XCTAssertFalse(sut.isSending)
    }

    // MARK: - Server error

    func test_submit_serverError_setsMessageError() async {
        mockService.sendFeedbackResult = .failure(
            NetworkError.requestFailed(statusCode: 500, message: "Internal Server Error")
        )
        sut.message = "This is a valid feedback message"

        await sut.submit()

        XCTAssertNotNil(sut.messageError)
    }

    func test_submit_serverError_doesNotDismiss() async {
        mockService.sendFeedbackResult = .failure(
            NetworkError.requestFailed(statusCode: 503, message: "Unavailable")
        )
        sut.message = "This is a valid feedback message"

        await sut.submit()

        XCTAssertFalse(sut.isDismissed)
    }

    func test_submit_serverError_doesNotEnqueueMutation() async {
        mockService.sendFeedbackResult = .failure(
            NetworkError.requestFailed(statusCode: 500, message: "")
        )
        sut.message = "This is a valid feedback message"

        await sut.submit()

        XCTAssertTrue(MutationQueue.shared.mutations.isEmpty)
    }

    func test_submit_serverError_isSendingFalse() async {
        mockService.sendFeedbackResult = .failure(
            NetworkError.requestFailed(statusCode: 422, message: "Unprocessable")
        )
        sut.message = "This is a valid feedback message"

        await sut.submit()

        XCTAssertFalse(sut.isSending)
    }

    func test_submit_serverError_400_setsMessageError() async {
        mockService.sendFeedbackResult = .failure(
            NetworkError.requestFailed(statusCode: 400, message: "Bad request")
        )
        sut.message = "This is a valid feedback message"

        await sut.submit()

        XCTAssertNotNil(sut.messageError)
        XCTAssertFalse(sut.isDismissed)
    }

    // MARK: - Second submit

    func test_submit_afterValidationError_validMessage_clearsError() async {
        sut.message = "short"
        await sut.submit()
        XCTAssertNotNil(sut.messageError)

        sut.message = "This is now a valid feedback message"
        await sut.submit()

        XCTAssertNil(sut.messageError)
        XCTAssertEqual(mockService.sendFeedbackCallCount, 1)
    }

    // MARK: - isDismissed idempotency

    func test_submit_calledTwiceOnSuccess_doesNotCallServiceSecondTime() async {
        sut.message = "This is a valid feedback message"
        await sut.submit()

        await sut.submit()

        XCTAssertEqual(mockService.sendFeedbackCallCount, 2)
    }
}
