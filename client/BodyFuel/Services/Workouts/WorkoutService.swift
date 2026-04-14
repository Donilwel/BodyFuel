import Foundation

protocol WorkoutServiceProtocol {
    func generateWorkout(level: WorkoutLevel) async throws -> (workoutID: String, workout: Workout)
    func fetchWorkout(id: String) async throws -> (workoutID: String, workout: Workout)
    func fetchWorkoutHistory(limit: Int, offset: Int) async throws -> WorkoutHistoryResponseBody
    func updateWorkout(id: String, status: WorkoutStatus?, duration: Int64?) async throws
    func deleteWorkout(id: String) async throws
}

final class WorkoutService: WorkoutServiceProtocol {
    static let shared = WorkoutService()

    private let networkClient = NetworkClient.shared
    private let sharedWidgetStorage = SharedWidgetStorage.shared

    private init() {}

    func generateWorkout(level: WorkoutLevel) async throws -> (workoutID: String, workout: Workout) {
        guard let url = URL(string: API.baseURLString + API.Workouts.base) else {
            print("[ERROR] [WorkoutService/generateWorkout]: Invalid URL")
            throw NetworkError.invalidURL
        }

        let requestBody = GenerateWorkoutRequestBody(level: level, duration: nil)

        do {
            let response: WorkoutResponseBody = try await networkClient.request(
                url: url,
                method: .post,
                requestBody: requestBody
            )

            let workout = mapToWorkout(response)
            saveWidgetData(workout: workout, from: response)

            print("[INFO] [WorkoutService/generateWorkout]: Successfully generated workout \(response.id)")
            return (response.id, workout)
        } catch {
            print("[ERROR] [WorkoutService/generateWorkout]: \(error.localizedDescription)")
            throw error
        }
    }

    func fetchWorkout(id: String) async throws -> (workoutID: String, workout: Workout) {
        guard let url = URL(string: API.baseURLString + API.Workouts.workout(id: id)) else {
            print("[ERROR] [WorkoutService/fetchWorkout]: Invalid URL")
            throw NetworkError.invalidURL
        }

        do {
            let response: WorkoutResponseBody = try await networkClient.request(
                url: url,
                method: .get
            )

            let workout = mapToWorkout(response)
            saveWidgetData(workout: workout, from: response)

            print("[INFO] [WorkoutService/fetchWorkout]: Successfully fetched workout \(id)")
            return (response.id, workout)
        } catch {
            print("[ERROR] [WorkoutService/fetchWorkout]: \(error.localizedDescription)")
            throw error
        }
    }

    func fetchWorkoutHistory(limit: Int = 20, offset: Int = 0) async throws -> WorkoutHistoryResponseBody {
        var components = URLComponents(string: API.baseURLString + API.Workouts.history)
        components?.queryItems = [
            URLQueryItem(name: "limit", value: "\(limit)"),
            URLQueryItem(name: "offset", value: "\(offset)")
        ]

        guard let url = components?.url else {
            print("[ERROR] [WorkoutService/fetchWorkoutHistory]: Invalid URL")
            throw NetworkError.invalidURL
        }

        do {
            let response: WorkoutHistoryResponseBody = try await networkClient.request(
                url: url,
                method: .get
            )

            print("[INFO] [WorkoutService/fetchWorkoutHistory]: Successfully fetched \(response.total) workouts")
            return response
        } catch {
            print("[ERROR] [WorkoutService/fetchWorkoutHistory]: \(error.localizedDescription)")
            throw error
        }
    }

    func updateWorkout(id: String, status: WorkoutStatus?, duration: Int64?) async throws {
        guard let url = URL(string: API.baseURLString + API.Workouts.workout(id: id)) else {
            print("[ERROR] [WorkoutService/updateWorkout]: Invalid URL")
            throw NetworkError.invalidURL
        }

        let requestBody = UpdateWorkoutRequestBody(status: status, duration: duration)

        do {
            let _: DefaultDecodable = try await networkClient.request(
                url: url,
                method: .patch,
                requestBody: requestBody
            )

            print("[INFO] [WorkoutService/updateWorkout]: Successfully updated workout \(id)")
        } catch {
            print("[ERROR] [WorkoutService/updateWorkout]: \(error.localizedDescription)")
            throw error
        }
    }

