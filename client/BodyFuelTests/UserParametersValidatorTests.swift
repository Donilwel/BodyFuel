import XCTest
@testable import BodyFuel

final class UserParametersValidatorTests: XCTestCase {

    let sut = UserParametersValidator.shared

    // MARK: - validateHeight

    func test_validateHeight_100_isValid() {
        XCTAssertNil(sut.validateHeight(100))
    }

    func test_validateHeight_99_isInvalid() {
        XCTAssertNotNil(sut.validateHeight(99))
    }

    func test_validateHeight_250_isValid() {
        XCTAssertNil(sut.validateHeight(250))
    }

    func test_validateHeight_251_isInvalid() {
        XCTAssertNotNil(sut.validateHeight(251))
    }

    func test_validateHeight_typicalValue_isValid() {
        XCTAssertNil(sut.validateHeight(175))
    }

    func test_validateHeight_zero_isInvalid() {
        XCTAssertNotNil(sut.validateHeight(0))
    }

    func test_validateHeight_negative_isInvalid() {
        XCTAssertNotNil(sut.validateHeight(-10))
    }

    func test_validateHeight_returnsNonEmptyMessage_onError() {
        let error = sut.validateHeight(50)
        XCTAssertFalse(error?.isEmpty ?? true)
    }

    // MARK: - validateWeight

    func test_validateWeight_40_isInvalid() {
        XCTAssertNotNil(sut.validateWeight(40.0))
    }

    func test_validateWeight_40point1_isValid() {
        XCTAssertNil(sut.validateWeight(40.1))
    }

    func test_validateWeight_300_isInvalid() {
        XCTAssertNotNil(sut.validateWeight(300.0))
    }

    func test_validateWeight_299point9_isValid() {
        XCTAssertNil(sut.validateWeight(299.9))
    }

    func test_validateWeight_typicalValue_isValid() {
        XCTAssertNil(sut.validateWeight(70.0))
    }

    func test_validateWeight_zero_isInvalid() {
        XCTAssertNotNil(sut.validateWeight(0))
    }

    func test_validateWeight_negative_isInvalid() {
        XCTAssertNotNil(sut.validateWeight(-5))
    }

    func test_validateWeight_returnsNonEmptyMessage_onError() {
        let error = sut.validateWeight(10)
        XCTAssertFalse(error?.isEmpty ?? true)
    }

    // MARK: - validateTargetWeight — loseWeight

    func test_validateTargetWeight_loseWeight_targetLessThanCurrent_isValid() {
        XCTAssertNil(sut.validateTargetWeight(65, weight: 80, goal: .loseWeight))
    }

    func test_validateTargetWeight_loseWeight_targetEqualsCurrent_isInvalid() {
        XCTAssertNotNil(sut.validateTargetWeight(70, weight: 70, goal: .loseWeight))
    }

    func test_validateTargetWeight_loseWeight_targetGreaterThanCurrent_isInvalid() {
        XCTAssertNotNil(sut.validateTargetWeight(80, weight: 70, goal: .loseWeight))
    }

    func test_validateTargetWeight_loseWeight_justBelow_isValid() {
        XCTAssertNil(sut.validateTargetWeight(69.9, weight: 70, goal: .loseWeight))
    }

    func test_validateTargetWeight_loseWeight_errorMessageContainsGoalContext() {
        let error = sut.validateTargetWeight(80, weight: 70, goal: .loseWeight)
        XCTAssertNotNil(error)
        XCTAssertFalse(error?.isEmpty ?? true)
    }

    // MARK: - validateTargetWeight — gainMuscle

    func test_validateTargetWeight_gainMuscle_targetGreaterThanCurrent_isValid() {
        XCTAssertNil(sut.validateTargetWeight(80, weight: 70, goal: .gainMuscle))
    }

    func test_validateTargetWeight_gainMuscle_targetEqualsCurrent_isInvalid() {
        XCTAssertNotNil(sut.validateTargetWeight(70, weight: 70, goal: .gainMuscle))
    }

    func test_validateTargetWeight_gainMuscle_targetLessThanCurrent_isInvalid() {
        XCTAssertNotNil(sut.validateTargetWeight(60, weight: 70, goal: .gainMuscle))
    }

    func test_validateTargetWeight_gainMuscle_justAbove_isValid() {
        XCTAssertNil(sut.validateTargetWeight(70.1, weight: 70, goal: .gainMuscle))
    }

    func test_validateTargetWeight_gainMuscle_errorMessageContainsGoalContext() {
        let error = sut.validateTargetWeight(60, weight: 70, goal: .gainMuscle)
        XCTAssertNotNil(error)
        XCTAssertFalse(error?.isEmpty ?? true)
    }

    // MARK: - validateTargetWeight — maintain

    func test_validateTargetWeight_maintain_targetEqualsCurrent_isValid() {
        XCTAssertNil(sut.validateTargetWeight(70, weight: 70, goal: .maintain))
    }

