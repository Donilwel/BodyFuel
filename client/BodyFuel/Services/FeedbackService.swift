import Foundation

protocol FeedbackServiceProtocol {
    func sendFeedback(message: String, email: String?) async throws
}

final class FeedbackService: FeedbackServiceProtocol {
    static let shared = FeedbackService()

    private let networkClient = NetworkClient.shared

    private init() {}

    func sendFeedback(message: String, email: String?) async throws {
        guard let url = URL(string: API.baseURLString + API.feedback) else {
            throw NetworkError.invalidURL
        }
        let body = SendFeedbackRequest(message: message, email: email)
        let _: APIMessageResponse = try await networkClient.request(
            requiresAuthorization: false,
            url: url,
            method: .post,
            requestBody: body
        )
    }
}

private struct SendFeedbackRequest: Encodable {
    let message: String
    let email: String?
}
