import Foundation

enum FitnessLevel: String, CaseIterable, Identifiable {
    case beginner     = "beginner"
    case intermediate = "intermediate"
    case professional = "professional"

    var id: String { rawValue }

    var title: String {
        switch self {
        case .beginner:     return "Начальный"
        case .intermediate: return "Средний"
        case .professional: return "Профессионал"
        }
    }
}
