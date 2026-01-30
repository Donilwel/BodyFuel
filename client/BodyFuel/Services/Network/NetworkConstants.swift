import Foundation

enum API {
    static let baseURLString = "http://127.0.0.1:8080/api/v1"
    
    static let userParameters = "/user/params"
    static let weight = "/user/weight"
    static let uploadAvatar = "/avatars"
    static let userInfo = "/user/info"
    
    enum Auth {
        static let register = "/auth/register"
        static let login = "/auth/login"
    }
}
