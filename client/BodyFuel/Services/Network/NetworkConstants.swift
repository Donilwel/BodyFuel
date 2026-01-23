import Foundation

enum API {
    static let baseURLString = "http://localhost:8080/api/v2"
    
    enum Auth {
        static let register = "/auth/register"
        static let login = "/auth/login"
    }
}