    func test_validateTargetWeight_maintain_targetAboveCurrent_isInvalid() {
        XCTAssertNotNil(sut.validateTargetWeight(71, weight: 70, goal: .maintain))
    }

    func test_validateTargetWeight_maintain_targetBelowCurrent_isInvalid() {
        XCTAssertNotNil(sut.validateTargetWeight(69, weight: 70, goal: .maintain))
    }

    // MARK: - validateGoal

    func test_validateGoal_loseWeight_currentGreaterThanTarget_isValid() {
        XCTAssertNil(sut.validateGoal(.loseWeight, weight: 80, targetWeight: 70))
    }

    func test_validateGoal_loseWeight_currentEqualsTarget_isInvalid() {
        XCTAssertNotNil(sut.validateGoal(.loseWeight, weight: 70, targetWeight: 70))
    }

    func test_validateGoal_loseWeight_currentLessThanTarget_isInvalid() {
        XCTAssertNotNil(sut.validateGoal(.loseWeight, weight: 60, targetWeight: 70))
    }

    func test_validateGoal_gainMuscle_currentLessThanTarget_isValid() {
        XCTAssertNil(sut.validateGoal(.gainMuscle, weight: 70, targetWeight: 80))
    }

    func test_validateGoal_gainMuscle_currentEqualsTarget_isInvalid() {
        XCTAssertNotNil(sut.validateGoal(.gainMuscle, weight: 70, targetWeight: 70))
    }

    func test_validateGoal_maintain_currentEqualsTarget_isValid() {
        XCTAssertNil(sut.validateGoal(.maintain, weight: 70, targetWeight: 70))
    }

    func test_validateGoal_maintain_currentDiffersFromTarget_isInvalid() {
        XCTAssertNotNil(sut.validateGoal(.maintain, weight: 70, targetWeight: 71))
    }

    // MARK: - validateTargetWorkoutsWeekly

    func test_validateTargetWorkoutsWeekly_0_isValid() {
        XCTAssertNil(sut.validateTargetWorkoutsWeekly(0))
    }

    func test_validateTargetWorkoutsWeekly_7_isValid() {
        XCTAssertNil(sut.validateTargetWorkoutsWeekly(7))
    }

    func test_validateTargetWorkoutsWeekly_negative_isInvalid() {
        XCTAssertNotNil(sut.validateTargetWorkoutsWeekly(-1))
    }

    func test_validateTargetWorkoutsWeekly_8_isInvalid() {
        XCTAssertNotNil(sut.validateTargetWorkoutsWeekly(8))
    }

    func test_validateTargetWorkoutsWeekly_typicalValue_isValid() {
        XCTAssertNil(sut.validateTargetWorkoutsWeekly(3))
    }

    // MARK: - validateCaloriesNorm — deficit cases

    func test_validateCaloriesNorm_criticalDeficit_below40Percent() {
        let result = sut.validateCaloriesNorm(500, dailyEnergyExpenditure: 1000)
        XCTAssertTrue(result.contains("Критический дефицит"))
    }

    func test_validateCaloriesNorm_strongDeficit_40Percent_boundary() {
        let result = sut.validateCaloriesNorm(600, dailyEnergyExpenditure: 1000)
        XCTAssertTrue(result.contains("Сильный дефицит"))
    }

    func test_validateCaloriesNorm_markedDeficit_30Percent_boundary() {
        let result = sut.validateCaloriesNorm(700, dailyEnergyExpenditure: 1000)
        XCTAssertTrue(result.contains("Выраженный дефицит"))
    }

    func test_validateCaloriesNorm_moderateDeficit_20Percent_boundary() {
        let result = sut.validateCaloriesNorm(800, dailyEnergyExpenditure: 1000)
        XCTAssertTrue(result.contains("Умеренный дефицит"))
    }

    func test_validateCaloriesNorm_smallDeficit_10Percent_boundary() {
        let result = sut.validateCaloriesNorm(900, dailyEnergyExpenditure: 1000)
        XCTAssertTrue(result.contains("Небольшой дефицит"))
    }

    func test_validateCaloriesNorm_smallDeficit_midRange() {
        let result = sut.validateCaloriesNorm(950, dailyEnergyExpenditure: 1000)
        XCTAssertTrue(result.contains("Небольшой дефицит"))
    }

    // MARK: - validateCaloriesNorm — balance

    func test_validateCaloriesNorm_perfectBalance_zeroPercent() {
        let result = sut.validateCaloriesNorm(1000, dailyEnergyExpenditure: 1000)
        XCTAssertTrue(result.contains("баланс"))
    }

    // MARK: - validateCaloriesNorm — surplus cases

    func test_validateCaloriesNorm_smallSurplus_midRange() {
        let result = sut.validateCaloriesNorm(1050, dailyEnergyExpenditure: 1000)
        XCTAssertTrue(result.contains("Небольшой профицит"))
    }

