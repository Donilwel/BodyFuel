import Foundation
import Combine
import SwiftUI

@MainActor
final class UserParametersViewModel: ObservableObject {
    @Published var screenState: ScreenState = .idle

    @Published var heightString = ""
    @Published var heightError: String? = nil
    @Published var weightString = ""
    @Published var weightError: String? = nil
    @Published var lifestyle: Lifestyle?
    @Published var goal: Goal?
    @Published var targetWeight: Float = 0.0
    @Published var targetWeightError: String? = nil
    @Published var targetStepsDaily: Float = 0.0
    @Published var targetWorkoutsWeekly: Float = 0.0
    
    var weight: Float {
        Float(weightString) ?? 0.0
    }
    
    private let authService: AuthServiceProtocol = AuthService.shared
    
    private var height: Int {
        Int(heightString) ?? 0
    }

    func submit() async {
        validateLive()
        
        do {
            try validate()
            screenState = .loading
            defer { screenState = .idle }

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

    private func validate() throws {
        let hasEmptyFields = height == 0 || weight == 0.0
        let hasErrors = [heightError, weightError].contains { $0 != nil }
        
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
