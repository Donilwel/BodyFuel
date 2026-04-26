import Foundation

struct Workout: Identifiable, Codable {
    var id: UUID = UUID()
    let title: String
    let type: ExerciseType
    let duration: Int
    let calories: Int
    let place: WorkoutPlace
    let exercises: [Exercise]
}

enum WorkoutPlace: String, CaseIterable, Codable {
    case home = "Дом"
    case gym = "Спортзал"
    case outdoor = "На улице"

    var apiValue: String {
        switch self {
        case .home: return "home"
        case .gym: return "gym"
        case .outdoor: return "street"
        }
    }
}
