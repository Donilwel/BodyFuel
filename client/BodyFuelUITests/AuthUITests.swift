import XCTest

final class AuthUITests: XCTestCase {

    var app: XCUIApplication!

    override func setUp() {
        continueAfterFailure = false
        app = XCUIApplication()
        app.launchEnvironment["UI_TESTING"] = "1"

        addUIInterruptionMonitor(withDescription: "System dialog") { alert in
            let dismissLabels = ["Don't Allow", "Не разрешать", "Deny", "Отказать",
                                 "Not Now", "Не сейчас", "Never", "Никогда"]
            for label in dismissLabels {
                if alert.buttons[label].exists {
                    alert.buttons[label].tap()
                    return true
                }
            }
            alert.buttons.firstMatch.tap()
            return true
        }
    }

    override func tearDown() {
        app = nil
    }

    private func dismissSavePasswordPromptIfNeeded() {
        let dismissLabels = ["Not Now", "Не сейчас", "Never", "Никогда"]
        for label in dismissLabels {
            let button = app.buttons[label]
            if button.waitForExistence(timeout: 1) {
                button.tap()
                return
            }
        }
    }

    private func launch(authResult: String = "success", recoveryResult: String = "success") {
        app.launchEnvironment["UI_TESTING_AUTH_RESULT"] = authResult
        app.launchEnvironment["UI_TESTING_RECOVERY_RESULT"] = recoveryResult
        app.launch()
    }

    // MARK: - Initial state

    func test_authScreen_segmentedControl_visibleOnLaunch() {
        launch()
        XCTAssertTrue(app.segmentedControls.firstMatch.exists)
    }

    func test_authScreen_loginButton_existsAndEnabled() {
        launch()
        XCTAssertTrue(app.buttons["Войти"].exists)
        XCTAssertTrue(app.buttons["Войти"].isEnabled)
    }

    func test_authScreen_loginFields_visible() {
        launch()
        XCTAssertTrue(app.textFields["Логин"].exists)
        XCTAssertTrue(app.secureTextFields["Пароль"].exists)
    }

    func test_authScreen_forgotPasswordLink_visibleInLoginMode() {
        launch()
        XCTAssertTrue(app.buttons["Забыли пароль?"].exists)
    }

    func test_authScreen_registerButton_notVisibleInLoginMode() {
        launch()
        XCTAssertFalse(app.buttons["Зарегистрироваться"].exists)
    }

    // MARK: - Mode switching

    func test_switchToRegister_showsRegisterForm() {
        launch()

        app.segmentedControls.firstMatch.buttons["Регистрация"].tap()

        XCTAssertTrue(app.buttons["Зарегистрироваться"].waitForExistence(timeout: 2))
        XCTAssertTrue(app.textFields["Имя"].exists)
        XCTAssertTrue(app.textFields["Фамилия"].exists)
        XCTAssertTrue(app.textFields["Почта"].exists)
    }

    func test_switchToRegister_hidesLoginButton() {
        launch()

        app.segmentedControls.firstMatch.buttons["Регистрация"].tap()

        XCTAssertFalse(app.buttons["Войти"].waitForExistence(timeout: 1))
    }

    func test_switchToRegister_hidesForgotPasswordLink() {
        launch()

        app.segmentedControls.firstMatch.buttons["Регистрация"].tap()

        XCTAssertFalse(app.buttons["Забыли пароль?"].waitForExistence(timeout: 1))
    }

    func test_switchBackToLogin_showsLoginForm() {
        launch()

        app.segmentedControls.firstMatch.buttons["Регистрация"].tap()
        XCTAssertTrue(app.buttons["Зарегистрироваться"].waitForExistence(timeout: 2))

        app.segmentedControls.firstMatch.buttons["Вход"].tap()

        XCTAssertTrue(app.buttons["Войти"].waitForExistence(timeout: 2))
        XCTAssertFalse(app.buttons["Зарегистрироваться"].exists)
    }

    func test_switchBackToLogin_restoresForgotPasswordLink() {
        launch()

        app.segmentedControls.firstMatch.buttons["Регистрация"].tap()
        app.segmentedControls.firstMatch.buttons["Вход"].tap()

        XCTAssertTrue(app.buttons["Забыли пароль?"].waitForExistence(timeout: 2))
    }

    // MARK: - Submit button state

    func test_registerButton_disabledWithEmptyFields() {
        launch()

        app.segmentedControls.firstMatch.buttons["Регистрация"].tap()
        let button = app.buttons["Зарегистрироваться"]
        XCTAssertTrue(button.waitForExistence(timeout: 2))

        XCTAssertFalse(button.isEnabled)
    }

