import Foundation

struct ExerciseStats: Identifiable {
    var id: UUID = UUID()
    let exercise: Exercise
    let repCount: [String]
}
