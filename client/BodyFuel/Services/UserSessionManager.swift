import Foundation
import SwiftKeychainWrapper

final class UserSessionManager {
    static let shared = UserSessionManager()
    
    var hasCompletedOnboarding = false
    
    private let keychain = KeychainWrapper.standard
    private let defaults = UserDefaults.standard
    
    private enum Keys {
        static let currentUserId = "currentUserId"
        static let usersList = "usersList"
        
        static func setupStatus(for userId: String) -> String {
            return "setup_status_\(userId)"
        }
        
        static func authToken(for userId: String) -> String {
            return "auth_token_\(userId)"
        }
        
        static func userData(for userId: String) -> String {
            return "user_data_\(userId)"
        }
    }
    
    var currentUserId: String? {
        get { defaults.string(forKey: Keys.currentUserId) }
        set { defaults.set(newValue, forKey: Keys.currentUserId) }
    }
    
    var allUsers: [String] {
        get { defaults.array(forKey: Keys.usersList) as? [String] ?? [] }
        set { defaults.set(newValue, forKey: Keys.usersList) }
    }
    
    func hasCompletedParametersSetup(for userId: String) -> Bool {
        let key = Keys.setupStatus(for: userId)
        return keychain.bool(forKey: key) ?? false
    }
    
    func setHasCompletedParametersSetup(_ value: Bool, for userId: String) {
        let key = Keys.setupStatus(for: userId)
        keychain.set(value, forKey: key)
    }
    
    var hasCompletedParametersSetup: Bool {
        get {
            guard let userId = currentUserId else { return false }
            return hasCompletedParametersSetup(for: userId)
        }
        set {
            guard let userId = currentUserId else { return }
            setHasCompletedParametersSetup(newValue, for: userId)
        }
    }
    
    func authToken(for userId: String) -> String? {
        let key = Keys.authToken(for: userId)
        return keychain.string(forKey: key)
    }
    
    func setAuthToken(_ token: String, for userId: String) {
        let key = Keys.authToken(for: userId)
        keychain.set(token, forKey: key)
    }
    
    func login(userId: String, token: String) {
        setAuthToken(token, for: userId)
        
        currentUserId = userId
        
        var users = allUsers
        if !users.contains(userId) {
            users.append(userId)
            allUsers = users
        }
    }
    
    func logout(userId: String? = nil) {
        let userId = userId ?? currentUserId
        guard let userId = userId else { return }
        
        let tokenKey = Keys.authToken(for: userId)
        keychain.removeObject(forKey: tokenKey)
        
        if userId == currentUserId {
            currentUserId = nil
        }
    }
    
    func deleteUser(userId: String) {
        let keysToRemove = [
            Keys.setupStatus(for: userId),
            Keys.authToken(for: userId),
            Keys.userData(for: userId)
        ]
        
        keysToRemove.forEach { key in
            keychain.removeObject(forKey: key)
        }
        
        var users = allUsers
        users.removeAll { $0 == userId }
        allUsers = users
        
        if userId == currentUserId {
            currentUserId = nil
        }
    }
}

extension UserDefaults {
    private static let sessionManager = UserSessionManager.shared
    
    var hasCompletedParametersSetup: Bool {
        get { Self.sessionManager.hasCompletedParametersSetup }
        set { Self.sessionManager.hasCompletedParametersSetup = newValue }
    }
}
