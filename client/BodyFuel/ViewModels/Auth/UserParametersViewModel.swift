import Foundation
import HealthKit
import Combine
import PhotosUI
import SwiftUI

@MainActor
final class UserParametersViewModel: ObservableObject {
    enum CaloriesFormState {
        case preview
        case counting
        case editing
    }
    
    @Published var screenState: ScreenState = .idle
    @Published var caloriesFormState: CaloriesFormState = .preview
    
    @Published var avatarData: Data?
    @Published var avatarItem: PhotosPickerItem?
    @Published var heightString = ""
    @Published var heightError: String? = nil
    @Published var weightString = ""
    @Published var weightError: String? = nil
    @Published var lifestyle: Lifestyle?
    @Published var goal: MainGoal?
    @Published var targetWeight: Float = 0.0
    @Published var targetWeightError: String? = nil
    @Published var targetCaloriesDaily: Float = 0.0
    @Published var dailyEnergyExpenditure: Float = 0.0
    @Published var targetCaloriesError: String? = nil
    @Published var targetWorkoutsWeekly: Float = 0.0
    @Published var healthIntegrationError: String? = nil
    
    var weight: Float {
        Float(weightString) ?? 0.0
    }
    
    private let authService: AuthServiceProtocol = AuthService.shared
    private let healthService: HealthKitServiceProtocol = HealthKitService.shared
    
    private var dateOfBirth: Date?
    private var gender: HKBiologicalSex?
    private var height: Int {
        Int(heightString) ?? 0
    }
    
    func loadAvatar() async {
        avatarData = try? await avatarItem?.loadTransferable(type: Data.self)
    }
    
    func countRecommendedCalories() async {
        await fetchHealthInfo()
        
        guard let lifestyle, let goal, weight != 0, height != 0 else {
            targetCaloriesError = "Заполните предыдущие поля для оценки суточных трат калорий"
            return
        }
        
        let age = dateOfBirth != nil ? Float(Calendar.current.dateComponents([.year], from: dateOfBirth!, to: Date()).year!) : 30
        
        var basalMetabolicRate = (10 * weight) + (6.25 * Float(height)) - (5 * age)
        
        switch gender {
        case .female: basalMetabolicRate -= 161
        default: basalMetabolicRate += 5
        }
        
        dailyEnergyExpenditure = lifestyle.physicalActivityLevel * basalMetabolicRate
        
        targetCaloriesDaily = dailyEnergyExpenditure
        
        switch goal {
        case .loseWeight: targetCaloriesDaily *= 0.9
        case .gainMuscle: targetCaloriesDaily *= 1.1
        default: break
        }
        
        print("[INFO] [UserParametersViewModel/countRecommendedCalories]: is male: \(gender == .male), age: \(age)")
    }

    func submit() async {
        validateLive()
        
        do {
            try validate()
            screenState = .loading
            defer {
                screenState = .idle
                caloriesFormState = .preview
            }

            try await authService.sendUserParameters()
        }
        catch let error as AuthError {
            screenState = .error(error.errorDescription ?? "Заполните все поля")
        } catch {
            print("[ERROR] [UserParametersViewModel/submit]: \(error.localizedDescription)")
            screenState = .error("Попробуйте еще раз позже")
        }
    }
    
    func validateLive() {
        heightError = height >= 100 && height <= 250 ? nil : "Введите корректное значение"
        weightError = weight > 40 ? nil : "Введите корректное значение"
        
        switch goal {
        case .loseWeight:
            targetWeightError = weight > targetWeight ? nil : "Введите корректное значение"
        case .gainMuscle:
            targetWeightError = weight < targetWeight ? nil : "Введите корректное значение"
        case .maintain:
            targetWeightError = nil
            targetWeight = weight
        default:
            break
        }
    }
    
    func validateCaloriesNorm() -> String {
        let diff = Int((targetCaloriesDaily - dailyEnergyExpenditure) / dailyEnergyExpenditure * 100)
        
        switch diff {
        case ..<(-41):
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
    
    private func fetchHealthInfo() async {
        do {
            dateOfBirth = try healthService.fetchDateOfBirth()
            gender = try healthService.fetchGender()
        } catch {
            screenState = .error(error.localizedDescription)
        }
    }

    private func validate() throws {
        let hasEmptyFields = height == 0 || weight == 0.0 || lifestyle == .none || goal == .none || targetWeight == 0.0
        
        let hasErrors = [heightError, weightError, healthIntegrationError, targetCaloriesError].contains { $0 != nil }
        
        guard targetWeightError == nil else {
            throw AuthError.invalidData("Введите корректное значение желаемого веса и/или цели")
        }
        
        guard !hasEmptyFields, !hasErrors else {
            throw AuthError.invalidData("Заполните все поля")
        }
    }
}

#Preview {
    UserParametersView()
}
