import Foundation

enum API {
    static let baseURLString = "http://192.168.1.15:8080/api/v1"
    
    static let userParameters = "/user/params"
    static let weight = "/user/weight"
    static let uploadAvatar = "/avatars"
    static let userInfo = "/user/info"
    
    enum Auth {
        static let register = "/auth/register"
        static let login = "/auth/login"
        static let refresh = "/auth/refresh"
        static let recover = "/auth/recover"
        static let resetPassword = "/auth/reset-password"
    }

    enum Workouts {
        static let base = "/workouts"
        static let history = "/workouts/history"
        static func workout(id: String) -> String { "/workouts/\(id)" }
    }

    static let weightHistory = "/user/weight/history"
    static let recommendations = "/recommendations"
    static let recommendationsRefresh = "/recommendations/refresh"
    static let caloriesHistory = "/user/calories/history"

    enum Nutrition {
        static let diary        = "/nutrition/diary"
        static let report       = "/nutrition/report"
        static let recipes      = "/nutrition/recipes"
        static let entries      = "/nutrition/entries"
        static let analyze      = "/nutrition/analyze"
        static let uploadPhoto  = "/nutrition/analyze/upload"
        static func entry(id: String) -> String { "/nutrition/entries/\(id)" }
    }
}