    func test_loginButton_alwaysEnabledInLoginMode() {
        launch()
        XCTAssertTrue(app.buttons["Войти"].isEnabled)
    }

    // MARK: - Login

    func test_login_emptyFields_showsErrorAlert() {
        launch()

        app.buttons["Войти"].tap()

        XCTAssertTrue(app.alerts["Ошибка"].waitForExistence(timeout: 3))
    }

    func test_login_emptyFields_alertDismissedOnOK() {
        launch()

        app.buttons["Войти"].tap()
        XCTAssertTrue(app.alerts["Ошибка"].waitForExistence(timeout: 3))

        app.alerts["Ошибка"].buttons["OK"].tap()

        XCTAssertFalse(app.alerts.firstMatch.exists)
        XCTAssertTrue(app.segmentedControls.firstMatch.exists)
    }

    func test_login_serverError_showsErrorAlert() {
        launch(authResult: "error")

        app.textFields["Логин"].tap()
        app.textFields["Логин"].typeText("wronguser")
        app.secureTextFields["Пароль"].tap()
        app.secureTextFields["Пароль"].typeText("wrongpass")

        app.buttons["Войти"].tap()

        XCTAssertTrue(app.alerts["Ошибка"].waitForExistence(timeout: 3))
    }

    func test_login_serverError_remainsOnAuthScreen() {
        launch(authResult: "error")

        app.textFields["Логин"].tap()
        app.textFields["Логин"].typeText("wronguser")
        app.secureTextFields["Пароль"].tap()
        app.secureTextFields["Пароль"].typeText("wrongpass")

        app.buttons["Войти"].tap()
        XCTAssertTrue(app.alerts["Ошибка"].waitForExistence(timeout: 3))
        app.alerts["Ошибка"].buttons["OK"].tap()

        XCTAssertTrue(app.segmentedControls.firstMatch.exists)
        XCTAssertFalse(app.tabBars.firstMatch.exists)
    }

    func test_login_success_navigatesToMainScreen() {
        launch(authResult: "success")

        app.textFields["Логин"].tap()
        sleep(1)
        app.textFields["Логин"].typeText("uitestuser")
        app.secureTextFields["Пароль"].tap()
        sleep(1)
        app.secureTextFields["Пароль"].typeText("uitestpass")

        app.buttons["Войти"].tap()

        XCTAssertTrue(app.tabBars.firstMatch.waitForExistence(timeout: 4))
        XCTAssertTrue(app.tabBars.buttons["Главный экран"].exists)
    }

    func test_login_success_tabBar_hasAllTabs() {
        launch(authResult: "success")

        app.textFields["Логин"].tap()
        app.textFields["Логин"].typeText("uitestuser")
        app.secureTextFields["Пароль"].tap()
        app.secureTextFields["Пароль"].typeText("pass123")

        app.buttons["Войти"].tap()
        XCTAssertTrue(app.tabBars.firstMatch.waitForExistence(timeout: 4))

        XCTAssertTrue(app.tabBars.buttons["Главный экран"].exists)
        XCTAssertTrue(app.tabBars.buttons["Питание"].exists)
        XCTAssertTrue(app.tabBars.buttons["Статистика"].exists)
        XCTAssertTrue(app.tabBars.buttons["Профиль"].exists)
    }

    // MARK: - Password recovery

    func test_forgotPassword_navigatesToRecoveryScreen() {
        launch()

        app.buttons["Забыли пароль?"].tap()

        XCTAssertTrue(app.staticTexts["Восстановление пароля"].waitForExistence(timeout: 2))
    }

    func test_recovery_emailStep_hasEmailFieldAndSendButton() {
        launch()

        app.buttons["Забыли пароль?"].tap()
        XCTAssertTrue(app.staticTexts["Восстановление пароля"].waitForExistence(timeout: 2))

        XCTAssertTrue(app.textFields["Email"].exists)
        XCTAssertTrue(app.buttons["Отправить код"].exists)
    }

    func test_recovery_emptyEmail_showsError() {
        launch()

        app.buttons["Забыли пароль?"].tap()
        XCTAssertTrue(app.textFields["Email"].waitForExistence(timeout: 2))

        app.buttons["Отправить код"].tap()

        XCTAssertTrue(app.alerts["Ошибка"].waitForExistence(timeout: 2))
    }

