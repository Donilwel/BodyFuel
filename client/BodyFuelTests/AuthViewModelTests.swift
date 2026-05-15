import XCTest
@testable import BodyFuel

@MainActor
final class AuthViewModelTests: XCTestCase {

    var mockAuthService: MockAuthService!
    var sut: AuthViewModel!

    override func setUp() async throws {
        mockAuthService = MockAuthService()
        sut = AuthViewModel(authService: mockAuthService)
    }

    override func tearDown() async throws {
        sut = nil
        mockAuthService = nil
    }

    private func setValidLoginFields() {
        sut.mode = .login
        sut.login = "testuser"
        sut.password = "password123"
    }

    private func setValidRegisterFields() {
        sut.mode = .register
        sut.login = "testuser"
        sut.name = "Ivan"
        sut.surname = "Ivanov"
        sut.email = "test@example.com"
        sut.phone = "79991234567"
        sut.password = "password123"
        sut.confirmPassword = "password123"
    }
    
    // MARK: - initialState

    func test_initialState_modeIsLogin() {
        XCTAssertEqual(sut.mode, .login)
    }

    func test_initialState_eventIsNil() {
        XCTAssertNil(sut.event)
    }

    func test_initialState_screenStateIsIdle() {
        XCTAssertEqual(sut.screenState, .idle)
    }
    
    // MARK: - submit()

    func test_submit_loginMode_callsAuthServiceLogin() async throws {
        setValidLoginFields()

        await sut.submit()

        XCTAssertEqual(mockAuthService.loginCallCount, 1)
        XCTAssertEqual(mockAuthService.registerCallCount, 0)
    }

    func test_submit_loginMode_sendsCorrectPayload() async throws {
        setValidLoginFields()
        sut.login = "myuser"
        sut.password = "secret123"

        await sut.submit()

        XCTAssertEqual(mockAuthService.lastLoginPayload?.username, "myuser")
        XCTAssertEqual(mockAuthService.lastLoginPayload?.password, "secret123")
    }

    func test_submit_loginMode_onSuccess_setsLoginSuccessEvent() async throws {
        setValidLoginFields()

        await sut.submit()

        XCTAssertEqual(sut.event, .loginSuccess)
    }

    func test_submit_loginMode_setsScreenStateToIdle_afterSuccess() async throws {
        setValidLoginFields()

        await sut.submit()

        XCTAssertEqual(sut.screenState, .idle)
    }

    func test_submit_loginMode_doesNotSetRegisterEvent() async throws {
        setValidLoginFields()

        await sut.submit()

        XCTAssertNotEqual(sut.event, .registrationSuccess)
    }

    func test_submit_registerMode_callsRegisterThenLogin() async throws {
        setValidRegisterFields()

        await sut.submit()

        XCTAssertEqual(mockAuthService.registerCallCount, 1)
        XCTAssertEqual(mockAuthService.loginCallCount, 1)
    }

    func test_submit_registerMode_sendsCorrectRegisterPayload() async throws {
        setValidRegisterFields()
        sut.login = "newuser"
        sut.name = "Maria"
        sut.surname = "Petrova"
        sut.email = "maria@example.com"
        sut.phone = "79001112233"
        sut.password = "secure123"
        sut.confirmPassword = "secure123"

        await sut.submit()

        let payload = mockAuthService.lastRegisterPayload
        XCTAssertEqual(payload?.username, "newuser")
        XCTAssertEqual(payload?.name, "Maria")
        XCTAssertEqual(payload?.surname, "Petrova")
        XCTAssertEqual(payload?.email, "maria@example.com")
    }

    func test_submit_registerMode_stripsPhoneFormatting() async throws {
        setValidRegisterFields()
        sut.phone = "7 (999) 123-45-67"
        sut.confirmPassword = sut.password

        await sut.submit()

        XCTAssertEqual(mockAuthService.lastRegisterPayload?.phone, "79991234567")
    }

    func test_submit_registerMode_autoLoginUsesRegisteredCredentials() async throws {
        setValidRegisterFields()
        sut.login = "newuser"
        sut.password = "mypassword"
        sut.confirmPassword = "mypassword"

        await sut.submit()

        XCTAssertEqual(mockAuthService.lastLoginPayload?.username, "newuser")
        XCTAssertEqual(mockAuthService.lastLoginPayload?.password, "mypassword")
    }

    func test_submit_registerMode_onSuccess_setsRegistrationSuccessEvent() async throws {
        setValidRegisterFields()

        await sut.submit()

        XCTAssertEqual(sut.event, .registrationSuccess)
    }

    func test_submit_registerMode_loginNotCalledIfRegisterFails() async throws {
        setValidRegisterFields()
        mockAuthService.registerResult = .failure(NetworkError.requestFailed(statusCode: 409, message: "User exists"))

        await sut.submit()

        XCTAssertEqual(mockAuthService.registerCallCount, 1)
        XCTAssertEqual(mockAuthService.loginCallCount, 0)
    }