    func deleteWorkout(id: String) async throws {
        guard let url = URL(string: API.baseURLString + API.Workouts.workout(id: id)) else {
            print("[ERROR] [WorkoutService/deleteWorkout]: Invalid URL")
            throw NetworkError.invalidURL
        }

        do {
            let _: DefaultDecodable = try await networkClient.request(
                url: url,
                method: .delete
            )

            print("[INFO] [WorkoutService/deleteWorkout]: Successfully deleted workout \(id)")
        } catch {
            print("[ERROR] [WorkoutService/deleteWorkout]: \(error.localizedDescription)")
            throw error
        }
    }
}

// MARK: - Mapping

private extension WorkoutService {
    func mapToWorkout(_ response: WorkoutResponseBody) -> Workout {
        let exercises = (response.exercises ?? []).map { mapToExercise($0) }

        return Workout(
            id: UUID(uuidString: response.id) ?? UUID(),
            title: mapWorkoutTitle(response.level),
            type: resolveWorkoutType(from: response.exercises ?? []),
            duration: Int((response.duration ?? 0) / 1_000_000_000 / 60),
            calories: response.totalCalories,
            muscles: resolveMuscles(from: exercises),
            place: mapPlace(response.exercises?.first?.placeExercise ?? ""),
            exercises: exercises
        )
    }

    func mapToExercise(_ body: WorkoutExerciseResponseBody) -> Exercise {
        let isCardio = body.typeExercise == "cardio"
        return Exercise(
            id: UUID(uuidString: body.exerciseID) ?? UUID(),
            name: body.name,
            type: mapExerciseType(body.typeExercise),
            gifName: body.linkGif.isEmpty ? nil : body.linkGif,
            duration: isCardio ? body.modifyReps : 0,
            repCount: isCardio ? nil : body.modifyReps,
            setCount: max(body.steps, 1),
            rest: body.modifyRelaxTime > 0 ? body.modifyRelaxTime : body.baseRelaxTime
        )
    }

    func mapExerciseType(_ typeString: String) -> ExerciseType {
        switch typeString {
        case "cardio": return .cardio
        case "upper_body": return .upperBody
        case "lower_body": return .lowerBody
        case "full_body": return .fullBody
        case "flexibility": return .flexibility
        default: return .fullBody
        }
    }

    func mapWorkoutTitle(_ level: String) -> String {
        switch level {
        case "workout_light": return "Лёгкая тренировка"
        case "workout_middle": return "Средняя тренировка"
        case "workout_hard": return "Интенсивная тренировка"
        default: return "Тренировка"
        }
    }

    func mapPlace(_ placeString: String) -> WorkoutPlace {
        switch placeString {
        case "gym": return .gym
        case "home": return .home
        case "street": return .outdoor
        default: return .outdoor
        }
    }

    func resolveWorkoutType(from exercises: [WorkoutExerciseResponseBody]) -> ExerciseType {
        let types = exercises.map { mapExerciseType($0.typeExercise) }
        if types.allSatisfy({ $0 == .cardio }) { return .cardio }
        if types.allSatisfy({ $0 == .flexibility }) { return .flexibility }
        return .fullBody
    }

    func resolveMuscles(from exercises: [Exercise]) -> [String] {
        var result: [String] = []
        var seen: Set<String> = []

        for exercise in exercises {
            let label: String
            switch exercise.type {
            case .upperBody: label = "Верхняя часть тела"
            case .lowerBody: label = "Нижняя часть тела"
            case .fullBody: label = "Full body"
            case .cardio: label = "Кардио"
            case .flexibility: label = "Гибкость"
            }
            if seen.insert(label).inserted {
                result.append(label)
            }
        }

        return result
    }

    func saveWidgetData(workout: Workout, from response: WorkoutResponseBody) {
        let typeLabel: String
        switch workout.type {
        case .cardio: typeLabel = "Cardio"
        case .flexibility: typeLabel = "Flexibility"
        default: typeLabel = "Full body"
        }

        let workoutModel = WorkoutModel(
            name: workout.title,
            duration: workout.duration,
            calories: workout.calories,
            location: workout.place.rawValue,
            type: typeLabel
        )

        sharedWidgetStorage.saveWorkout(workoutModel)
    }
}
