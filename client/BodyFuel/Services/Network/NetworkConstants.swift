import Foundation

enum API {
    static let baseURLString = "http://localhost:8080/api/v1"
    
    enum Auth {
        static let register = "/auth/register"
        static let login = "/auth/login"
    }
}
