import Foundation
import SwiftKeychainWrapper

final class TokenStorage {
    static let shared = TokenStorage()

    private init() {}

    var token: String? {
        get {
            KeychainWrapper.standard.string(forKey: "auth_token")
        }
        set {
            if let newValue {
                KeychainWrapper.standard.set(newValue, forKey: "auth_token")
            } else {
                KeychainWrapper.standard.removeObject(forKey: "auth_token")
            }
        }
    }

    func deleteToken() {
        self.token = nil
    }
}
