import Foundation

protocol WorkoutServiceProtocol {
    func fetchTodayWorkout() async throws -> Workout
}

final class WorkoutService: WorkoutServiceProtocol {
    static let shared = WorkoutService()
    
    private let sharedWidgetStorage = SharedWidgetStorage.shared
    
    private init() {}

    func fetchTodayWorkout() async throws -> Workout {
        let workoutModel = WorkoutModel(
            name: "Кардио + пресс",
            duration: 45,
            calories: 320,
            location: "Зал",
            type: "Кардио"
        )
        
        let workout = Workout(
            title: "Кардио + пресс",
            duration: 45,
            calories: 320,
            muscles: ["Пресс", "Выносливость"],
            place: .outdoor,
            exercises: [
                Exercise(
                    name: "Скручивания",
                    type: .core,
                    duration: 12,
                    repCount: 10,
                    setCount: 3,
                    rest: 20
                ),
                Exercise(
                    name: "Бег",
                    type: .cardio,
                    duration: 900,
                    repCount: nil,
                    setCount: 1,
                    rest: 20
                )
            ]
        )
        
        sharedWidgetStorage.saveWorkout(workoutModel)
        
        return workout
    }
}
