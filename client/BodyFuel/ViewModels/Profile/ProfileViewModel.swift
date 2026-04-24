import Foundation
import HealthKit
import Combine
import SwiftUI
import PhotosUI

@MainActor
final class ProfileViewModel: ObservableObject {
    enum ProfileEvent {
        case idle
        case logoutSuccess
    }

    enum CaloriesFormState {
        case preview
        case counting
        case editing
    }

    @Published var avatarUrl: String = ""
    @Published var avatarData: Data?
    @Published var avatarItem: PhotosPickerItem?

    @Published var showCaloriesSheet = false
    @Published var caloriesFormState: CaloriesFormState = .preview
    @Published var sheetDailyExpenditure: Float = 0.0
    @Published var sheetBasalMetabolicRate: Float = 0.0
    @Published var sheetTargetCalories: Float = 0.0
    @Published var sheetHealthIntegrationError: String? = nil

    private let validator = UserParametersValidator.shared

    @Published var height: Int = 0 {
        willSet { heightError = validator.validateHeight(newValue) }
    }
    @Published var heightError: String? = nil
    @Published var weight: Float = 0.0 {
        willSet { weightError = validator.validateWeight(newValue) }
    }
    @Published var weightError: String? = nil
    @Published var lifestyle: Lifestyle = .active
    @Published var fitnessLevel: FitnessLevel = .beginner
    @Published var goal: MainGoal = .maintain {
        willSet {
            if newValue == .maintain { targetWeight = weight }
            targetWeightError = validator.validateTargetWeight(targetWeight, weight: weight, goal: newValue)
        }
    }
    @Published var targetWeight: Float = 0.0 {
        willSet {
            let weightErr = validator.validateWeight(newValue)
            if weightErr != nil {
                targetWeightError = weightErr
            } else {
                targetWeightError = validator.validateTargetWeight(newValue, weight: weight, goal: goal)
            }
        }
    }
    @Published var targetWeightError: String? = nil
    @Published var targetCaloriesDaily: Int = 0
    @Published var targetCaloriesError: String? = nil
    @Published var targetWorkoutsWeekly: Int = 0 {
        willSet {
            if newValue < 0 || newValue > 7 {
                targetWorkoutsError = "Введите значение от 0 до 7"
            } else {
                targetWorkoutsError = nil
            }
        }
    }
    @Published var targetWorkoutsError: String? = nil
    
    @Published var profile: UserProfile? {
        didSet {
            guard let profile else { return }
            avatarUrl = profile.photo
            height = profile.height
            weight = Float(profile.currentWeight)
            lifestyle = profile.lifestyle
            fitnessLevel = profile.fitnessLevel
            goal = profile.goal
            targetWeight = Float(profile.targetWeight)
            targetCaloriesDaily = profile.targetCaloriesDaily
            targetWorkoutsWeekly = profile.targetWorkoutsWeekly
        }
    }
    @Published var screenState: ScreenState = .idle
    @Published var event: ProfileEvent = .idle
    @Published var isEditing = false
    
    private var dateOfBirth: Date?
    private var gender: HKBiologicalSex?

    private let healthService: HealthKitServiceProtocol = HealthKitService.shared
    private let service: ProfileServiceProtocol = ProfileService.shared

    func loadAvatar() async {
        avatarData = try? await avatarItem?.loadTransferable(type: Data.self)
    }

    func load() async {
        do {
            screenState = .loading
            await UserStore.shared.load()
            if let storedProfile = UserStore.shared.profile {
                profile = storedProfile
            } else {
                profile = try await service.fetchProfile()
            }
            screenState = .idle
        } catch {
            if AppRouter.shared.handleIfUnauthorized(error) { return }
            let appError = ErrorMapper.map(error)
            screenState = .error(appError.errorDescription ?? "Не удалось загрузить профиль")
        }
    }

    func save() async {
        do {
            try validate()
            screenState = .loading

            if let data = avatarData {
                avatarUrl = try await PhotoService.shared.uploadUserAvatar(data: data)
                avatarData = nil
                avatarItem = nil
            }

            let profile = UserProfile(
                height: height,
                photo: avatarUrl,
                goal: goal,
                lifestyle: lifestyle,
                fitnessLevel: fitnessLevel,
                currentWeight: Double(weight),
                targetWeight: Double(targetWeight),
                targetCaloriesDaily: targetCaloriesDaily,
                targetWorkoutsWeekly: targetWorkoutsWeekly
            )
            
            try await service.updateProfile(profile)
            let basalMetabolicRate = await calculateBasalMetabolicRate()

            UserStore.shared.setTargetCalories(targetCaloriesDaily)
            UserStore.shared.setBasalMetabolicRate(Int(basalMetabolicRate))
            UserStore.shared.setProfile(profile)
            
            isEditing = false
            screenState = .idle
        } catch {
            if AppRouter.shared.handleIfUnauthorized(error) { return }
            screenState = .error("Ошибка сохранения")
        }
    }

