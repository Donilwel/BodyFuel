import XCTest
@testable import BodyFuel

final class ErrorMapperTests: XCTestCase {

    // MARK: - NetworkError.requestFailed

    func test_requestFailed_401_mapsToUnauthorized() {
        let error = NetworkError.requestFailed(statusCode: 401, message: "Unauthorized")
        XCTAssertEqual(ErrorMapper.map(error), .unauthorized)
    }

    func test_requestFailed_404_mapsToNotFound() {
        let error = NetworkError.requestFailed(statusCode: 404, message: "Not found")
        XCTAssertEqual(ErrorMapper.map(error), .notFound)
    }

    func test_requestFailed_400_mapsToValidation_withMessage() {
        let error = NetworkError.requestFailed(statusCode: 400, message: "some message")
        XCTAssertEqual(ErrorMapper.map(error), .validation(message: "some message"))
    }

    func test_requestFailed_400_preservesMessageInValidation() {
        let error = NetworkError.requestFailed(statusCode: 400, message: "Email already taken")
        if case .validation(let message) = ErrorMapper.map(error) {
            XCTAssertEqual(message, "Email already taken")
        } else {
            XCTFail("Expected .validation")
        }
    }

    func test_requestFailed_422_mapsToValidation() {
        let error = NetworkError.requestFailed(statusCode: 422, message: "Validation error")
        if case .validation = ErrorMapper.map(error) { } else {
            XCTFail("Expected .validation for 422")
        }
    }

    func test_requestFailed_499_mapsToValidation() {
        let error = NetworkError.requestFailed(statusCode: 499, message: "Client error")
        if case .validation = ErrorMapper.map(error) { } else {
            XCTFail("Expected .validation for 499")
        }
    }

    func test_requestFailed_500_mapsToServerUnavailable() {
        let error = NetworkError.requestFailed(statusCode: 500, message: "Internal server error")
        XCTAssertEqual(ErrorMapper.map(error), .serverUnavailable)
    }

    func test_requestFailed_503_mapsToServerUnavailable() {
        let error = NetworkError.requestFailed(statusCode: 503, message: "Service unavailable")
        XCTAssertEqual(ErrorMapper.map(error), .serverUnavailable)
    }

    func test_requestFailed_599_mapsToServerUnavailable() {
        let error = NetworkError.requestFailed(statusCode: 599, message: "Server error")
        XCTAssertEqual(ErrorMapper.map(error), .serverUnavailable)
    }

    func test_requestFailed_200_mapsToUnknown() {
        let error = NetworkError.requestFailed(statusCode: 200, message: "")
        XCTAssertEqual(ErrorMapper.map(error), .unknown)
    }

    func test_requestFailed_300_mapsToUnknown() {
        let error = NetworkError.requestFailed(statusCode: 300, message: "")
        XCTAssertEqual(ErrorMapper.map(error), .unknown)
    }

    // MARK: - NetworkError.missingToken

    func test_missingToken_mapsToUnauthorized() {
        let error = NetworkError.missingToken
        XCTAssertEqual(ErrorMapper.map(error), .unauthorized)
    }

    // MARK: - NetworkError.network

    func test_networkError_notConnectedToInternet_mapsToNoInternet() {
        let urlError = URLError(.notConnectedToInternet)
        let error = NetworkError.network(urlError)
        XCTAssertEqual(ErrorMapper.map(error), .noInternet)
    }

    func test_networkError_networkConnectionLost_mapsToNoInternet() {
        let urlError = URLError(.networkConnectionLost)
        let error = NetworkError.network(urlError)
        XCTAssertEqual(ErrorMapper.map(error), .noInternet)
    }

    func test_networkError_timedOut_mapsToUnknown() {
        let urlError = URLError(.timedOut)
        let error = NetworkError.network(urlError)
        XCTAssertEqual(ErrorMapper.map(error), .unknown)
    }

    func test_networkError_nonURLError_mapsToUnknown() {
        struct SomeError: Error {}
        let error = NetworkError.network(SomeError())
        XCTAssertEqual(ErrorMapper.map(error), .unknown)
    }

    // MARK: - NetworkError.decodingFailed / encodingFailed / invalidURL

    func test_decodingFailed_mapsToDecoding() {
        XCTAssertEqual(ErrorMapper.map(NetworkError.decodingFailed), .decoding)
    }

    func test_encodingFailed_mapsToEncoding() {
        XCTAssertEqual(ErrorMapper.map(NetworkError.encodingFailed), .encoding)
    }

    func test_invalidURL_mapsToServerUnavailable() {
        XCTAssertEqual(ErrorMapper.map(NetworkError.invalidURL), .serverUnavailable)
    }

    // MARK: - AuthError

    func test_authError_invalidCredentials_mapsToValidation() {
        if case .validation = ErrorMapper.map(AuthError.invalidCredentials) { } else {
            XCTFail("Expected .validation for invalidCredentials")
        }
    }

    func test_authError_validation_mapsToValidation() {
        if case .validation = ErrorMapper.map(AuthError.validation) { } else {
            XCTFail("Expected .validation for AuthError.validation")
        }
    }

    func test_authError_userExists_mapsToValidation() {
        if case .validation = ErrorMapper.map(AuthError.userExists) { } else {
            XCTFail("Expected .validation for userExists")
        }
    }

