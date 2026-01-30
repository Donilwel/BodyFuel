import Foundation

extension UserDefaults {
    enum Keys {
        static let hasCompletedProfileSetup = "hasCompletedProfileSetup"
    }
    
    var hasCompletedProfileSetup: Bool {
        get { bool(forKey: Keys.hasCompletedProfileSetup) }
        set { set(newValue, forKey: Keys.hasCompletedProfileSetup) }
    }
}