    func test_submit_loginServiceError_setsErrorState() async throws {
        setValidLoginFields()
        mockAuthService.loginResult = .failure(NetworkError.requestFailed(statusCode: 401, message: "Unauthorized"))

        await sut.submit()

        if case .error(let msg) = sut.screenState {
            XCTAssertFalse(msg.isEmpty)
        } else {
            XCTFail("Expected .error state, got \(sut.screenState)")
        }
    }

    func test_submit_loginServiceError_doesNotSetEvent() async throws {
        setValidLoginFields()
        mockAuthService.loginResult = .failure(NetworkError.requestFailed(statusCode: 401, message: ""))

        await sut.submit()

        XCTAssertNil(sut.event)
    }

    func test_submit_registerServiceError_setsErrorState() async throws {
        setValidRegisterFields()
        mockAuthService.registerResult = .failure(NetworkError.requestFailed(statusCode: 409, message: "User exists"))

        await sut.submit()

        if case .error = sut.screenState { } else {
            XCTFail("Expected .error state")
        }
    }

    func test_submit_loginMode_emptyLogin_setsErrorState() async throws {
        sut.mode = .login
        sut.login = ""
        sut.password = "password123"

        await sut.submit()

        if case .error = sut.screenState { } else {
            XCTFail("Expected .error state from empty login")
        }
        XCTAssertEqual(mockAuthService.loginCallCount, 0)
    }

    func test_submit_loginMode_emptyPassword_setsErrorState_doesNotCallService() async throws {
        sut.mode = .login
        sut.login = "user"
        sut.password = ""

        await sut.submit()

        if case .error = sut.screenState { } else {
            XCTFail("Expected .error state from empty password")
        }
        XCTAssertEqual(mockAuthService.loginCallCount, 0)
    }

    func test_submit_registerMode_invalidEmail_setsErrorState_doesNotCallService() async throws {
        setValidRegisterFields()
        sut.email = "not-an-email"

        await sut.submit()

        if case .error = sut.screenState { } else {
            XCTFail("Expected .error state from invalid email")
        }
        XCTAssertEqual(mockAuthService.registerCallCount, 0)
    }

    func test_submit_registerMode_passwordMismatch_setsErrorState() async throws {
        setValidRegisterFields()
        sut.confirmPassword = "differentpassword"

        await sut.submit()

        if case .error = sut.screenState { } else {
            XCTFail("Expected .error state from password mismatch")
        }
        XCTAssertEqual(mockAuthService.registerCallCount, 0)
    }
    
    // MARK: - validateLive()

    func test_validateLive_emptyPassword_setsPasswordError() {
        sut.password = ""
        sut.validateLive()

        XCTAssertNotNil(sut.passwordError)
    }

    func test_validateLive_shortPassword_setsPasswordError() {
        sut.password = "abc"
        sut.validateLive()

        XCTAssertNotNil(sut.passwordError)
    }

    func test_validateLive_validPassword_clearsPasswordError() {
        sut.password = "password123"
        sut.validateLive()

        XCTAssertNil(sut.passwordError)
    }

    func test_validateLive_exactSixChars_isValidPassword() {
        sut.password = "abcdef"
        sut.validateLive()

        XCTAssertNil(sut.passwordError)
    }

    func test_validateLive_passwordMatch_clearsConfirmPasswordError() {
        sut.password = "password123"
        sut.confirmPassword = "password123"
        sut.validateLive()

        XCTAssertNil(sut.confirmPasswordError)
    }

    func test_validateLive_passwordMismatch_setsConfirmPasswordError() {
        sut.password = "password123"
        sut.confirmPassword = "different"
        sut.validateLive()

        XCTAssertNotNil(sut.confirmPasswordError)
    }

    func test_validateLive_validEmail_clearsEmailError() {
        sut.mode = .register
        sut.email = "user@example.com"
        sut.validateLive()

        XCTAssertNil(sut.emailError)
    }

    func test_validateLive_emptyEmail_setsEmailError() {
        sut.mode = .register
        sut.email = ""
        sut.validateLive()

        XCTAssertNotNil(sut.emailError)
    }

    func test_validateLive_invalidEmail_noAt_setsEmailError() {
        sut.mode = .register
        sut.email = "notanemail"
        sut.validateLive()

        XCTAssertNotNil(sut.emailError)
    }

    func test_validateLive_invalidEmail_noTld_setsEmailError() {
        sut.mode = .register
        sut.email = "user@example"
        sut.validateLive()

        XCTAssertNotNil(sut.emailError)
    }

    func test_validateLive_emailNotValidatedInLoginMode() {
        sut.mode = .login
        sut.email = "notanemail"
        sut.validateLive()

        XCTAssertNil(sut.emailError)
    }

    func test_validateLive_validEmailSubdomains_clearsEmailError() {
        sut.mode = .register
        sut.email = "user.name+tag@sub.example.co.uk"
        sut.validateLive()

        XCTAssertNil(sut.emailError)
    }

    func test_validateLive_validPhone_digits_clearsPhoneError() {
        sut.mode = .register
        sut.phone = "79991234567"
        sut.validateLive()

        XCTAssertNil(sut.phoneError)
    }