    func test_authError_invalidData_preservesMessage() {
        let error = AuthError.invalidData("Заполните все поля")
        if case .validation(let msg) = ErrorMapper.map(error) {
            XCTAssertEqual(msg, "Заполните все поля")
        } else {
            XCTFail("Expected .validation with message")
        }
    }

    // MARK: - ProfileError

    func test_profileError_validation_mapsToValidation() {
        if case .validation = ErrorMapper.map(ProfileError.validation) { } else {
            XCTFail("Expected .validation for ProfileError.validation")
        }
    }

    func test_profileError_unauthorized_mapsToUnauthorized() {
        XCTAssertEqual(ErrorMapper.map(ProfileError.unauthorized), .unauthorized)
    }

    func test_profileError_invalidData_preservesMessage() {
        let error = ProfileError.invalidData("Invalid weight")
        if case .validation(let msg) = ErrorMapper.map(error) {
            XCTAssertEqual(msg, "Invalid weight")
        } else {
            XCTFail("Expected .validation with message")
        }
    }

    // MARK: - HealthError

    func test_healthError_noPermission_mapsToValidation_withSettingsMessage() {
        if case .validation(let msg) = ErrorMapper.map(HealthError.noPermission) {
            XCTAssertFalse(msg.isEmpty)
        } else {
            XCTFail("Expected .validation for noPermission")
        }
    }

    func test_healthError_emptyValue_preservesMessage() {
        let error = HealthError.emptyValue(message: "No steps data")
        if case .validation(let msg) = ErrorMapper.map(error) {
            XCTAssertEqual(msg, "No steps data")
        } else {
            XCTFail("Expected .validation with message")
        }
    }

    // MARK: - Unknown error

    func test_unknownErrorType_mapsToUnknown() {
        struct RandomError: Error {}
        XCTAssertEqual(ErrorMapper.map(RandomError()), .unknown)
    }
}

// MARK: - NetworkErrorClassifierTests

final class NetworkErrorClassifierTests: XCTestCase {

    // MARK: - isUserParamsNotFoundError

    func test_isUserParamsNotFoundError_trueForExactMessage() {
        let error = NetworkError.requestFailed(statusCode: 404, message: "user params not found")
        XCTAssertTrue(isUserParamsNotFoundError(error))
    }

    func test_isUserParamsNotFoundError_caseInsensitive_uppercase() {
        let error = NetworkError.requestFailed(statusCode: 404, message: "User Params Not Found")
        XCTAssertTrue(isUserParamsNotFoundError(error))
    }

    func test_isUserParamsNotFoundError_caseInsensitive_mixed() {
        let error = NetworkError.requestFailed(statusCode: 404, message: "USER PARAMS NOT FOUND")
        XCTAssertTrue(isUserParamsNotFoundError(error))
    }

    func test_isUserParamsNotFoundError_trueWhenMessageContainsPhrase() {
        let error = NetworkError.requestFailed(statusCode: 404, message: "Error: user params not found for this user")
        XCTAssertTrue(isUserParamsNotFoundError(error))
    }

    func test_isUserParamsNotFoundError_falseForDifferentMessage() {
        let error = NetworkError.requestFailed(statusCode: 404, message: "not found")
        XCTAssertFalse(isUserParamsNotFoundError(error))
    }

    func test_isUserParamsNotFoundError_falseForEmptyMessage() {
        let error = NetworkError.requestFailed(statusCode: 404, message: "")
        XCTAssertFalse(isUserParamsNotFoundError(error))
    }

    func test_isUserParamsNotFoundError_falseForMissingToken() {
        XCTAssertFalse(isUserParamsNotFoundError(NetworkError.missingToken))
    }

    func test_isUserParamsNotFoundError_falseForNonNetworkError() {
        struct OtherError: Error {}
        XCTAssertFalse(isUserParamsNotFoundError(OtherError()))
    }

    func test_isUserParamsNotFoundError_falseForNetworkConnectionError() {
        let error = NetworkError.network(URLError(.notConnectedToInternet))
        XCTAssertFalse(isUserParamsNotFoundError(error))
    }

    // MARK: - isAuthError

    func test_isAuthError_trueForMissingToken() {
        XCTAssertTrue(isAuthError(NetworkError.missingToken))
    }

    func test_isAuthError_falseForRequestFailed401() {
        XCTAssertFalse(isAuthError(NetworkError.requestFailed(statusCode: 401, message: "")))
    }

    func test_isAuthError_falseForDecodingFailed() {
        XCTAssertFalse(isAuthError(NetworkError.decodingFailed))
    }

    func test_isAuthError_falseForNonNetworkError() {
        struct OtherError: Error {}
        XCTAssertFalse(isAuthError(OtherError()))
    }

    // MARK: - isTransportError

    func test_isTransportError_trueForNetworkURLError() {
        let error = NetworkError.network(URLError(.notConnectedToInternet))
        XCTAssertTrue(isTransportError(error))
    }

    func test_isTransportError_falseForRequestFailed() {
        XCTAssertFalse(isTransportError(NetworkError.requestFailed(statusCode: 500, message: "")))
    }

    func test_isTransportError_falseForMissingToken() {
        XCTAssertFalse(isTransportError(NetworkError.missingToken))
    }

    func test_isTransportError_trueForNonNetworkError() {
        struct OtherError: Error {}
        XCTAssertTrue(isTransportError(OtherError()))
    }

    func test_isTransportError_falseForDecodingFailed() {
        XCTAssertFalse(isTransportError(NetworkError.decodingFailed))
    }
}
