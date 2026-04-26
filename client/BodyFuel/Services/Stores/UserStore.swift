import Foundation
import Combine
import WidgetKit

@MainActor
final class UserStore: ObservableObject {
    static let shared = UserStore()

    @Published var profile: UserProfile?
    @Published var targetCalories: Int = 0
    @Published var basalMetabolicRate: Int = 0
    @Published var caloriesBurned: Int = 0
    @Published var todaySteps: Int = 0
    @Published var isDataStale = false

    private let profileService: ProfileServiceProtocol = ProfileService.shared
    private let sharedWidgetStorage = SharedWidgetStorage.shared
    private let diskCache = DiskCache.shared
    private let sessionManager = UserSessionManager.shared

    private var isLoadingProfile = false
    private var activityCancellable: AnyCancellable?
    private var stepsCancellable: AnyCancellable?

    private static let profileTTL: TimeInterval = 7 * 24 * 60 * 60

    private var profileKey: String {
        "user_profile_\(sessionManager.currentUserId ?? "anon")"
    }

    private init() {
        activityCancellable = HealthKitService.shared.$activeCalories
            .dropFirst()
            .receive(on: RunLoop.main)
            .sink { [weak self] calories in
                guard let self else { return }
                let value = Int(calories)
                guard value != self.caloriesBurned else { return }
                self.caloriesBurned = value
                self.sharedWidgetStorage.saveTodayBurnedCalories(value)
                WidgetCenter.shared.reloadAllTimelines()
            }

        stepsCancellable = HealthKitService.shared.$todaySteps
            .dropFirst()
            .receive(on: RunLoop.main)
            .sink { [weak self] steps in
                guard let self else { return }
                self.todaySteps = steps
                self.sharedWidgetStorage.saveTodaySteps(steps)
            }
    }

    func load() async {
        targetCalories = sharedWidgetStorage.getTargetCalories() ?? 0
        basalMetabolicRate = sharedWidgetStorage.getBasalMetabolicRate() ?? 0
        caloriesBurned = sharedWidgetStorage.getTodayBurnedCalories() ?? 0
        todaySteps = sharedWidgetStorage.getTodaySteps() ?? 0

        if profile == nil {
            loadProfileFromDisk()
        }

        applyProfileFallbacks()

        if NetworkMonitor.shared.isOnline {
            Task { await fetchProfileFromServer() }
        }
    }

    private func loadProfileFromDisk() {
        let cached = diskCache.load(UserProfile.self, key: profileKey)
        let expired = diskCache.isExpired(key: profileKey, ttl: Self.profileTTL)
        if let cached {
            profile = cached
            isDataStale = expired
            print("[INFO] [UserStore]: Loaded profile from disk (stale=\(expired))")
        }
    }

    private func fetchProfileFromServer() async {
        guard !isLoadingProfile else { return }
        isLoadingProfile = true
        defer { isLoadingProfile = false }
        do {
            let fetched = try await profileService.fetchProfile()
            NetworkMonitor.shared.markServerReachable()
            profile = fetched
            isDataStale = false
            diskCache.save(fetched, key: profileKey)
            print("[INFO] [UserStore]: Loaded profile from server")
        } catch {
            if isTransportError(error) {
                NetworkMonitor.shared.markServerUnreachable()
            }
        }
    }

    private func applyProfileFallbacks() {
        guard let profile else { return }
        if targetCalories == 0 {
            targetCalories = profile.targetCaloriesDaily
            sharedWidgetStorage.saveTargetCalories(profile.targetCaloriesDaily)
        }
        if basalMetabolicRate == 0 {
            let weightPart = 10.0 * profile.currentWeight
            let heightPart = 6.25 * Double(profile.height)
            let agePart = 5.0 * 30.0
            let bmr = Int(weightPart + heightPart - agePart + 5.0)
            basalMetabolicRate = max(bmr, 1200)
            sharedWidgetStorage.saveBasalMetabolicRate(basalMetabolicRate)
        }
    }

    // MARK: - Setters

    func setTargetCalories(_ calories: Int) {
        targetCalories = calories
        sharedWidgetStorage.saveTargetCalories(calories)
    }

    func setBasalMetabolicRate(_ bmr: Int) {
        basalMetabolicRate = bmr
        sharedWidgetStorage.saveBasalMetabolicRate(bmr)
    }

    func setCaloriesBurned(_ calories: Double) {
        caloriesBurned = Int(calories)
        sharedWidgetStorage.saveTodayBurnedCalories(Int(calories))
    }

    func setProfile(_ updated: UserProfile) {
        profile = updated
        isDataStale = false
        diskCache.save(updated, key: profileKey)
    }

    func invalidateProfile() {
        profile = nil
        diskCache.remove(key: profileKey)
    }

    func reset() {
        profile = nil
        targetCalories = 0
        basalMetabolicRate = 0
        caloriesBurned = 0
        todaySteps = 0
        isDataStale = false
        diskCache.remove(key: profileKey)
    }
}