    func test_validateCaloriesNorm_smallSurplus_10Percent_boundary() {
        let result = sut.validateCaloriesNorm(1100, dailyEnergyExpenditure: 1000)
        XCTAssertTrue(result.contains("Небольшой профицит"))
    }

    func test_validateCaloriesNorm_moderateSurplus_10to20Percent() {
        let result = sut.validateCaloriesNorm(1150, dailyEnergyExpenditure: 1000)
        XCTAssertTrue(result.contains("Умеренный профицит"))
    }

    func test_validateCaloriesNorm_elevatedSurplus_20to30Percent() {
        let result = sut.validateCaloriesNorm(1200, dailyEnergyExpenditure: 1000)
        XCTAssertTrue(result.contains("Повышенный профицит"))
    }

    func test_validateCaloriesNorm_strongSurplus_30to40Percent() {
        let result = sut.validateCaloriesNorm(1300, dailyEnergyExpenditure: 1000)
        XCTAssertTrue(result.contains("Сильный профицит"))
    }

    func test_validateCaloriesNorm_criticalSurplus_above40Percent() {
        let result = sut.validateCaloriesNorm(1400, dailyEnergyExpenditure: 1000)
        XCTAssertTrue(result.contains("Критический профицит"))
    }

    func test_validateCaloriesNorm_criticalSurplus_farAbove() {
        let result = sut.validateCaloriesNorm(2000, dailyEnergyExpenditure: 1000)
        XCTAssertTrue(result.contains("Критический профицит"))
    }

    func test_validateCaloriesNorm_returnsNonEmptyString_always() {
        XCTAssertFalse(sut.validateCaloriesNorm(1000, dailyEnergyExpenditure: 2000).isEmpty)
        XCTAssertFalse(sut.validateCaloriesNorm(2000, dailyEnergyExpenditure: 1000).isEmpty)
    }

    // MARK: - getCaloriesNormHint

    func test_getCaloriesNormHint_loseWeight_negativeActivity_mentionsDietDeficit() {
        let result = sut.getCaloriesNormHint(1200, basalMetabolicRate: 1500, goal: .loseWeight)
        XCTAssertFalse(result.isEmpty)
        XCTAssertTrue(result.contains("Дефицит"))
    }

    func test_getCaloriesNormHint_loseWeight_zeroActivity_mentionsTraining() {
        let result = sut.getCaloriesNormHint(1500, basalMetabolicRate: 1500, goal: .loseWeight)
        XCTAssertFalse(result.isEmpty)
        XCTAssertTrue(result.contains("тренировок") || result.contains("базовый обмен"))
    }

    func test_getCaloriesNormHint_loseWeight_positiveActivity_mentionsBurning() {
        let result = sut.getCaloriesNormHint(2000, basalMetabolicRate: 1500, goal: .loseWeight)
        XCTAssertFalse(result.isEmpty)
        XCTAssertTrue(result.contains("сжигать") || result.contains("ккал"))
    }

    func test_getCaloriesNormHint_gainMuscle_deficitRelativeToRest_warnsAboutMuscle() {
        let result = sut.getCaloriesNormHint(1200, basalMetabolicRate: 1500, goal: .gainMuscle)
        XCTAssertFalse(result.isEmpty)
        XCTAssertTrue(result.contains("мышцы") || result.contains("Дефицит"))
    }

    func test_getCaloriesNormHint_gainMuscle_smallSurplus_suggestsMore() {
        let result = sut.getCaloriesNormHint(1600, basalMetabolicRate: 1500, goal: .gainMuscle)
        XCTAssertFalse(result.isEmpty)
        XCTAssertTrue(result.contains("профицит") || result.contains("Профицит"))
    }

    func test_getCaloriesNormHint_gainMuscle_goodSurplus_positiveMessage() {
        let result = sut.getCaloriesNormHint(1900, basalMetabolicRate: 1500, goal: .gainMuscle)
        XCTAssertFalse(result.isEmpty)
        XCTAssertTrue(result.contains("Профицит") || result.contains("роста"))
    }

    func test_getCaloriesNormHint_maintain_smallDeficit_suggestsIncrease() {
        let result = sut.getCaloriesNormHint(1200, basalMetabolicRate: 1500, goal: .maintain)
        XCTAssertFalse(result.isEmpty)
        XCTAssertTrue(result.contains("дефицит") || result.contains("худеть"))
    }

    func test_getCaloriesNormHint_maintain_nearBalance_positiveMessage() {
        let result = sut.getCaloriesNormHint(1500, basalMetabolicRate: 1500, goal: .maintain)
        XCTAssertFalse(result.isEmpty)
        XCTAssertTrue(result.contains("баланс") || result.contains("стабильным") || result.contains("активности"))
    }

    func test_getCaloriesNormHint_maintain_surplusActivity_mentionsBurning() {
        let result = sut.getCaloriesNormHint(1800, basalMetabolicRate: 1500, goal: .maintain)
        XCTAssertFalse(result.isEmpty)
        XCTAssertTrue(result.contains("сжигать") || result.contains("ккал"))
    }
}
