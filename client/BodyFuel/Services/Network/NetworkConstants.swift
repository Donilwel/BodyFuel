import Foundation

enum API {
    static let baseURLString = "http://172.20.10.12:8080/api/v1"
    
    static let userParameters = "/crud/user/params"
    static let weight = "/crud/user/weight"
    static let uploadAvatar = "/avatars"
    
    enum Auth {
        static let register = "/auth/register"
        static let login = "/auth/login"
    }
}
