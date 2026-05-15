import XCTest
@testable import BodyFuel

final class DeepLinkTests: XCTestCase {

    // MARK: - .workouts

    func test_workouts_noPath_returnsWorkouts() {
        let url = URL(string: "bodyfuel://workouts")!
        XCTAssertEqual(DeepLink(url: url), .workouts)
    }

    func test_workouts_trailingSlash_returnsWorkouts() {
        let url = URL(string: "bodyfuel://workouts/")!
        XCTAssertEqual(DeepLink(url: url), .workouts)
    }

    // MARK: - .workoutsWithID

    func test_workoutsWithID_returnsWorkoutsWithID() {
        let url = URL(string: "bodyfuel://workouts/abc123")!
        XCTAssertEqual(DeepLink(url: url), .workoutsWithID("abc123"))
    }

    func test_workoutsWithID_uuid_returnsCorrectID() {
        let id = "550e8400-e29b-41d4-a716-446655440000"
        let url = URL(string: "bodyfuel://workouts/\(id)")!
        XCTAssertEqual(DeepLink(url: url), .workoutsWithID(id))
    }

    func test_workoutsWithID_numericID_returnsCorrectID() {
        let url = URL(string: "bodyfuel://workouts/42")!
        XCTAssertEqual(DeepLink(url: url), .workoutsWithID("42"))
    }

    // MARK: - .food

    func test_food_returnsFood() {
        let url = URL(string: "bodyfuel://food")!
        XCTAssertEqual(DeepLink(url: url), .food)
    }

    // MARK: - .calories

    func test_calories_returnsCalories() {
        let url = URL(string: "bodyfuel://calories")!
        XCTAssertEqual(DeepLink(url: url), .calories)
    }

    // MARK: - nil cases

    func test_wrongScheme_https_returnsNil() {
        let url = URL(string: "https://example.com/workouts")!
        XCTAssertNil(DeepLink(url: url))
    }

    func test_wrongScheme_http_returnsNil() {
        let url = URL(string: "http://workouts")!
        XCTAssertNil(DeepLink(url: url))
    }

    func test_unknownHost_returnsNil() {
        let url = URL(string: "bodyfuel://unknown")!
        XCTAssertNil(DeepLink(url: url))
    }

    func test_emptyHost_returnsNil() {
        let url = URL(string: "bodyfuel://")!
        XCTAssertNil(DeepLink(url: url))
    }

    func test_unknownHost_withPath_returnsNil() {
        let url = URL(string: "bodyfuel://settings/profile")!
        XCTAssertNil(DeepLink(url: url))
    }

    func test_wrongScheme_noScheme_returnsNil() {
        let url = URL(string: "workouts")!
        XCTAssertNil(DeepLink(url: url))
    }

    // MARK: - Equatable helpers

    func test_workouts_doesNotMatchFood() {
        let url = URL(string: "bodyfuel://workouts")!
        XCTAssertNotEqual(DeepLink(url: url), .food)
    }

    func test_workoutsWithID_doesNotMatchWorkoutsNoID() {
        let url = URL(string: "bodyfuel://workouts/someID")!
        XCTAssertNotEqual(DeepLink(url: url), .workouts)
    }

    func test_workoutsWithID_differentIDs_notEqual() {
        let url1 = URL(string: "bodyfuel://workouts/id-1")!
        let url2 = URL(string: "bodyfuel://workouts/id-2")!
        XCTAssertNotEqual(DeepLink(url: url1), DeepLink(url: url2))
    }
}

extension DeepLink: Equatable {
    public static func == (lhs: DeepLink, rhs: DeepLink) -> Bool {
        switch (lhs, rhs) {
        case (.workouts, .workouts): return true
        case (.food, .food): return true
        case (.calories, .calories): return true
        case (.workoutsWithID(let a), .workoutsWithID(let b)): return a == b
        default: return false
        }
    }
}
