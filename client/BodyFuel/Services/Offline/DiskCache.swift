import Foundation

final class DiskCache {
    static let shared = DiskCache()

    private let baseURL: URL
    private let encoder: JSONEncoder = {
        let e = JSONEncoder()
        e.dateEncodingStrategy = .iso8601
        return e
    }()
    private let decoder: JSONDecoder = {
        let d = JSONDecoder()
        d.dateDecodingStrategy = .iso8601
        return d
    }()

    private init() {
        let appSupport = FileManager.default.urls(
            for: .applicationSupportDirectory,
            in: .userDomainMask
        ).first!
        baseURL = appSupport.appendingPathComponent("BodyFuel/cache", isDirectory: true)
        try? FileManager.default.createDirectory(at: baseURL, withIntermediateDirectories: true)
    }

    // MARK: - Envelope

    private struct Envelope<T: Codable>: Codable {
        let savedAt: Date
        let value: T
    }

    private struct EnvelopeMeta: Decodable {
        let savedAt: Date
    }

    // MARK: - Public API

    func save<T: Codable>(_ value: T, key: String) {
        guard let data = try? encoder.encode(Envelope(savedAt: Date(), value: value)) else { return }
        try? data.write(to: fileURL(key: key), options: .atomic)
    }

    func load<T: Codable>(_ type: T.Type, key: String) -> T? {
        guard let data = try? Data(contentsOf: fileURL(key: key)) else { return nil }
        return (try? decoder.decode(Envelope<T>.self, from: data))?.value
    }

    func isExpired(key: String, ttl: TimeInterval) -> Bool {
        guard let data = try? Data(contentsOf: fileURL(key: key)),
              let meta = try? decoder.decode(EnvelopeMeta.self, from: data) else {
            return true
        }
        return Date().timeIntervalSince(meta.savedAt) > ttl
    }

    func isFromDifferentDay(key: String) -> Bool {
        guard let data = try? Data(contentsOf: fileURL(key: key)),
              let meta = try? decoder.decode(EnvelopeMeta.self, from: data) else {
            return true
        }
        return !Calendar.current.isDateInToday(meta.savedAt)
    }

    func savedAt(key: String) -> Date? {
        guard let data = try? Data(contentsOf: fileURL(key: key)),
              let meta = try? decoder.decode(EnvelopeMeta.self, from: data) else {
            return nil
        }
        return meta.savedAt
    }

    func remove(key: String) {
        try? FileManager.default.removeItem(at: fileURL(key: key))
    }

    func removeAll() {
        try? FileManager.default.removeItem(at: baseURL)
        try? FileManager.default.createDirectory(at: baseURL, withIntermediateDirectories: true)
    }

    // MARK: - Private

    private func fileURL(key: String) -> URL {
        baseURL.appendingPathComponent("\(key).json")
    }
}
