import WidgetKit
import SwiftUI
import ActivityKit

struct WorkoutLiveActivityView: View {
    
    let context: ActivityViewContext<WorkoutAttributes>
    
    var body: some View {
        VStack(alignment: .leading) {
            Text("Тренировка: \(context.attributes.workoutName.lowercased())")
                .font(.system(.headline, design: .rounded, weight: .bold))
                .foregroundColor(.white)
            
            HStack {
                if context.state.workoutPhase == .exercise {
                    VStack(alignment: .leading, spacing: 6) {
                        Text(context.state.exerciseName)
                            .font(.system(.body, design: .rounded, weight: .bold))
                            .foregroundColor(.white)
                        
                        if context.state.exerciseType == .cardio {
                            Text(formatTime(context.state.exerciseDuration))
                                .font(.system(.callout, design: .rounded, weight: .medium))
                                .foregroundColor(.white.opacity(0.9))
                        } else {
                            Text("\(context.state.exerciseDuration) раз")
                                .font(.system(.callout, design: .rounded, weight: .medium))
                                .foregroundColor(.white.opacity(0.9))
                        }
                    }
                } else {
                    VStack(alignment: .leading, spacing: 6) {
                        Text(workoutPhaseTitle(context.state.workoutPhase))
                            .font(.system(.body, design: .rounded, weight: .bold))
                            .foregroundColor(.white)
                        
                        Text("Следующее: \(context.state.exerciseName.lowercased())")
                            .font(.system(.callout, design: .rounded))
                            .foregroundColor(.white)
                    }
                }
                
                Spacer()
                
                VStack(alignment: .center) {
                    Text("\(Int(context.state.workoutProgress * 100))%")
                        .font(.system(.body, design: .rounded, weight: .bold))
                        .foregroundColor(.white)
                    
                    Text("тренировки выполнено")
                        .frame(width: 100)
                        .multilineTextAlignment(.center)
                        .font(.system(.body, design: .rounded, weight: .bold))
                        .foregroundColor(.white)
                }
                .padding()
                .background(Circle().fill(AppColors.primary.opacity(0.2)))
            }
        }
        .padding()
        .background(
            AppColors.widgetBackground
        )
        .foregroundColor(.white)
    }
    
    private func formatTime(_ seconds: Int) -> String {
        if seconds < 60 {
            return "\(seconds) сек"
        } else if seconds % 60 == 0 {
            return "\(seconds) мин"
        } else {
            return "\(seconds / 60) мин \(seconds % 60) сек"
        }
    }
    
    private func workoutPhaseTitle(_ phase: WorkoutPhase) -> String {
        switch phase {
        case .waitingForStart:
            return "Ожидание начала тренировки"
        case .restBetweenSets:
            return "Отдых между подходами"
        case .restBetweenExercises:
            return "Отдых между упражнениями"
        case .finished:
            return "Тренировка окончена!"
        default:
            return ""
        }
    }
}
