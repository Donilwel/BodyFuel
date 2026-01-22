enum Lifestyle: String, CaseIterable, Identifiable {
    case active
    case sedentary

    var id: String { rawValue }

    var title: String {
        switch self {
        case .active: return "Активный"
        case .sedentary: return "Малоподвижный"
        }
    }
}
