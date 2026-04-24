enum ExerciseType: String, Codable, CaseIterable {
    case cardio = "Кардио"
    case upperBody = "Верхняя часть тела"
    case lowerBody = "Нижняя часть тела"
    case fullBody = "Full body"
    case flexibility = "Гибкость"

    var apiValue: String {
        switch self {
        case .cardio: return "cardio"
        case .upperBody: return "upper_body"
        case .lowerBody: return "lower_body"
        case .fullBody: return "full_body"
        case .flexibility: return "flexibility"
        }
    }
}
