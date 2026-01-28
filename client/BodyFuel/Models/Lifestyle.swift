enum Lifestyle: String, CaseIterable, Identifiable {
    case sedentary
    case active
    case sporty

    var id: String { rawValue }

    var title: String {
        switch self {
        case .sedentary: return "Малоподвижный"
        case .active: return "Умеренная активность"
        case .sporty: return "Очень высокая активность"
        }
    }
    
    var description: String {
        switch self {
        case .sedentary: return "Сидячая работа, мало или нет тренировок"
        case .active: return "Умеренные тренировки 3-5 раз в неделю"
        case .sporty: return "Очень интенсивные тренировки или тяжелая физическая работа"
        }
    }
    
    var physicalActivityLevel: Float {
        switch self {
        case .sedentary: return 1.2
        case .active: return 1.55
        case .sporty: return 1.9
        }
    }
}
