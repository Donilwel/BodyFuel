enum MainGoal: String, CaseIterable, Identifiable {
    case loseWeight = "lose_weight"
    case gainMuscle = "build_muscle"
    case maintain = "stay_fit"

    var id: String { rawValue }

    var title: String {
        switch self {
        case .loseWeight: return "Похудение"
        case .gainMuscle: return "Набор мышечной массы"
        case .maintain: return "Сохранение веса"
        }
    }
}
