import Foundation
@testable import BodyFuel

final class MockPhotoService: PhotoServiceProtocol {

    // MARK: - Configurable results

    var uploadResult: Result<String, Error> = .success("https://example.com/avatar.jpg")

    // MARK: - Call tracking

    var uploadCallCount = 0
    var lastUploadedData: Data?

    // MARK: - Protocol

    func uploadUserAvatar(data: Data) async throws -> String {
        uploadCallCount += 1
        lastUploadedData = data
        return try uploadResult.get()
    }
}