    func countSheetCalories() async {
        await fetchHealthInfo()
        guard height != 0, weight != 0 else { return }

        let age = dateOfBirth != nil
            ? Float(Calendar.current.dateComponents([.year], from: dateOfBirth!, to: Date()).year!)
            : 30

        sheetBasalMetabolicRate = (10 * weight) + (6.25 * Float(height)) - (5 * age)
        switch gender {
        case .female: sheetBasalMetabolicRate -= 161
        default: sheetBasalMetabolicRate += 5
        }

        sheetDailyExpenditure = lifestyle.physicalActivityLevel * sheetBasalMetabolicRate
        sheetTargetCalories = sheetDailyExpenditure
        switch goal {
        case .loseWeight: sheetTargetCalories *= 0.9
        case .gainMuscle: sheetTargetCalories *= 1.1
        default: break
        }

        sheetHealthIntegrationError = (dateOfBirth == nil || gender == nil)
            ? "Данные о поле и возрасте не найдены в приложении Здоровье — используются средние значения"
            : nil

        caloriesFormState = .counting
    }

    func applySheetCalories() async {
        targetCaloriesDaily = Int(sheetTargetCalories)
        showCaloriesSheet = false
        caloriesFormState = .preview

        do {
            let updatedProfile = UserProfile(
                height: height,
                photo: avatarUrl,
                goal: goal,
                lifestyle: lifestyle,
                fitnessLevel: fitnessLevel,
                currentWeight: Double(weight),
                targetWeight: Double(targetWeight),
                targetCaloriesDaily: Int(sheetTargetCalories),
                targetWorkoutsWeekly: targetWorkoutsWeekly
            )
            try await service.updateProfile(updatedProfile)
            UserStore.shared.setTargetCalories(Int(sheetTargetCalories))
            UserStore.shared.setBasalMetabolicRate(Int(sheetBasalMetabolicRate))
            UserStore.shared.setProfile(updatedProfile)
        } catch {
            if AppRouter.shared.handleIfUnauthorized(error) { return }
            screenState = .error("Ошибка сохранения")
        }
    }

    func validateSheetCaloriesNorm() -> String {
        validator.validateCaloriesNorm(sheetTargetCalories, dailyEnergyExpenditure: sheetDailyExpenditure)
    }

    func getSheetCaloriesNormHint() -> String {
        validator.getCaloriesNormHint(sheetTargetCalories, basalMetabolicRate: sheetBasalMetabolicRate, goal: goal)
    }

    func logout() {
        AppRouter.shared.logout()
        event = .logoutSuccess
    }
    
    func deleteProfile() async {
        do {
            try await service.deleteProfile()
            event = .logoutSuccess
        } catch {
            screenState = .error("Не удалось удалить профиль, попробуйте позже")
        }
    }
    
    private func validate() throws {
        let hasEmptyFields = height == 0 || weight == 0.0 || targetWeight == 0.0
        
        let hasErrors = [heightError, weightError, targetCaloriesError, targetWorkoutsError].contains { $0 != nil }
        
        guard targetWeightError == nil else {
            throw AuthError.invalidData("Введите корректное значение желаемого веса и/или цели")
        }
        
        guard !hasEmptyFields, !hasErrors else {
            throw AuthError.invalidData("Заполните все поля")
        }
    }

    private func calculateBasalMetabolicRate() async -> Float {
        await fetchHealthInfo()
        
        let age = dateOfBirth != nil ? Float(Calendar.current.dateComponents([.year], from: dateOfBirth!, to: Date()).year!) : 30
        
        var basalMetabolicRate = (10 * weight) + (6.25 * Float(height)) - (5 * age)
        
        switch gender {
        case .female: basalMetabolicRate -= 161
        default: basalMetabolicRate += 5
        }
        
        return basalMetabolicRate
    }

    private func fetchHealthInfo() async {
        do { dateOfBirth = try healthService.fetchDateOfBirth() } catch { dateOfBirth = nil }
        do { gender = try healthService.fetchGender() } catch { gender = nil }
    }
}

#Preview {
    ProfileView()
}
