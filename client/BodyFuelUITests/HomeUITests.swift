import XCTest

final class HomeUITests: XCTestCase {

    var app: XCUIApplication!

    override func setUp() {
        continueAfterFailure = false
        app = XCUIApplication()
        app.launchEnvironment["UI_TESTING"] = "1"
        app.launchEnvironment["UI_TESTING_AUTH_RESULT"] = "success"

        addUIInterruptionMonitor(withDescription: "System dialog") { alert in
            let dismissLabels = ["Don't Allow", "Не разрешать", "Deny", "Запретить",
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

    // MARK: - Helpers

    private func launchAndLogin(hasTodayWorkout: Bool = false) {
        if hasTodayWorkout {
            app.launchEnvironment["UI_TESTING_HAS_TODAY_WORKOUT"] = "1"
        }
        app.launch()

        app.textFields["Логин"].tap()
        sleep(1)
        app.typeText("uitestuser")
        app.secureTextFields["Пароль"].tap()
        sleep(1)
        app.typeText("uitestpass")
        app.buttons["Войти"].tap()
        
        dismissSavePasswordPromptIfNeeded()

        XCTAssertTrue(app.tabBars.firstMatch.waitForExistence(timeout: 4))
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

    // MARK: - Calories ring

    func test_homeScreen_caloriesRing_isVisible() {
        launchAndLogin()
        XCTAssertTrue(app.otherElements["calories-ring"].waitForExistence(timeout: 3))
    }

    func test_homeScreen_caloriesRing_showsConsumedLabel() {
        launchAndLogin()
        XCTAssertTrue(app.staticTexts["Потреблено"].waitForExistence(timeout: 3))
    }

    func test_homeScreen_caloriesRing_showsBurnedLabel() {
        launchAndLogin()
        XCTAssertTrue(app.staticTexts["Сожжено"].waitForExistence(timeout: 3))
    }

    func test_homeScreen_caloriesRing_showsRemainingLabel() {
        launchAndLogin()
        XCTAssertTrue(app.staticTexts["Осталось"].waitForExistence(timeout: 3))
    }

    // MARK: - Workout card

    func test_homeScreen_workoutCard_showsSectionTitle() {
        launchAndLogin()
        XCTAssertTrue(app.staticTexts["Тренировка сегодня"].waitForExistence(timeout: 3))
    }

    func test_homeScreen_workoutCard_showsWorkoutTitle() {
        launchAndLogin()
        XCTAssertTrue(app.staticTexts["Тестовая тренировка"].waitForExistence(timeout: 3))
    }

    func test_homeScreen_workoutCard_showsStartButton() {
        launchAndLogin()
        XCTAssertTrue(app.buttons["Начать"].waitForExistence(timeout: 3))
    }

    func test_homeScreen_workoutCard_showsChangeButton() {
        launchAndLogin()
        XCTAssertTrue(app.buttons["Выбрать другую"].waitForExistence(timeout: 3))
    }

    // MARK: - Done block (hasTodayWorkout = true)

    func test_homeScreen_hasTodayWorkout_showsDoneBlock() {
        launchAndLogin(hasTodayWorkout: true)
        XCTAssertTrue(app.staticTexts["Отличная работа!"].waitForExistence(timeout: 3))
    }

    func test_homeScreen_hasTodayWorkout_hidesSectionTitle() {
        launchAndLogin(hasTodayWorkout: true)
        XCTAssertFalse(app.staticTexts["Тренировка сегодня"].waitForExistence(timeout: 2))
    }

    func test_homeScreen_hasTodayWorkout_showsReadyButton() {
        launchAndLogin(hasTodayWorkout: true)
        XCTAssertTrue(app.buttons["Готов к новому рекорду"].waitForExistence(timeout: 3))
    }

    func test_homeScreen_tapReadyForNewRecord_switchesToWorkoutCard() {
        launchAndLogin(hasTodayWorkout: true)
        XCTAssertTrue(app.buttons["Готов к новому рекорду"].waitForExistence(timeout: 3))

        app.buttons["Готов к новому рекорду"].tap()

        XCTAssertTrue(app.staticTexts["Тренировка сегодня"].waitForExistence(timeout: 3))
        XCTAssertFalse(app.staticTexts["Отличная работа!"].exists)
    }

    func test_homeScreen_tapReadyForNewRecord_showsStartButton() {
        launchAndLogin(hasTodayWorkout: true)
        XCTAssertTrue(app.buttons["Готов к новому рекорду"].waitForExistence(timeout: 3))

        app.buttons["Готов к новому рекорду"].tap()

        XCTAssertTrue(app.buttons["Начать"].waitForExistence(timeout: 3))
    }

    // MARK: - Nutrition card

    func test_homeScreen_nutritionCard_showsSectionTitle() {
        launchAndLogin()
        XCTAssertTrue(app.staticTexts["Питание сегодня"].waitForExistence(timeout: 3))
    }

    func test_homeScreen_nutritionCard_showsAddMealButton() {
        launchAndLogin()
        XCTAssertTrue(app.buttons["Добавить приём пищи"].waitForExistence(timeout: 3))
    }

    func test_homeScreen_addMealButton_switchesToFoodTab() {
        launchAndLogin()
        XCTAssertTrue(app.buttons["Добавить приём пищи"].waitForExistence(timeout: 3))

        app.buttons["Добавить приём пищи"].tap()

        XCTAssertTrue(app.tabBars.buttons["Питание"].waitForExistence(timeout: 2))
        let foodTab = app.tabBars.buttons["Питание"]
        XCTAssertTrue(foodTab.isSelected)
    }
}
