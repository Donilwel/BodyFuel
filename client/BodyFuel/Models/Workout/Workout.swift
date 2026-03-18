import Foundation

struct Workout: Identifiable {
    var id: UUID = UUID()
    let title: String
    let type: ExerciseType
    let duration: Int
    let calories: Int
    let muscles: [String]
    let place: WorkoutPlace
    let exercises: [Exercise]
}

enum WorkoutPlace: String {
    case home = "Дом"
    case gym = "Спортзал"
    case outdoor = "На улице"
}
