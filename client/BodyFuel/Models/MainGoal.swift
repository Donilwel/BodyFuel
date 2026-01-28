enum MainGoal: String, CaseIterable, Identifiable {
    case loseWeight
    case gainMuscle
    case maintain

    var id: String { rawValue }

    var title: String {
        switch self {
        case .loseWeight: return "Похудение"
        case .gainMuscle: return "Набор мышечной массы"
        case .maintain: return "Сохранение веса"
        }
    }
    
    var backendValue: String {
        switch self {
        case .loseWeight: return "lose_weight"
        case .gainMuscle: return "build_muscle"
        case .maintain: return "stay_fit"
        }
    }
}
