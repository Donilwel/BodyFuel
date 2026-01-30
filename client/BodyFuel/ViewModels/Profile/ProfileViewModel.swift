import Foundation
import Combine

@MainActor
final class ProfileViewModel: ObservableObject {
    @Published var profile: UserProfile?
    @Published var screenState: ScreenState = .idle
    @Published var isEditing = false

    private let service: ProfileServiceProtocol = ProfileService.shared

    func load() async {
        do {
            screenState = .loading
            profile = try await service.fetchProfile()
            screenState = .idle
        } catch {
            screenState = .error("Не удалось загрузить профиль")
        }
    }

    func save() async {
        guard let profile else { return }

        do {
            screenState = .loading
            try await service.updateProfile(profile)
            isEditing = false
            screenState = .idle
        } catch {
            screenState = .error("Ошибка сохранения")
        }
    }
}
