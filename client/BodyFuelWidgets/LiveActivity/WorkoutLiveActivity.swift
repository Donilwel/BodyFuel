import WidgetKit
import SwiftUI

struct WorkoutLiveActivity: Widget {
    var body: some WidgetConfiguration {
        ActivityConfiguration(for: WorkoutAttributes.self) { context in
            WorkoutLiveActivityView(context: context)
        } dynamicIsland: { context in
            DynamicIsland {
                DynamicIslandExpandedRegion(.leading) {
                    HStack {
                        workoutPhaseImage(context.state.workoutPhase)
                    }
                    .padding(6)
                    .background(Circle().fill(AppColors.primary.opacity(0.2)))
                }
                
                DynamicIslandExpandedRegion(.trailing) {
                    Text("\(Int(context.state.workoutProgress * 100))%")
                        .font(.system(.title3, design: .rounded, weight: .bold))
                        .foregroundColor(AppColors.primary)
                        .padding(.horizontal, 8)
                        .padding(.vertical, 4)
                        .background(Capsule().fill(AppColors.primary.opacity(0.15)))
                }
                
                DynamicIslandExpandedRegion(.bottom) {
                    VStack(alignment: .leading, spacing: 6) {
                        if context.state.workoutPhase == .exercise {
                            HStack {
                                Text(context.state.exerciseName)
                                    .font(.system(.body, design: .rounded, weight: .semibold))
                                    .foregroundColor(.white)
                                
                                Spacer()
                                
                                if context.state.exerciseType == .cardio {
                                    Text(formatTime(context.state.exerciseDuration))
                                        .font(.system(.body, design: .rounded, weight: .medium))
                                        .foregroundColor(.white.opacity(0.9))
                                } else {
                                    Text("\(context.state.exerciseDuration) раз")
                                        .font(.system(.body, design: .rounded, weight: .medium))
                                        .foregroundColor(.white.opacity(0.9))
                                }
                            }
                        } else {
                            HStack {
                                Text(workoutPhaseTitle(context.state.workoutPhase))
                                    .font(.system(.body, design: .rounded, weight: .semibold))
                                    .foregroundColor(.white)
                                Spacer()
                                if context.state.workoutPhase == .restBetweenSets || context.state.workoutPhase == .restBetweenExercises {
                                    Text(formatTime(context.state.exerciseDuration))
                                        .font(.system(.body, design: .rounded, weight: .medium))
                                        .foregroundColor(.white.opacity(0.9))
                                }
                            }
                            
                            Text("Следующее: \(context.state.exerciseName)")
                                .font(.system(.caption, design: .rounded))
                                .foregroundColor(.white.opacity(0.7))
                        }
                    }
                    .padding(.horizontal, 16)
                    .padding(.vertical, 12)
                    .background(
                        RoundedRectangle(cornerRadius: 16)
                            .fill(.regularMaterial)
                            .opacity(0.9)
                            .shadow(color: .black.opacity(0.2), radius: 5, x: 0, y: 2)
                    )
                }
            } compactLeading: {
                workoutPhaseImage(context.state.workoutPhase)
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
                    .padding()
                    .background(Circle().fill(AppColors.primary.opacity(0.2)))
            } compactTrailing: {
                Text("\(Int(context.state.workoutProgress * 100))%")
                    .font(.system(.body, design: .rounded, weight: .bold))
                    .foregroundColor(AppColors.primary)
            } minimal: {
                workoutPhaseImage(context.state.workoutPhase)
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
                    .background(Circle().fill(AppColors.primary.opacity(0.2)))
            }
            .keylineTint(AppColors.primary)
        }
    }
    
    private func formatTime(_ seconds: Int) -> String {
        let minutes = seconds / 60
        let secs = seconds % 60
        return String(format: "%02d:%02d", minutes, secs)
    }
    
    private func workoutPhaseImage(_ phase: WorkoutPhase) -> some View {
        var imageSystemName = ""
        switch phase {
        case .waitingForStart:
            imageSystemName = "figure.dance"
        case .exercise:
            imageSystemName = "figure.step.training"
        case .restBetweenExercises, .restBetweenSets:
            imageSystemName = "gauge.with.needle.fill"
        case .finished:
            imageSystemName = "flag.pattern.checkered.2.crossed"
        }
        return Image(systemName: imageSystemName)
            .font(.system(.body, design: .rounded))
            .foregroundColor(AppColors.primary)
    }
    
    private func workoutPhaseTitle(_ phase: WorkoutPhase) -> String {
        switch phase {
        case .waitingForStart:
            return "Ожидание"
        case .restBetweenSets, .restBetweenExercises:
            return "Отдых"
        case .finished:
            return "Тренировка окончена!"
        default:
            return ""
        }
    }
}
