import XCTest
@testable import BodyFuel

final class DiskCacheTests: XCTestCase {

    var sut: DiskCache!

    private func key(_ name: String) -> String { "test_diskcache_\(name)" }

    override func setUp() {
        sut = DiskCache.shared
        cleanUp()
    }

    override func tearDown() {
        cleanUp()
    }

    private func cleanUp() {
        let keys = [
            "alpha", "beta", "gamma", "int", "string", "bool",
            "array", "nested", "optional", "multi_a", "multi_b",
            "ttl_expired", "ttl_fresh", "remove_me", "keep_me",
            "day_check", "savedAt_check"
        ]
        keys.forEach { sut.remove(key: key($0)) }
    }

    // MARK: - Helpers

    private struct SimpleModel: Codable, Equatable {
        let id: Int
        let name: String
    }

    private struct NestedModel: Codable, Equatable {
        let value: Double
        let tags: [String]
        let inner: SimpleModel
    }

    // MARK: - save / load roundtrip

    func test_saveLoad_string_roundtrip() {
        sut.save("Hello, BodyFuel!", key: key("string"))
        let loaded = sut.load(String.self, key: key("string"))
        XCTAssertEqual(loaded, "Hello, BodyFuel!")
    }

    func test_saveLoad_int_roundtrip() {
        sut.save(42, key: key("int"))
        let loaded = sut.load(Int.self, key: key("int"))
        XCTAssertEqual(loaded, 42)
    }

    func test_saveLoad_bool_roundtrip() {
        sut.save(true, key: key("bool"))
        let loaded = sut.load(Bool.self, key: key("bool"))
        XCTAssertEqual(loaded, true)
    }

    func test_saveLoad_customStruct_roundtrip() {
        let model = SimpleModel(id: 7, name: "Test Item")
        sut.save(model, key: key("alpha"))
        let loaded = sut.load(SimpleModel.self, key: key("alpha"))
        XCTAssertEqual(loaded, model)
    }

    func test_saveLoad_array_roundtrip() {
        let models = [SimpleModel(id: 1, name: "One"), SimpleModel(id: 2, name: "Two")]
        sut.save(models, key: key("array"))
        let loaded = sut.load([SimpleModel].self, key: key("array"))
        XCTAssertEqual(loaded, models)
    }

    func test_saveLoad_emptyArray_roundtrip() {
        let empty: [SimpleModel] = []
        sut.save(empty, key: key("array"))
        let loaded = sut.load([SimpleModel].self, key: key("array"))
        XCTAssertEqual(loaded, [])
    }

    func test_saveLoad_nestedStruct_roundtrip() {
        let model = NestedModel(value: 3.14, tags: ["swift", "ios"], inner: SimpleModel(id: 99, name: "Inner"))
        sut.save(model, key: key("nested"))
        let loaded = sut.load(NestedModel.self, key: key("nested"))
        XCTAssertEqual(loaded, model)
    }

    func test_saveLoad_overwrite_returnsLatestValue() {
        sut.save(SimpleModel(id: 1, name: "First"), key: key("alpha"))
        sut.save(SimpleModel(id: 2, name: "Second"), key: key("alpha"))
        let loaded = sut.load(SimpleModel.self, key: key("alpha"))
        XCTAssertEqual(loaded?.id, 2)
        XCTAssertEqual(loaded?.name, "Second")
    }

    func test_saveLoad_multipleKeys_areIndependent() {
        sut.save(SimpleModel(id: 1, name: "A"), key: key("multi_a"))
        sut.save(SimpleModel(id: 2, name: "B"), key: key("multi_b"))
        XCTAssertEqual(sut.load(SimpleModel.self, key: key("multi_a"))?.id, 1)
        XCTAssertEqual(sut.load(SimpleModel.self, key: key("multi_b"))?.id, 2)
    }

    // MARK: - load — missing / wrong type

    func test_load_missingKey_returnsNil() {
        let loaded = sut.load(SimpleModel.self, key: key("nonexistent_\(UUID())"))
        XCTAssertNil(loaded)
    }

    func test_load_wrongType_returnsNil() {
        sut.save(SimpleModel(id: 1, name: "X"), key: key("alpha"))
        let loaded = sut.load([Int].self, key: key("alpha"))
        XCTAssertNil(loaded)
    }

    // MARK: - isExpired

    func test_isExpired_negativeTTL_alwaysTrue() {
        sut.save(SimpleModel(id: 1, name: "X"), key: key("ttl_expired"))
        XCTAssertTrue(sut.isExpired(key: key("ttl_expired"), ttl: -1))
    }

    func test_isExpired_largeTTL_justSaved_isFalse() {
        sut.save(SimpleModel(id: 1, name: "X"), key: key("ttl_fresh"))
        XCTAssertFalse(sut.isExpired(key: key("ttl_fresh"), ttl: 3600))
    }

