import Foundation

protocol UserParametersValidatorProtocol {
    func validateHeight(_ height: Int) -> String?
    func validateWeight(_ weight: Float) -> String?
    func validateGoal(_ goal: MainGoal, weight: Float, targetWeight: Float) -> String?
    func validateTargetWeight(_ targetWeight: Float, weight: Float, goal: MainGoal) -> String?
    func validateTargetWorkoutsWeekly(_ targetWorkoutsWeekly: Int) -> String?
    func getCaloriesNormHint(_ targetCaloriesDaily: Float, basalMetabolicRate: Float, goal: MainGoal) -> String
    func validateCaloriesNorm(_ targetCaloriesDaily: Float, dailyEnergyExpenditure: Float) -> String
}

final class UserParametersValidator: UserParametersValidatorProtocol {
    static let shared = UserParametersValidator()
    
    private init() {}
    
    func validateHeight(_ height: Int) -> String? {
        return height >= 100 && height <= 250 ? nil : "Введите корректное значение"
    }
    
    func validateWeight(_ weight: Float) -> String? {
        return weight > 40 && weight < 300 ? nil : "Введите корректное значение"
    }
    
    func validateGoal(_ goal: MainGoal, weight: Float, targetWeight: Float) -> String? {
        var errorMessage: String? = nil
        switch goal {
        case .loseWeight:
            errorMessage = weight > targetWeight ? nil : "Для похудения целевой вес должен быть меньше текущего"
        case .gainMuscle:
            errorMessage = weight < targetWeight ? nil : "Для набора массы целевой вес должен быть больше текущего"
        case .maintain:
            errorMessage = weight == targetWeight ? nil : "Для поддержания веса целевой вес должен быть равен текущему"
        }
        return errorMessage
    }
    
    func validateTargetWeight(_ targetWeight: Float, weight: Float, goal: MainGoal) -> String? {
        var errorMessage: String? = nil
        errorMessage = weight > 40 ? nil : "Введите корректное значение"
        switch goal {
        case .loseWeight:
            errorMessage = weight > targetWeight ? nil : "Значение не соответствует цели - похудение"
        case .gainMuscle:
            errorMessage = weight < targetWeight ? nil : "Значение не соответствует цели - набор мышечной массы"
        case .maintain:
            errorMessage = weight == targetWeight ? nil : "Значение не соответствует цели - поддержание веса"
        }
        return errorMessage
    }
    
    func validateTargetWorkoutsWeekly(_ targetWorkoutsWeekly: Int) -> String? {
        return targetWorkoutsWeekly < 0 || targetWorkoutsWeekly > 7 ? "Введите значение от 0 до 7" : nil
    }
    
    func getCaloriesNormHint(_ targetCaloriesDaily: Float, basalMetabolicRate: Float, goal: MainGoal) -> String {
        let activityCalories = Int(targetCaloriesDaily) - Int(basalMetabolicRate)
        let monthlyKg = String(format: "%.1f", Float(abs(activityCalories)) * 30 / 7700)
        let runMinutes = max(10, abs(activityCalories) / 8)
        let walkMinutes = max(15, abs(activityCalories) / 4)

        switch goal {
        case .loseWeight:
            if activityCalories < 0 {
                return "Дефицит создаётся уже за счёт питания — тело тратит в покое больше, чем получает. Потеря веса: ~\(monthlyKg) кг/мес. Добавь тренировки, чтобы не терять мышцы"
            } else if activityCalories == 0 {
                return "Покрываешь ровно базовый обмен — похудеть получится только за счёт тренировок. Чем активнее занимаешься, тем быстрее результат"
            } else {
                return "Нужно сжигать на активности ~\(activityCalories) ккал/день (~\(runMinutes) мин бега). Без тренировок этот излишек будет откладываться"
            }

        case .gainMuscle:
            if activityCalories <= 0 {
                return "Дефицит относительно покоя — мышцы расти не будут. Подними калорийность выше \(Int(basalMetabolicRate)) ккал"
            } else if activityCalories < 200 {
                return "Профицит небольшой (~\(activityCalories) ккал/день). Для уверенного роста мышц рекомендуется 200–400 ккал сверх нормы — рассмотри увеличение"
            } else {
                return "Профицит ~\(activityCalories) ккал/день — хорошая база для роста. Ориентировочный набор: ~\(monthlyKg) кг/мес. Без силовых тренировок часть уйдёт в жир"
            }

        case .maintain:
            if activityCalories < -100 {
                return "Небольшой дефицит даже без тренировок — будешь постепенно худеть (~\(monthlyKg) кг/мес). Если цель — именно поддержание, подними калорийность"
            } else if activityCalories <= 100 {
                return "Практически в балансе с покоем — достаточно лёгкой ежедневной активности, чтобы вес оставался стабильным"
            } else {
                return "Для поддержания веса нужно сжигать ~\(activityCalories) ккал/день на активности — около \(runMinutes) мин бега или \(walkMinutes) мин ходьбы"
            }
        }
    }
    
    func validateCaloriesNorm(_ targetCaloriesDaily: Float, dailyEnergyExpenditure: Float) -> String {
        let diff = Int((targetCaloriesDaily - dailyEnergyExpenditure) / dailyEnergyExpenditure * 100)
        
        switch diff {
        case ..<(-40):
            return "Критический дефицит — может быть опасно для здоровья, советуем вернуться к более безопасному уровню"
        case -40..<(-30):
            return "Сильный дефицит — рекомендуем увеличить норму, чтобы избежать замедления обмена веществ и потери мышечной массы"
        case -30..<(-20):
            return "Выраженный дефицит — возможно снижение энергии и повышенная утомляемость, стоит следить за самочувствием"
        case -20..<(-10):
            return "Умеренный дефицит — подходит для стабильного снижения веса без сильного стресса для организма"
        case -10..<0:
            return "Небольшой дефицит — комфортный и безопасный темп снижения веса"
        case 0:
            return "Идеальный баланс для поддержания веса"
        case 0...10:
            return "Небольшой профицит — оптимально для постепенного и контролируемого набора массы"
        case 10..<20:
            return "Умеренный профицит — подходит для роста мышц при регулярных тренировках"
        case 20..<30:
            return "Повышенный профицит — часть избытка будет откладываться в виде жира, стоит скорректировать рацион"
        case 30..<40:
            return "Сильный профицит — рекомендуем снизить норму, чтобы уменьшить нагрузку на организм"
        default:
            return "Критический профицит — может негативно сказаться на здоровье, лучше выбрать более умеренное значение"
        }
    }
}
