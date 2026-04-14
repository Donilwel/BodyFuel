import SwiftUI

struct WorkoutSummaryView: View {
    @EnvironmentObject var viewModel: WorkoutViewModel
    @Environment(\.dismiss) var dismiss

    var body: some View {
        ZStack {
            AnimatedBackground()
            
            VStack(spacing: 32) {
                Text("Тренировка завершена")
                    .font(.title.bold())
                    .foregroundColor(.white)
                
                VStack(spacing: 16) {
                    statRow(
                        title: "Время тренировки",
                        value: String(
                            format: "%02d:%02d",
                            viewModel.totalWorkoutElapsedTime / 60,
                            viewModel.totalWorkoutElapsedTime % 60
                        )
                    )
                    
                    VStack(spacing: 12) {
//                        ForEach(viewModel.exerciseStats.indices, id: \.self) { index in
//                            exerciseStatsRow(for: viewModel.exerciseStats[index])
//                        }
                        exerciseStatsRow(for: ExerciseStats(exercise: viewModel.exerciseStats[0].exercise, repCount: ["8", "6", "5"]))
                        exerciseStatsRow(for: ExerciseStats(exercise: viewModel.exerciseStats[1].exercise, repCount: ["10", "10", "0"]))
                        exerciseStatsRow(for: ExerciseStats(exercise: viewModel.exerciseStats[2].exercise, repCount: ["5", "0", "0"]))
                        exerciseStatsRow(for: ExerciseStats(exercise: viewModel.exerciseStats[3].exercise, repCount: ["5", "0", "0"]))
                    }
                    .cardStyle()
                    
                    if let calories = viewModel.totalCaloriesBurned {
                        statRow(
                            title: "Сожженные калории",
                            value: String(format: "%.0f", calories)
                        )
                    }
                    
                    statRow(
                        title: "Выполнено",
                        value: "\(viewModel.workoutProgress * 100)%"
                    )
                }
                
                PrimaryButton(title: "На главный экран") {
                    viewModel.moveToNextPhase()
                }
                
            }
            .navigationBarBackButtonHidden()
            .padding()
        }
    }
    
    private func statRow(title: String, value: String) -> some View {
        HStack {
            Text(title)
                .foregroundColor(.white)
            Spacer()
            Text(value)
                .bold()
                .foregroundColor(.white)
        }
        .cardStyle()
    }
    
    private func exerciseStatsRow(for stat: ExerciseStats) -> some View {
        HStack(alignment: .top, spacing: 16) {
            Text(stat.exercise.name)
                .bold()
                .foregroundColor(.white)
                .lineLimit(2)
            
            Spacer()

            VStack(alignment: .leading, spacing: 8) {
                ForEach(0..<stat.repCount.count, id: \.self) { setIndex in
                    HStack {
                        Text("\(setIndex + 1) подход:")
                            .foregroundColor(.white.opacity(0.7))
                        if setIndex < stat.repCount.count {
                            Text(repCountText(repCount: stat.repCount[setIndex], exercise: stat.exercise))
                                .foregroundColor(.white)
                        } else {
                            Text("—")
                                .foregroundColor(.white)
                        }
                    }
                    .font(.subheadline)
                }
            }
        }
    }
    
    private func repCountText(repCount: String, exercise: Exercise) -> String {
        guard let count = Int(repCount) else { return "" }
        
        if exercise.type == .cardio {
            return String(
                format: "%02d:%02d",
                count / 60,
                count % 60
            )
        } else {
            let remainder10 = count % 10
            let remainder100 = count % 100
            
            if remainder10 == 1 && remainder100 != 11 {
                return "\(count) повторение"
            } else if remainder10 >= 2 && remainder10 <= 4 && (remainder100 < 10 || remainder100 >= 20) {
                return "\(count) повторения"
            } else {
                return "\(count) повторений"
            }
        }
    }
    
    private func formatTime(_ seconds: Int) -> String {
        let minutes = seconds / 60
        let remainingSeconds = seconds % 60

        return String(format: "%02d:%02d", minutes, remainingSeconds)
    }

}
