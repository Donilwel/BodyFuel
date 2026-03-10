import Foundation

enum DeepLink {
    case workouts
    case food
    case calories

    init?(url: URL) {
        guard url.scheme == "bodyfuel" else { return nil }

        switch url.host {
        case "workouts":
            self = .workouts
            
        case "food":
            self = .food
            
        case "calories":
            self = .calories

        default:
            return nil
        }
    }
}