    func test_recovery_serverError_onEmailStep_showsError() {
        launch(recoveryResult: "error")

        app.buttons["Забыли пароль?"].tap()
        XCTAssertTrue(app.textFields["Email"].waitForExistence(timeout: 2))

        app.textFields["Email"].tap()
        app.textFields["Email"].typeText("unknown@example.com")
        app.buttons["Отправить код"].tap()

        XCTAssertTrue(app.alerts["Ошибка"].waitForExistence(timeout: 3))
    }

    func test_recovery_validEmail_advancesToCodeStep() {
        launch()

        app.buttons["Забыли пароль?"].tap()
        XCTAssertTrue(app.textFields["Email"].waitForExistence(timeout: 2))

        app.textFields["Email"].tap()
        app.textFields["Email"].typeText("user@example.com")
        app.buttons["Отправить код"].tap()

        XCTAssertTrue(app.textFields["Код из письма"].waitForExistence(timeout: 3))
        XCTAssertTrue(app.secureTextFields["Новый пароль"].exists)
        XCTAssertTrue(app.buttons["Сменить пароль"].exists)
        XCTAssertFalse(app.buttons["Отправить код"].exists)
    }

    func test_recovery_fullFlow_showsSuccessScreen() {
        launch()

        app.buttons["Забыли пароль?"].tap()
        XCTAssertTrue(app.textFields["Email"].waitForExistence(timeout: 2))

        app.textFields["Email"].tap()
        app.textFields["Email"].typeText("user@example.com")
        app.buttons["Отправить код"].tap()

        XCTAssertTrue(app.textFields["Код из письма"].waitForExistence(timeout: 3))
        app.textFields["Код из письма"].tap()
        sleep(1)
        app.textFields["Код из письма"].typeText("123456")
        app.secureTextFields["Новый пароль"].tap()
        sleep(1)
        app.secureTextFields["Новый пароль"].typeText("newpassword123")
        app.swipeDown()
        app.buttons["Сменить пароль"].tap()
        dismissSavePasswordPromptIfNeeded()

        XCTAssertTrue(app.staticTexts["Пароль успешно изменён"].waitForExistence(timeout: 3))
    }

    func test_recovery_successScreen_hasLoginButton() {
        launch()

        app.buttons["Забыли пароль?"].tap()
        XCTAssertTrue(app.textFields["Email"].waitForExistence(timeout: 2))
        app.textFields["Email"].tap()
        app.textFields["Email"].typeText("user@example.com")
        app.buttons["Отправить код"].tap()
        XCTAssertTrue(app.textFields["Код из письма"].waitForExistence(timeout: 3))
        app.textFields["Код из письма"].tap()
        sleep(1)
        app.textFields["Код из письма"].typeText("000000")
        app.secureTextFields["Новый пароль"].tap()
        sleep(1)
        app.secureTextFields["Новый пароль"].typeText("newpass123")
        app.swipeDown()
        app.buttons["Сменить пароль"].tap()
        dismissSavePasswordPromptIfNeeded()
        XCTAssertTrue(app.staticTexts["Пароль успешно изменён"].waitForExistence(timeout: 3))

        XCTAssertTrue(app.buttons["Войти"].exists)
    }

    func test_recovery_successScreen_loginReturnsToAuth() {
        launch()

        app.buttons["Забыли пароль?"].tap()
        XCTAssertTrue(app.textFields["Email"].waitForExistence(timeout: 2))
        app.textFields["Email"].tap()
        app.textFields["Email"].typeText("user@example.com")
        app.buttons["Отправить код"].tap()
        XCTAssertTrue(app.textFields["Код из письма"].waitForExistence(timeout: 3))
        app.textFields["Код из письма"].tap()
        sleep(1)
        app.textFields["Код из письма"].typeText("000000")
        app.secureTextFields["Новый пароль"].tap()
        sleep(1)
        app.secureTextFields["Новый пароль"].typeText("newpass123")
        app.swipeDown()
        app.buttons["Сменить пароль"].tap()
        dismissSavePasswordPromptIfNeeded()
        XCTAssertTrue(app.staticTexts["Пароль успешно изменён"].waitForExistence(timeout: 3))

        app.buttons["Войти"].tap()

        XCTAssertTrue(app.segmentedControls.firstMatch.waitForExistence(timeout: 2))
    }

    func test_recovery_serverErrorOnCodeStep_showsError() {
        launch(recoveryResult: "error")

        app.buttons["Забыли пароль?"].tap()
        XCTAssertTrue(app.textFields["Email"].waitForExistence(timeout: 2))
        app.textFields["Email"].tap()
        app.textFields["Email"].typeText("user@example.com")
        app.buttons["Отправить код"].tap()

        XCTAssertTrue(app.alerts["Ошибка"].waitForExistence(timeout: 3))
    }
}
