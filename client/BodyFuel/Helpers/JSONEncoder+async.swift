import Foundation

extension JSONEncoder {
    func encodeAsync<T: Encodable>(_ value: T) async throws -> Data {
        try await Task.detached(priority: .userInitiated) {
            try self.encode(value)
        }.value
    }
}
