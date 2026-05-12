import XCTest
@testable import BodyFuel

@MainActor
final class AppRouterTests: XCTestCase {

    var sut: AppRouter!

    override func setUp() async throws {
        sut = AppRouter.shared
        sut.rootRoute = .main
        sut.selectedTab = .stats
        sut.pendingAddMeal = true
    }

    override func tearDown() async throws {
        sut = nil
    }

    func test_handleIfUnauthorized_returnsTrue_forRequestFailed401() {
        let error = NetworkError.requestFailed(statusCode: 401, message: "")
        XCTAssertTrue(sut.handleIfUnauthorized(error))
    }

    func test_handleIfUnauthorized_returnsTrue_forMissingToken() {
        XCTAssertTrue(sut.handleIfUnauthorized(NetworkError.missingToken))
    }

    func test_handleIfUnauthorized_returnsTrue_forUserParamsNotFound() {
        let error = NetworkError.requestFailed(statusCode: 404, message: "user params not found")
        XCTAssertTrue(sut.handleIfUnauthorized(error))
    }

    func test_handleIfUnauthorized_returnsTrue_forUserParamsNotFound_caseInsensitive() {
        let error = NetworkError.requestFailed(statusCode: 404, message: "User Params Not Found")
        XCTAssertTrue(sut.handleIfUnauthorized(error))
    }

    func test_handleIfUnauthorized_returnsFalse_forNoInternet() {
        let error = NetworkError.network(URLError(.notConnectedToInternet))
        XCTAssertFalse(sut.handleIfUnauthorized(error))
    }

    func test_handleIfUnauthorized_returnsFalse_for500() {
        let error = NetworkError.requestFailed(statusCode: 500, message: "")
        XCTAssertFalse(sut.handleIfUnauthorized(error))
    }

    func test_handleIfUnauthorized_returnsFalse_for400() {
        let error = NetworkError.requestFailed(statusCode: 400, message: "Bad request")
        XCTAssertFalse(sut.handleIfUnauthorized(error))
    }

    func test_handleIfUnauthorized_returnsFalse_forDecodingFailed() {
        XCTAssertFalse(sut.handleIfUnauthorized(NetworkError.decodingFailed))
    }

    func test_handleIfUnauthorized_returnsFalse_forGenericError() {
        struct SomeError: Error {}
        XCTAssertFalse(sut.handleIfUnauthorized(SomeError()))
    }

    func test_handleIfUnauthorized_returnsFalse_for404WithOtherMessage() {
        let error = NetworkError.requestFailed(statusCode: 404, message: "not found")
        XCTAssertFalse(sut.handleIfUnauthorized(error))
    }

    func test_handleIfUnauthorized_unauthorized_setsRootRouteToAuth() {
        let error = NetworkError.requestFailed(statusCode: 401, message: "")
        _ = sut.handleIfUnauthorized(error)
        XCTAssertEqual(sut.rootRoute, .auth)
    }

    func test_handleIfUnauthorized_missingToken_setsRootRouteToAuth() {
        _ = sut.handleIfUnauthorized(NetworkError.missingToken)
        XCTAssertEqual(sut.rootRoute, .auth)
    }

    func test_handleIfUnauthorized_userParamsNotFound_setsRootRouteToAuth() {
        let error = NetworkError.requestFailed(statusCode: 404, message: "user params not found")
        _ = sut.handleIfUnauthorized(error)
        XCTAssertEqual(sut.rootRoute, .auth)
    }

    func test_handleIfUnauthorized_unauthorized_resetsSelectedTabToHome() {
        let error = NetworkError.requestFailed(statusCode: 401, message: "")
        _ = sut.handleIfUnauthorized(error)
        XCTAssertEqual(sut.selectedTab, .home)
    }

    func test_handleIfUnauthorized_unauthorized_clearsPendingAddMeal() {
        sut.pendingAddMeal = true
        let error = NetworkError.requestFailed(statusCode: 401, message: "")
        _ = sut.handleIfUnauthorized(error)
        XCTAssertFalse(sut.pendingAddMeal)
    }

    func test_handleIfUnauthorized_noInternet_doesNotChangeRootRoute() {
        sut.rootRoute = .main
        let error = NetworkError.network(URLError(.notConnectedToInternet))
        _ = sut.handleIfUnauthorized(error)
        XCTAssertEqual(sut.rootRoute, .main)
    }

    func test_handleIfUnauthorized_noInternet_doesNotResetSelectedTab() {
        sut.selectedTab = .stats
        let error = NetworkError.network(URLError(.notConnectedToInternet))
        _ = sut.handleIfUnauthorized(error)
        XCTAssertEqual(sut.selectedTab, .stats)
    }

    func test_handleIfUnauthorized_serverError_doesNotCallLogout() {
        sut.rootRoute = .main
        let error = NetworkError.requestFailed(statusCode: 500, message: "")
        _ = sut.handleIfUnauthorized(error)
        XCTAssertEqual(sut.rootRoute, .main)
    }

    func test_handleIfUnauthorized_profileUnauthorized_returnsTrue() {
        XCTAssertTrue(sut.handleIfUnauthorized(ProfileError.unauthorized))
    }

    func test_handleIfUnauthorized_profileUnauthorized_setsRootRouteToAuth() {
        _ = sut.handleIfUnauthorized(ProfileError.unauthorized)
        XCTAssertEqual(sut.rootRoute, .auth)
    }
}
