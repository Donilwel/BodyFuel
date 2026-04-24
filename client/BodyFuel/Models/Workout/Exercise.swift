import Foundation

struct Exercise: Identifiable, Codable {
    var id: UUID = UUID()
    let name: String
    let type: ExerciseType
    var gifName: String? = nil
    let duration: Int
    let repCount: Int?
    let setCount: Int
    let rest: Int
}