    func test_validateLive_validPhone_formatted_clearsPhoneError() {
        sut.mode = .register
        sut.phone = "+7 (999) 123-45-67"
        sut.validateLive()

        XCTAssertNil(sut.phoneError)
    }

    func test_validateLive_emptyPhone_setsPhoneError() {
        sut.mode = .register
        sut.phone = ""
        sut.validateLive()

        XCTAssertNotNil(sut.phoneError)
    }

    func test_validateLive_phoneStartingWith8_setsPhoneError() {
        sut.mode = .register
        sut.phone = "89991234567"
        sut.validateLive()

        XCTAssertNotNil(sut.phoneError)
    }

    func test_validateLive_phoneTooShort_setsPhoneError() {
        sut.mode = .register
        sut.phone = "7999123456"
        sut.validateLive()

        XCTAssertNotNil(sut.phoneError)
    }

    func test_validateLive_phoneTooLong_setsPhoneError() {
        sut.mode = .register
        sut.phone = "799912345678"
        sut.validateLive()

        XCTAssertNotNil(sut.phoneError)
    }

    func test_validateLive_phoneNotValidatedInLoginMode() {
        sut.mode = .login
        sut.phone = "invalid"
        sut.validateLive()

        XCTAssertNil(sut.phoneError)
    }
    
    // MARK: - isRegisterFormComplete

    func test_isRegisterFormComplete_trueInLoginMode_always() {
        sut.mode = .login
        XCTAssertTrue(sut.isRegisterFormComplete)
    }

    func test_isRegisterFormComplete_true_whenAllFieldsValid() {
        sut.mode = .register
        sut.login = "user"
        sut.name = "Ivan"
        sut.surname = "Ivanov"
        sut.email = "user@example.com"
        sut.emailError = nil
        sut.phone = "79991234567"
        sut.phoneError = nil
        sut.password = "password123"
        sut.passwordError = nil
        sut.confirmPassword = "password123"
        sut.confirmPasswordError = nil

        XCTAssertTrue(sut.isRegisterFormComplete)
    }

    func test_isRegisterFormComplete_false_whenLoginEmpty() {
        sut.mode = .register
        setValidRegisterFields()
        sut.login = ""

        XCTAssertFalse(sut.isRegisterFormComplete)
    }

    func test_isRegisterFormComplete_false_whenNameEmpty() {
        sut.mode = .register
        setValidRegisterFields()
        sut.name = ""

        XCTAssertFalse(sut.isRegisterFormComplete)
    }

    func test_isRegisterFormComplete_false_whenSurnameEmpty() {
        sut.mode = .register
        setValidRegisterFields()
        sut.surname = ""

        XCTAssertFalse(sut.isRegisterFormComplete)
    }

    func test_isRegisterFormComplete_false_whenEmailEmpty() {
        sut.mode = .register
        setValidRegisterFields()
        sut.email = ""

        XCTAssertFalse(sut.isRegisterFormComplete)
    }

    func test_isRegisterFormComplete_false_whenEmailErrorPresent() {
        sut.mode = .register
        setValidRegisterFields()
        sut.emailError = "Некорректный email"

        XCTAssertFalse(sut.isRegisterFormComplete)
    }

    func test_isRegisterFormComplete_false_whenPhoneEmpty() {
        sut.mode = .register
        setValidRegisterFields()
        sut.phone = ""

        XCTAssertFalse(sut.isRegisterFormComplete)
    }

    func test_isRegisterFormComplete_false_whenPhoneErrorPresent() {
        sut.mode = .register
        setValidRegisterFields()
        sut.phoneError = "Некорректный телефон"

        XCTAssertFalse(sut.isRegisterFormComplete)
    }

    func test_isRegisterFormComplete_false_whenPasswordEmpty() {
        sut.mode = .register
        setValidRegisterFields()
        sut.password = ""

        XCTAssertFalse(sut.isRegisterFormComplete)
    }

    func test_isRegisterFormComplete_false_whenPasswordErrorPresent() {
        sut.mode = .register
        setValidRegisterFields()
        sut.passwordError = "Минимум 6 символов"

        XCTAssertFalse(sut.isRegisterFormComplete)
    }

    func test_isRegisterFormComplete_false_whenConfirmPasswordEmpty() {
        sut.mode = .register
        setValidRegisterFields()
        sut.confirmPassword = ""

        XCTAssertFalse(sut.isRegisterFormComplete)
    }

    func test_isRegisterFormComplete_false_whenConfirmPasswordErrorPresent() {
        sut.mode = .register
        setValidRegisterFields()
        sut.confirmPasswordError = "Пароли не совпадают"

        XCTAssertFalse(sut.isRegisterFormComplete)
    }
    
    // MARK: - submit()

    func test_submit_setsPasswordError_forShortPassword() async throws {
        sut.mode = .login
        sut.login = "user"
        sut.password = "abc"

        await sut.submit()

        XCTAssertNotNil(sut.passwordError)
    }

    func test_submit_registerMode_setsEmailError_forInvalidEmail() async throws {
        setValidRegisterFields()
        sut.email = "bad-email"

        await sut.submit()

        XCTAssertNotNil(sut.emailError)
    }
}
