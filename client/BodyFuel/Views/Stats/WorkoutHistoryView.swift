import SwiftUI

struct WorkoutHistoryView: View {
    @ObservedObject private var store = WorkoutHistoryStore.shared
    @Environment(\.dismiss) private var dismiss

    private var completedWorkouts: [WorkoutHistoryItem] {
        store.workouts
            .filter { $0.status == "workout_done" || $0.status == "workout_failed" }
            .sorted { $0.date > $1.date }
    }

    var body: some View {
        ZStack {
            Color.clear
                .glassEffect(.regular.tint(AppColors.primary.opacity(0.6)).interactive(), in: .rect)
                .ignoresSafeArea()

            if store.isLoading && completedWorkouts.isEmpty {
                ProgressView()
                    .tint(.white)
            } else if completedWorkouts.isEmpty {
                VStack(spacing: 12) {
                    Image(systemName: "figure.run.circle")
                        .font(.system(size: 48))
                        .foregroundStyle(.white.opacity(0.4))
                    Text("Нет завершённых тренировок")
                        .font(.headline)
                        .foregroundStyle(.white.opacity(0.6))
                }
            } else {
                ScrollView {
                    VStack(spacing: 14) {
                        if store.isStale {
                            HStack(spacing: 6) {
                                Image(systemName: "clock.arrow.circlepath").font(.caption)
                                Text("Данные могут быть устаревшими")
                                    .font(.caption)
                            }
                            .foregroundStyle(.white.opacity(0.65))
                            .frame(maxWidth: .infinity, alignment: .leading)
                            .padding(.horizontal)
                            .padding(.top, 6)
                        }

                        ForEach(completedWorkouts) { workout in
                            WorkoutHistoryCard(workout: workout)
                        }
                    }
                    .padding()
                }
            }
        }
        .navigationTitle("История тренировок")
        .navigationBarTitleDisplayMode(.inline)
        .toolbarBackground(.hidden, for: .navigationBar)
        .toolbar {
            ToolbarItem(placement: .navigationBarLeading) {
                Button {
                    dismiss()
                } label: {
                    Image(systemName: "chevron.left")
                        .foregroundStyle(.white)
                }
            }
        }
        .task {
            await store.load()
        }
    }
}

// MARK: - Card

private struct WorkoutHistoryCard: View {
    let workout: WorkoutHistoryItem

    var body: some View {
        InfoCard {
            VStack(alignment: .leading, spacing: 10) {
                headerRow
                statsRow
                Divider().background(.white.opacity(0.15))
                exercisesList
            }
        }
    }

    private var headerRow: some View {
        HStack(alignment: .top) {
            VStack(alignment: .leading, spacing: 2) {
                if workout.status == "workout_failed" {
                    Text("Отменена")
                        .font(.subheadline)
                        .foregroundStyle(.white.opacity(0.7))
                }
                Text(workout.title)
                    .font(.headline)
                    .foregroundStyle(.white)
                Text(formattedDate)
                    .font(.caption)
                    .foregroundStyle(.white.opacity(0.55))
            }
            Spacer()
            levelBadge
        }
    }

    private var statsRow: some View {
        HStack(spacing: 16) {
            statItem(icon: "clock", value: formattedDuration)
            statItem(icon: "flame", value: "\(workout.totalCalories) ккал")
            statItem(icon: "checkmark.circle", value: "\(workout.completedCount)/\(workout.exercisesCount)")
        }
    }

    private func statItem(icon: String, value: String) -> some View {
        HStack(spacing: 4) {
            Image(systemName: icon)
                .font(.caption)
                .foregroundStyle(.white.opacity(0.6))
            Text(value)
                .font(.caption)
                .foregroundStyle(.white.opacity(0.85))
        }
    }

    private var exercisesList: some View {
        VStack(alignment: .leading, spacing: 6) {
            ForEach(workout.exercises) { exercise in
                WorkoutHistoryExerciseRow(exercise: exercise)
            }
        }
    }

    private var levelBadge: some View {
        Text(localizedLevel)
            .font(.caption2.bold())
            .foregroundStyle(.white)
            .padding(.horizontal, 8)
            .padding(.vertical, 3)
            .background(levelColor.opacity(0.3))
            .clipShape(Capsule())
    }

    private var localizedLevel: String {
        switch workout.level {
        case "workout_light": return "Лёгкая"
        case "workout_middle": return "Средняя"
        case "workout_hard": return "Интенсивная"
        default: return "—"
        }
    }

    private var levelColor: Color {
        switch workout.level {
        case "workout_light": return .green
        case "workout_middle": return .orange
        case "workout_hard": return .red
        default: return .white
        }
    }

    private var formattedDate: String {
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "ru_RU")
        formatter.dateStyle = .medium
        formatter.timeStyle = .none
        return formatter.string(from: workout.date)
    }

    private var formattedDuration: String {
        let seconds = Int(workout.duration)
        let h = seconds / 3600
        let m = (seconds % 3600) / 60
        let s = seconds % 60
        if h > 0 {
            return String(format: "%d:%02d:%02d", h, m, s)
        }
        return String(format: "%d:%02d", m, s)
    }
}

// MARK: - Exercise Row

private struct WorkoutHistoryExerciseRow: View {
    let exercise: WorkoutHistoryExercise

    var body: some View {
        HStack {
            Image(systemName: exercise.isCompleted ? "checkmark.circle.fill" : "circle")
                .font(.caption)
                .foregroundStyle(exercise.isCompleted ? .green : .white.opacity(0.35))

            Text(exercise.name)
                .font(.subheadline)
                .foregroundStyle(exercise.isCompleted ? .white : .white.opacity(0.45))
                .lineLimit(1)

            Spacer()

            if exercise.sets > 0 {
                Text(exerciseSummary)
                    .font(.caption)
                    .foregroundStyle(.white.opacity(0.6))
            }
        }
    }

    private var exerciseSummary: String {
        if exercise.reps > 0 {
            return "\(exercise.sets) × \(exercise.reps) повт."
        }
        return "\(exercise.sets) подх."
    }
}

#Preview {
    NavigationStack {
        WorkoutHistoryView()
    }
}
