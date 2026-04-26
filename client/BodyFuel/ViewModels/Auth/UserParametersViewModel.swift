import Foundation
import HealthKit
import Combine
import PhotosUI
import SwiftUI
import WidgetKit

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
    @Published var fitnessLevel: FitnessLevel?
    @Published var goal: MainGoal?
    @Published var goalError: String? = nil
    @Published var targetWeight: Float = 0.0
    @Published var targetWeightError: String? = nil
    @Published var targetCaloriesDaily: Float = 0.0
    @Published var basalMetabolicRate: Float = 0.0
    @Published var dailyEnergyExpenditure: Float = 0.0
    @Published var targetCaloriesError: String? = nil
    @Published var targetWorkoutsWeekly: Float = 0.0
    @Published var healthIntegrationError: String? = nil

    @Published var manualDateOfBirth: Date = Calendar.current.date(byAdding: .year, value: -25, to: Date()) ?? Date()
    @Published var manualGender: HKBiologicalSex = .male
    @Published var manualDateOfBirthError: String? = nil

    var weight: Float {
        Float(weightString) ?? 0.0
    }
    
    private let authService: AuthServiceProtocol = AuthService.shared
    private let userParametersService: UserParametersServiceProtocol = UserParametersService.shared
    private let healthService: HealthKitServiceProtocol = HealthKitService.shared
    private let userParametersValidator: UserParametersValidatorProtocol = UserParametersValidator.shared
    
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
        
        basalMetabolicRate = (10 * weight) + (6.25 * Float(height)) - (5 * age)
        
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
            
            guard let lifestyle, let fitnessLevel, let goal else {
                return
            }

            let userParametersPayload = UserParametersPayload(
                height: height,
                lifestyle: lifestyle,
                fitnessLevel: fitnessLevel,
                avatarData: avatarData ?? Data(),
                targetCaloriesDaily: Int(targetCaloriesDaily),
                targetWeight: targetWeight,
                targetWorkoutsWeeks: Int(targetWorkoutsWeekly),
                mainGoal: goal
            )
            
            async let sendUserWeight: () = userParametersService.sendCurrentWeight(weight)
            async let sendUserParameters: () = userParametersService.sendUserParameters(userParametersPayload)
            
            let (_, _) = try await (sendUserWeight, sendUserParameters)
            
            UserStore.shared.setTargetCalories(Int(targetCaloriesDaily))
            UserStore.shared.setBasalMetabolicRate(Int(basalMetabolicRate))

            UserSessionManager.shared.hasCompletedParametersSetup = true
            WidgetCenter.shared.reloadAllTimelines()
        } catch {
            print("[ERROR] [UserParametersViewModel/submit]: \(error.localizedDescription)")
            screenState = .error("Попробуйте еще раз позже")
        }
    }
    
    func validateLive() {
        heightError = userParametersValidator.validateHeight(height)
        weightError = userParametersValidator.validateWeight(weight)
        if goal == nil {
            goalError = "Выберите цель"
        } else {
            goal == .maintain ? targetWeight = weight : ()
            goalError = userParametersValidator.validateGoal(goal ?? .maintain, weight: weight, targetWeight: targetWeight)
        }
        targetWeightError = userParametersValidator.validateTargetWeight(targetWeight, weight: weight, goal: goal ?? .maintain)
    }
    
    func validateCaloriesNorm() -> String {
        userParametersValidator.validateCaloriesNorm(
            targetCaloriesDaily,
            dailyEnergyExpenditure: dailyEnergyExpenditure
        )
    }
    
    func getCaloriesNormHint() -> String {
        userParametersValidator.getCaloriesNormHint(
            targetCaloriesDaily,
            basalMetabolicRate: basalMetabolicRate,
            goal: goal ?? .maintain
        )
    }
    
    private func fetchHealthInfo() async {
        guard HealthKitService.shared.hasGrantedPermission else {
            dateOfBirth = manualDateOfBirth
            gender = manualGender
            healthIntegrationError = nil
            return
        }

        do {
            dateOfBirth = try healthService.fetchDateOfBirth()
        } catch {
            dateOfBirth = nil
        }

        do {
            gender = try healthService.fetchGender()
        } catch {
            gender = nil
        }

        if dateOfBirth == nil || gender == nil {
            healthIntegrationError = "Данные о поле и возрасте не найдены в приложении Здоровье — используются средние значения"
        } else {
            healthIntegrationError = nil
        }
    }

    private func validate() throws {
        let hasEmptyFields = height == 0 || weight == 0.0 || lifestyle == .none || fitnessLevel == .none || goal == .none || targetWeight == 0.0
        
        let hasErrors = [heightError, weightError, targetCaloriesError, goalError].contains { $0 != nil }
        
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
