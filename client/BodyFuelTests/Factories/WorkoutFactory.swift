import Foundation
@testable import BodyFuel

extension Exercise {
    static func stub(
        id: UUID = UUID(),
        name: String = "Приседания",
        type: ExerciseType = .lowerBody,
        description: String = "Базовое упражнение для ног",
        duration: Int = 30,
        repCount: Int? = 12,
        setCount: Int = 3,
        rest: Int = 60
    ) -> Exercise {
        Exercise(
            id: id,
            name: name,
            type: type,
            gifName: nil,
            description: description,
            duration: duration,
            repCount: repCount,
            setCount: setCount,
            rest: rest
        )
    }

    static func cardioStub(
        id: UUID = UUID(),
        name: String = "Бег на месте",
        duration: Int = 60,
        setCount: Int = 3,
        rest: Int = 30
    ) -> Exercise {
        Exercise(
            id: id,
            name: name,
            type: .cardio,
            gifName: nil,
            description: "Кардио упражнение",
            duration: duration,
            repCount: nil,
            setCount: setCount,
            rest: rest
        )
    }
}

extension Workout {
    static func stub(
        id: UUID = UUID(),
        title: String = "Тренировка",
        type: ExerciseType = .fullBody,
        duration: Int = 45,
        calories: Int = 300,
        place: WorkoutPlace = .home,
        exercises: [Exercise] = [.stub(), .stub(name: "Отжимания", type: .upperBody)]
    ) -> Workout {
        Workout(
            id: id,
            title: title,
            type: type,
            duration: duration,
            calories: calories,
            place: place,
            exercises: exercises
        )
    }
}
