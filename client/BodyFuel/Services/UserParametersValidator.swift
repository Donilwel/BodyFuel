import Foundation

protocol UserParametersValidatorProtocol {
    func validateHeight(_ height: Int) -> String?
    func validateWeight(_ weight: Float) -> String?
    func validateGoal(_ goal: MainGoal, weight: Float, targetWeight: Float) -> String?
    func validateTargetWeight(_ targetWeight: Float, weight: Float, goal: MainGoal) -> String?
    func validateTargetWorkoutsWeekly(_ targetWorkoutsWeekly: Int) -> String?
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