    func test_isExpired_missingKey_returnsTrue() {
        XCTAssertTrue(sut.isExpired(key: key("missing_\(UUID())"), ttl: 3600))
    }

    func test_isExpired_zeroTTL_expiredImmediately() {
        sut.save("value", key: key("ttl_expired"))
        Thread.sleep(forTimeInterval: 0.001)
        XCTAssertTrue(sut.isExpired(key: key("ttl_expired"), ttl: 0))
    }

    func test_isExpired_oneHourTTL_freshSave_notExpired() {
        sut.save(42, key: key("ttl_fresh"))
        XCTAssertFalse(sut.isExpired(key: key("ttl_fresh"), ttl: 60 * 60))
    }

    // MARK: - remove

    func test_remove_makesLoadReturnNil() {
        sut.save(SimpleModel(id: 5, name: "ToDelete"), key: key("remove_me"))
        sut.remove(key: key("remove_me"))
        XCTAssertNil(sut.load(SimpleModel.self, key: key("remove_me")))
    }

    func test_remove_doesNotAffectOtherKeys() {
        sut.save(SimpleModel(id: 1, name: "Stay"), key: key("keep_me"))
        sut.save(SimpleModel(id: 2, name: "Gone"), key: key("remove_me"))

        sut.remove(key: key("remove_me"))

        XCTAssertNotNil(sut.load(SimpleModel.self, key: key("keep_me")))
        XCTAssertNil(sut.load(SimpleModel.self, key: key("remove_me")))
    }

    func test_remove_nonExistentKey_noError() {
        sut.remove(key: key("never_saved_\(UUID())"))
    }

    func test_remove_makesIsExpiredReturnTrue() {
        sut.save("value", key: key("remove_me"))
        sut.remove(key: key("remove_me"))
        XCTAssertTrue(sut.isExpired(key: key("remove_me"), ttl: 3600))
    }

    func test_remove_afterRemove_canSaveAgain() {
        sut.save(SimpleModel(id: 1, name: "First"), key: key("remove_me"))
        sut.remove(key: key("remove_me"))
        sut.save(SimpleModel(id: 2, name: "Second"), key: key("remove_me"))
        XCTAssertEqual(sut.load(SimpleModel.self, key: key("remove_me"))?.id, 2)
    }

    // MARK: - removeAll

    func test_removeAll_makesAllKeysReturnNil() {
        sut.save(SimpleModel(id: 1, name: "A"), key: key("multi_a"))
        sut.save(SimpleModel(id: 2, name: "B"), key: key("multi_b"))
        sut.save("text", key: key("string"))

        sut.removeAll()

        XCTAssertNil(sut.load(SimpleModel.self, key: key("multi_a")))
        XCTAssertNil(sut.load(SimpleModel.self, key: key("multi_b")))
        XCTAssertNil(sut.load(String.self, key: key("string")))
    }

    func test_removeAll_afterRemoveAll_canSaveNewValues() {
        sut.save(SimpleModel(id: 1, name: "Before"), key: key("alpha"))
        sut.removeAll()
        sut.save(SimpleModel(id: 2, name: "After"), key: key("alpha"))
        XCTAssertEqual(sut.load(SimpleModel.self, key: key("alpha"))?.id, 2)
    }

    func test_removeAll_emptyCache_noError() {
        sut.removeAll()
    }

    // MARK: - isFromDifferentDay

    func test_isFromDifferentDay_justSaved_returnsFalse() {
        sut.save(SimpleModel(id: 1, name: "Today"), key: key("day_check"))
        XCTAssertFalse(sut.isFromDifferentDay(key: key("day_check")))
    }

    func test_isFromDifferentDay_missingKey_returnsTrue() {
        XCTAssertTrue(sut.isFromDifferentDay(key: key("missing_\(UUID())")))
    }

    // MARK: - savedAt

    func test_savedAt_returnsApproximateCurrentTime() {
        let before = Date()
        sut.save("value", key: key("savedAt_check"))
        let after = Date()

        let savedAt = sut.savedAt(key: key("savedAt_check"))

        XCTAssertNotNil(savedAt)
        XCTAssertLessThanOrEqual(savedAt!, before)
        XCTAssertLessThanOrEqual(savedAt!, after.addingTimeInterval(1))
    }

    func test_savedAt_missingKey_returnsNil() {
        XCTAssertNil(sut.savedAt(key: key("missing_\(UUID())")))
    }

    func test_savedAt_updatesOnOverwrite() {
        sut.save("first", key: key("savedAt_check"))
        let first = sut.savedAt(key: key("savedAt_check"))

        Thread.sleep(forTimeInterval: 0.01)

        sut.save("second", key: key("savedAt_check"))
        let second = sut.savedAt(key: key("savedAt_check"))

        XCTAssertNotNil(first)
        XCTAssertNotNil(second)
        XCTAssertLessThanOrEqual(second!, first!)
    }
}
