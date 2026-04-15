import Foundation
import HealthKit
import Combine
import SwiftUI

@MainActor
final class ProfileViewModel: ObservableObject {
    enum ProfileEvent {
        case idle
        case logoutSuccess
    }
    
    @Published var avatarUrl: String = ""
    @Published var height: Int = 0 {
        willSet {
            if newValue >= 100 && newValue <= 250 {
                heightError = nil
            } else {
                heightError = "Введите корректное значение"
            }
        }
    }
    @Published var heightError: String? = nil
    @Published var weight: Float = 0.0 {
        willSet {
            if newValue < 40 {
                weightError = "Введите корректное значение"
            } else {
                weightError = nil
            }
        }
    }
    @Published var weightError: String? = nil
    @Published var lifestyle: Lifestyle = .active
    @Published var goal: MainGoal = .maintain {
        willSet {
            switch newValue {
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
    }
    @Published var targetWeight: Float = 0.0 {
        willSet {
            if newValue < 40 {
                weightError = "Введите корректное значение"
            } else {
                weightError = nil
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

    func load() async {
        do {
            screenState = .loading
            profile = try await service.fetchProfile()

            screenState = .idle
        } catch {
            if AppRouter.shared.handleIfUnauthorized(error) { return }
            screenState = .error(error.localizedDescription)
        }
    }

    func save() async {
        do {
            try validate()
            screenState = .loading
            
            let profile = UserProfile(
                height: height,
                photo: avatarUrl,
                goal: goal,
                lifestyle: lifestyle,
                currentWeight: Double(weight),
                targetWeight: Double(targetWeight),
                targetCaloriesDaily: targetCaloriesDaily,
                targetWorkoutsWeekly: targetWorkoutsWeekly
            )
            
            try await service.updateProfile(profile)
            let basalMetabolicRate = await calculateBasalMetabolicRate()
            
            SharedWidgetStorage.shared.saveTargetCalories(Int(targetCaloriesDaily))
            SharedWidgetStorage.shared.saveBasalMetabolicRate(Int(basalMetabolicRate))
            
            isEditing = false
            screenState = .idle
        } catch {
            if AppRouter.shared.handleIfUnauthorized(error) { return }
            screenState = .error("Ошибка сохранения")
        }
    }

    func logout() {
        service.logout()
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
        do {
            dateOfBirth = try healthService.fetchDateOfBirth()
            gender = try healthService.fetchGender()
        } catch {
            screenState = .error(error.localizedDescription)
        }
    }
}

#Preview {
    ProfileView()
}
