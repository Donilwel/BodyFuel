import ActivityKit

protocol LiveActivityServiceProtocol {
    func start(
        workoutName: String,
        exerciseName: String,
        exerciseType: ExerciseType
    )
    func update(
        exerciseName: String,
        exerciseType: ExerciseType,
        exerciseDuration: Int,
        workoutPhase: WorkoutPhase,
        workoutProgress: Double
    )
    func end()
}

final class LiveActivityService: LiveActivityServiceProtocol {
    static let shared = LiveActivityService()
    
    private var activity: Activity<WorkoutAttributes>?
    
    private init() {}
    
    func start(
        workoutName: String,
        exerciseName: String,
        exerciseType: ExerciseType
    ) {
        guard ActivityAuthorizationInfo().areActivitiesEnabled else { return }
        
        let attributes = WorkoutAttributes(workoutName: workoutName)
        
        let state = WorkoutAttributes.ContentState(
            exerciseName: exerciseName,
            exerciseType: exerciseType,
            exerciseDuration: 0,
            workoutPhase: .waitingForStart,
            workoutProgress: 0
        )
        
        do {
            activity = try Activity.request(
                attributes: attributes,
                contentState: state,
                pushType: nil
            )
        } catch {
            print("[ERROR] [LiveActivityService/start]: Failed to start Live Activity: \(error)")
        }
    }
    
    func update(
        exerciseName: String,
        exerciseType: ExerciseType,
        exerciseDuration: Int,
        workoutPhase: WorkoutPhase,
        workoutProgress: Double
    ) {
        guard let activity else { return }
        
        let state = WorkoutAttributes.ContentState(
            exerciseName: exerciseName,
            exerciseType: exerciseType,
            exerciseDuration: exerciseDuration,
            workoutPhase: workoutPhase,
            workoutProgress: workoutProgress
        )
        
        Task {
            await activity.update(using: state)
        }
    }
    
    func end() {
        guard let activity else { return }
        
        Task {
            await activity.end(dismissalPolicy: .immediate)
        }
        
        self.activity = nil
    }
}
