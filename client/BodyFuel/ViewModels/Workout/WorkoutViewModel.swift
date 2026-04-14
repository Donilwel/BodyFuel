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

    private(set) var currentWorkoutID: String?

    private var restTimeBetweenExercises = 90

    private var exerciseTimerCancellable: AnyCancellable?
    private var workoutTimerCancellable: AnyCancellable?
    
    private let workoutService: WorkoutServiceProtocol = WorkoutService.shared
    private let healthKitService: HealthKitServiceProtocol = HealthKitService.shared
    private let sharedWidgetStorage = SharedWidgetStorage.shared
    private lazy var liveActivityService: LiveActivityServiceProtocol = LiveActivityService.shared
    
    var currentExercise: Exercise? {
        guard currentExerciseIndex < exercises.count else { return nil }
        
        switch phase {
        case .restBetweenExercises:
            let nextIndex = currentExerciseIndex + 1
            guard nextIndex < exercises.count else { return exercises.last }
            return exercises[nextIndex]
        default:
            return exercises[currentExerciseIndex]
        }
    }
    
    var progress: Double {
        let total: Int

        switch phase {
        case .exercise:
            total = currentExercise?.duration ?? 60
        case .restBetweenSets:
            total = currentExercise?.rest ?? 60
        case .restBetweenExercises:
            total = restTimeBetweenExercises
        default:
            return 0
        }

        guard total > 0 else { return 0 }

        return Double(elapsedTime) / Double(total)
    }
    
    var workoutProgress: Double {
        guard !exercises.isEmpty else { return 0 }
        
        var total = 0
        var completed = 0
        
        for (index, exercise) in exercises.enumerated() {
            total += exercise.setCount
            
            if index < currentExerciseIndex {
                if let stats = exerciseStats.first(where: { $0.exercise.name == exercise.name }) {
                    completed += stats.repCount.count
                }
            } else if index == currentExerciseIndex {
                if let stats = exerciseStats.first(where: { $0.exercise.name == exercise.name }) {
                    completed += stats.repCount.count
                } else {
                    completed += currentSetRepCount.count
                }
            }
        }
        
        return Double(completed) / Double(total)
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
            let (workoutID, workout) = try await workoutService.generateWorkout(level: .beginner)

            await MainActor.run {
                currentWorkoutID = workoutID
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

        exerciseStats = []
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
        
        DispatchQueue.main.async { [weak self] in
            guard let self else { return }
            self.liveActivityService.start(
                workoutName: workout.title,
                exerciseName: self.currentExercise?.name ?? "",
                exerciseType: self.currentExercise?.type ?? .fullBody
            )
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
        
        updateLiveActivity()
    }
    
    func skipWorkout() {
        guard let exercise = currentExercise else {
            phase = .finished
            isWorkoutActive = false
            return
        }
        stopExerciseTimer()
        stopWorkoutTimer()
        
        DispatchQueue.main.async { [weak self] in
            guard let self else { return }
            liveActivityService.end()
        }
        
        if currentExerciseIndex == 0 && currentSet == 1 && (phase == .waitingForStart || phase == .exercise) { // если вообще ничего не начали
            isWorkoutActive = false
            Task {
                await healthKitService.discardWorkout()
                if let id = currentWorkoutID {
                    try? await workoutService.updateWorkout(id: id, status: .cancelled, duration: nil)
                }
            }
        } else {
            exerciseStats.append(ExerciseStats(
                exercise: exercise,
                repCount: currentSetRepCount
            ))
            currentExerciseRepCount = ""
            if let id = currentWorkoutID {
                let durationNano = Int64(totalWorkoutElapsedTime) * 1_000_000_000
                Task {
                    try? await workoutService.updateWorkout(id: id, status: .cancelled, duration: durationNano)
                }
            }
            finishWorkout()
        }
        
        phase = .finished
        exerciseStats.forEach { stats in
            print("\(stats.exercise.name): \(stats.repCount.joined(separator: ", ")); \(totalWorkoutElapsedTime)")
        }
    }
    
    func skipExercise() {
        stopExerciseTimer()
        
        if let exercise = currentExercise {
            let skippedReps = Array(repeating: "0", count: exercise.setCount)
            exerciseStats.append(ExerciseStats(
                exercise: exercise,
                repCount: skippedReps
            ))
        }
        
        currentExerciseRepCount = ""
        currentExerciseRepCountError = ""
        currentSet = 1
        currentSetRepCount = []
        startRestBetweenExercises()
        updateLiveActivity()
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
        
        updateLiveActivity()
    }
    
    private func startRestBetweenExercises() {
        stopExerciseTimer()

        phase = .restBetweenExercises
        timeRemaining = 90
        elapsedTime = 0

        startExerciseTimer()
        
        updateLiveActivity()
    }
    
    private func nextExercise() {
        stopExerciseTimer()

        currentSet = 1
        currentExerciseIndex += 1

        if currentExerciseIndex >= exercises.count {
            finishWorkout()
            return
        }

        sharedWidgetStorage.saveWorkout(nil)
        
        phase = .waitingForStart
        
        updateLiveActivity()
    }
    
    private func finishWorkout() {
        stopWorkoutTimer()
        
        DispatchQueue.main.async { [weak self] in
            guard let self else { return }
            self.liveActivityService.end()
        }
        
        phase = .finished
        isWorkoutActive = false
        showWorkoutSummary = true
        
        Task {
            let (calories, _) = await healthKitService.endWorkout()

            if let id = currentWorkoutID {
                let durationNano = Int64(totalWorkoutElapsedTime) * 1_000_000_000
                try? await workoutService.updateWorkout(id: id, status: .completed, duration: durationNano)
            }

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
    
    private func updateLiveActivity() {
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) { [weak self] in
            guard let self, self.isWorkoutActive, let exercise = self.currentExercise else { return }
            
            let duration: Int
            switch self.phase {
            case .exercise:
                duration = exercise.duration
            case .restBetweenSets:
                duration = exercise.rest
            case .restBetweenExercises:
                duration = self.restTimeBetweenExercises
            default:
                duration = 0
            }
            
            self.liveActivityService.update(
                exerciseName: exercise.name,
                exerciseType: exercise.type,
                exerciseDuration: duration,
                workoutPhase: self.phase,
                workoutProgress: self.workoutProgress
            )
        }
    }
}
