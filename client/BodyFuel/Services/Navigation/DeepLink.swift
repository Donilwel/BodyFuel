import Foundation

enum DeepLink {
    case workouts
    case workoutsWithID(String)
    case food
    case calories

    init?(url: URL) {
        guard url.scheme == "bodyfuel" else { return nil }

        switch url.host {
        case "workouts":
            let pathID = url.pathComponents.first(where: { $0 != "/" })
            if let id = pathID {
                self = .workoutsWithID(id)
            } else {
                self = .workouts
            }

        case "food":
            self = .food

        case "calories":
            self = .calories

        default:
            return nil
        }
    }
}
