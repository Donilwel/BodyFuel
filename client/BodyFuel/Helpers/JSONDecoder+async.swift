import Foundation

extension JSONDecoder {
    func decodeAsync<T: Decodable>(_ type: T.Type, from data: Data) async throws -> T {
        try await Task.detached(priority: .userInitiated) {
            try self.decode(T.self, from: data)
        }.value
    }
}
