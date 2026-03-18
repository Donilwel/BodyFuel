import Foundation
import HealthKit
import Combine

@MainActor
final class WorkoutViewModel: ObservableObject {
    enum ScreenState {
        case loading
        case loaded
        case error(String)
    }
    
    @Published var shouldStartFromDeepLink = false
    
    @Published var exerciseStats: [ExerciseStats] = []
    @Published var currentSetRepCount: [String] = []
    @Published var currentExerciseRepCount: String = ""
    @Published var currentExerciseRepCountError: String = ""
    
    @Published var isWorkoutActive: Bool = false
    @Published var showWorkoutSummary: Bool = false
    
    @Published var screenState: ScreenState = .loading
    @Published var recommendedWorkout: Workout?
    @Published var currentWorkout: Workout?
    @Published var exercises: [Exercise] = []
    
    @Published var currentExerciseIndex = 0
    @Published var currentSet = 1

    @Published var phase: WorkoutPhase = .waitingForStart

    @Published var timeRemaining: Int = 0
    @Published var elapsedTime: Int = 0
    
    @Published var totalWorkoutElapsedTime: Int = 0
    @Published var totalWorkoutProgress: Double = 0
    
    @Published var totalCaloriesBurned: Double? = nil
    
    private let healthKitService: HealthKitServiceProtocol = HealthKitService.shared

    private var exerciseTimerCancellable: AnyCancellable?
    private var workoutTimerCancellable: AnyCancellable?
    
    private let workoutService: WorkoutServiceProtocol = WorkoutService.shared
    private let sharedWidgetStorage = SharedWidgetStorage.shared
    
    var currentExercise: Exercise? {
        guard currentExerciseIndex < exercises.count else { return nil }
        return exercises[currentExerciseIndex]
    }
    
    var progress: Double {
        let total: Int

        switch phase {
        case .exercise:
            total = currentExercise?.duration ?? 1
        case .restBetweenSets:
            total = currentExercise?.rest ?? 1
        case .restBetweenExercises:
            total = 90
        default:
            return 0
        }

        guard total > 0 else { return 0 }

        return Double(elapsedTime) / Double(total)
    }
    
    var workoutProgress: Double {
        guard exercises.count > 0 else { return 0 }
        
        var total: Int = 0
        var currentCount: Int = 0
        for i in 0..<exercises.count {
            total += exercises[i].setCount
            
            if i < currentExerciseIndex {
                currentCount += exercises[i].setCount
            } else if i == currentExerciseIndex && phase != .exercise && phase != .waitingForStart {
                currentCount += currentSet
            }
        }
        
        return Double(currentCount) / Double(total)
    }
    
    var phaseTitle: String {
        switch phase {
        case .waitingForStart:
            return "Нажмите кнопку для начала"
        case .exercise:
            return "Выполняйте упражнение"
        case .restBetweenSets:
            return "Отдых между подходами"
        case .restBetweenExercises:
            return "Отдых перед следующим упражнением"
        case .finished:
            return "Тренировка завершена"
        }
    }
    
    var isLastSet: Bool {
        currentExerciseIndex == exercises.count - 1 && currentSet == currentExercise?.setCount
    }
    
    func load() async {
        screenState = .loading
        do {
            let workout = try await workoutService.fetchTodayWorkout()
            
            await MainActor.run {
                recommendedWorkout = workout
                screenState = .loaded
            }
        } catch {
            let appError = ErrorMapper.map(error)
            screenState = .error(appError.errorDescription ?? "Попробуйте еще раз позже")
        }
    }
    
    func startWorkout() {
        guard let workout = recommendedWorkout else { return }

        currentWorkout = workout
        exercises = workout.exercises

        currentExerciseIndex = 0
        currentSet = 1

        phase = .waitingForStart
        isWorkoutActive = true
        
        startWorkoutTimer()
        
        var activityType: HKWorkoutActivityType = .traditionalStrengthTraining
        switch workout.type {
        case .cardio:
            activityType = .mixedCardio
        case .flexibility:
            activityType = .flexibility
        default:
            break
        }
        
        Task {
            await healthKitService.startWorkout(activityType: activityType)
        }
    }
    
    func changeWorkout() {
        
    }
    
    func startExercise() {
        guard let exercise = currentExercise else { return }

        stopExerciseTimer()
        
        currentExerciseRepCount = ""
        currentExerciseRepCountError = ""

        phase = .exercise

        timeRemaining = exercise.duration
        elapsedTime = 0

        startExerciseTimer()
    }
    
