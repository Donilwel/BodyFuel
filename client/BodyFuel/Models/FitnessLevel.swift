import Foundation

enum FitnessLevel: String, CaseIterable, Identifiable, Codable {
    case beginner     = "not_active"
    case intermediate = "active"
    case professional = "sportive"

    var id: String { rawValue }

    var title: String {
        switch self {
        case .beginner:     return "Начальный"
        case .intermediate: return "Средний"
        case .professional: return "Профессионал"
        }
    }
    
    var changingOptions: [FitnessLevel] {
        switch self {
        case .beginner: return [.beginner]
        case .intermediate: return [.beginner, .intermediate]
        case .professional: return [.beginner, .intermediate, .professional]
        }
    }
}