    func skipWorkout() {
        guard let exercise = currentExercise else {
            phase = .finished
            isWorkoutActive = false
            return
        }
        stopExerciseTimer()
        stopWorkoutTimer()
        
        if currentExerciseIndex == 0 && currentSet == 1 && (phase == .waitingForStart || phase == .exercise) { // если вообще ничего не начали
            isWorkoutActive = false
            Task {
                await healthKitService.discardWorkout()
            }
            // отправить на сервер статус failed
        } else {
            exerciseStats.append(ExerciseStats(
                exercise: exercise,
                repCount: currentSetRepCount
            ))
            currentExerciseRepCount = ""
            finishWorkout()
            // отправить ответ к серверу
        }
        
        phase = .finished
        exerciseStats.forEach { stats in
            print("\(stats.exercise.name): \(stats.repCount.joined(separator: ", ")); \(totalWorkoutElapsedTime)")
        }
    }
    
    func skipExercise() {
        stopExerciseTimer()
        currentExerciseRepCount = ""
        currentExerciseRepCountError = ""
        currentSet = 1
        startRestBetweenExercises()
    }
    
    func moveToNextPhase() {
        switch phase {
        case .waitingForStart:
            currentExerciseRepCount = ""
            currentExerciseRepCountError = ""
        case .exercise:
            finishExercise()
        case .restBetweenSets:
            currentSet += 1
            phase = .waitingForStart
        case .restBetweenExercises:
            nextExercise()
        case .finished:
            showWorkoutSummary = false
            isWorkoutActive = false
        }
    }
    
    func setupHealthKit() async {
        await healthKitService.requestAuthorization()
    }
    
    private func startExerciseTimer() {
        exerciseTimerCancellable?.cancel()
        exerciseTimerCancellable = Timer
            .publish(every: 1, on: .main, in: .common)
            .autoconnect()
            .sink { _ in
                self.tick()
            }
    }
    
    private func stopExerciseTimer() {
        exerciseTimerCancellable?.cancel()
    }
    
    private func startWorkoutTimer() {
        workoutTimerCancellable?.cancel()
        workoutTimerCancellable = Timer
            .publish(every: 1, on: .main, in: .common)
            .autoconnect()
            .sink { _ in
                self.totalWorkoutElapsedTime += 1
            }
    }
    
    private func stopWorkoutTimer() {
        workoutTimerCancellable?.cancel()
    }
    
    private func tick() {
        timeRemaining -= 1
        elapsedTime += 1
    }
    
    private func finishExercise() {
        guard let exercise = currentExercise else { return }
        
        if exercise.type != .cardio && (currentExerciseRepCount == "" || Int(currentExerciseRepCount) == nil) {
            currentExerciseRepCountError = "Введите корректное значение"
            return
        } else {
            currentExerciseRepCountError = ""
        }
        
        if exercise.type == .cardio {
            currentSetRepCount.append(String(exercise.duration - timeRemaining))
        } else {
            currentSetRepCount.append(currentExerciseRepCount)
        }
        
        if currentSet < exercise.setCount {
            startRestBetweenSets()
        } else {
            exerciseStats.append(ExerciseStats(
                exercise: exercise,
                repCount: currentSetRepCount
            ))
            
            currentExerciseRepCount = ""
            currentSetRepCount = []
            startRestBetweenExercises()
        }
    }
    
    private func startRestBetweenSets() {
        guard let exercise = currentExercise else { return }
        stopExerciseTimer()
        
        phase = .restBetweenSets

        timeRemaining = exercise.rest
        elapsedTime = 0

        startExerciseTimer()
    }
    
    private func startRestBetweenExercises() {
        stopExerciseTimer()

        phase = .restBetweenExercises
        timeRemaining = 90
        elapsedTime = 0

        startExerciseTimer()
    }
    
    private func nextExercise() {
        stopExerciseTimer()

        currentExerciseIndex += 1
        currentSet = 1

        if currentExerciseIndex >= exercises.count {
            finishWorkout()
            return
        }

        sharedWidgetStorage.saveWorkout(nil)
        
        phase = .waitingForStart
    }
    
    private func finishWorkout() {
        stopWorkoutTimer()
        phase = .finished
        isWorkoutActive = false
        showWorkoutSummary = true
        
        Task {
            let (calories, _) = await healthKitService.endWorkout()
            
            await MainActor.run {
                self.totalCaloriesBurned = calories
                self.phase = .finished
                self.isWorkoutActive = false
                self.showWorkoutSummary = true
                
                exerciseStats.forEach { stats in
                    print("\(stats.exercise.name): \(stats.repCount.joined(separator: ", ")), calories: \(calories)")
                }
            }
        }
    }
}
