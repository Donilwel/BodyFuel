import SwiftUI

struct WorkoutExecutionView: View {
    @EnvironmentObject var viewModel: WorkoutViewModel
    @State private var showFinishAlert = false
    @FocusState private var repCountFocused: ExerciseStatsField?

    private enum ExerciseStatsField: Hashable {
        case repCount
    }

    var body: some View {
        GeometryReader { geometry in
            ZStack(alignment: .top) {
                AnimatedBackground()
                    .ignoresSafeArea()
                ScrollView {
                    VStack(alignment: .center, spacing: 24) {
                        ProgressView(value: viewModel.workoutProgress)
                            .progressViewStyle(LinearProgressViewStyle(tint: .white.opacity(0.7)))

                        HStack {
                            Text(
                                String(
                                    format: "%02d:%02d",
                                    viewModel.totalWorkoutElapsedTime / 60,
                                    viewModel.totalWorkoutElapsedTime % 60
                                )
                            )
                            .font(.headline)
                            .foregroundColor(.white)

                            Spacer()

                            if viewModel.phase != .finished {
                                Button {
                                    viewModel.togglePause()
                                } label: {
                                    Image(systemName: viewModel.isPaused ? "play.fill" : "pause.fill")
                                        .font(.body)
                                        .foregroundColor(.white)
                                        .padding(.trailing, 12)
                                }
                            }

                            Button("Завершить", systemImage: "xmark") {
                                showFinishAlert = true
                            }
                            .labelStyle(.iconOnly)
                            .tint(.white)
                        }

                        if viewModel.isPaused {
                            Text("Пауза")
                                .font(.headline)
                                .foregroundColor(.white.opacity(0.7))
                        } else {
                            Text(viewModel.phaseTitle)
                                .font(.headline)
                                .foregroundColor(.white)
                        }

                        if let exercise = viewModel.currentExercise {
                            ExerciseCard(exercise: exercise)
                                .id(viewModel.currentExerciseIndex)

                            if exercise.type == .cardio && viewModel.phase == .exercise || viewModel.phase == .restBetweenExercises || viewModel.phase == .restBetweenSets {
                                ZStack {
                                    RingDiagramView(progress: viewModel.progress)
                                        .frame(width: 150, height: 150)

                                    Text(timeRemainingString(viewModel.timeRemaining))
                                        .font(.title2)
                                        .foregroundColor(.white)
                                }
                            } else if viewModel.phase == .exercise {
                                ValidatedField(error: viewModel.currentExerciseRepCountError) {
                                    CustomTextField(
                                        title: "Количество сделанных повторов",
                                        field: ExerciseStatsField.repCount,
                                        focusedField: $repCountFocused,
                                        text: $viewModel.currentExerciseRepCount
                                    )
                                }
                            }

                            Text("Подход \(viewModel.currentSet) / \(exercise.setCount)")
                                .foregroundColor(.white)
                        }
                        controls
                    }
                    .padding()
                }
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
        }
        .navigationBarBackButtonHidden()
        .toolbar(.hidden, for: .tabBar)
        .toolbarBackground(.hidden, for: .navigationBar)
        .alert("Завершить тренировку?", isPresented: $showFinishAlert) {
            Button("Отмена") {
                showFinishAlert = false
            }
            Button("Завершить") {
                viewModel.skipWorkout()
            }
        } message: {
            Text("Вы уверены, что хотите завершить тренировку?")
        }
        .onTapGesture {
            repCountFocused = nil
        }
    }

    private var controls: some View {
        VStack(spacing: 16) {
            if viewModel.isPaused {
                PrimaryButton(title: "Продолжить") {
                    viewModel.togglePause()
                }
            } else {
                switch viewModel.phase {
                case .waitingForStart:
                    PrimaryButton(title: "Начать упражнение") {
                        viewModel.startExercise()
                    }
                    SecondaryButton(title: "Пропустить упражнение") {
                        viewModel.skipExercise()
                    }
                case .exercise:
                    PrimaryButton(title: "Завершить подход") {
                        viewModel.moveToNextPhase()
                    }
                    SecondaryButton(title: "Пропустить упражнение") {
                        viewModel.skipExercise()
                    }
                case .restBetweenExercises:
                    PrimaryButton(title: viewModel.isLastSet ? "Завершить тренировку" : "Следующее упражнение") {
                        viewModel.moveToNextPhase()
                    }
                case .restBetweenSets:
                    PrimaryButton(title: "Следующий подход") {
                        viewModel.moveToNextPhase()
                    }
                case .finished:
                    Text("Тренировка завершена")
                        .foregroundColor(.white)
                }
            }
        }
    }

    private func timeRemainingString(_ value: Int) -> String {
        if abs(value) < 60 {
            return "\(value) сек"
        } else {
            return "\(value / 60) мин \(abs(value) % 60) сек"
        }
    }
}

// MARK: - Exercise Card

private struct ExerciseCard: View {
    let exercise: Exercise
    @State private var isDescriptionExpanded = false

    var body: some View {
        VStack(alignment: .center, spacing: 16) {
            Text(exercise.name)
                .font(.title2)
                .foregroundColor(.white)
                .multilineTextAlignment(.center)

            if exercise.type == .cardio {
                Text("\(exercise.setCount) \(setsLabel(exercise.setCount)), \(exercise.duration.formattedTime) каждый")
                    .font(.title3)
                    .foregroundColor(.white)
                    .multilineTextAlignment(.center)
            } else if let repCount = exercise.repCount {
                Text("\(exercise.setCount) подхода по \(repCount) раз")
                    .font(.title3)
                    .foregroundColor(.white)
                    .multilineTextAlignment(.center)
            }

            VStack(spacing: 8) {
                Text(exercise.description)
                    .foregroundColor(.white)
                    .multilineTextAlignment(.center)
                    .lineLimit(isDescriptionExpanded ? nil : 3)

                Button {
                    withAnimation(.easeInOut(duration: 0.25)) {
                        isDescriptionExpanded.toggle()
                    }
                } label: {
                    Image(systemName: isDescriptionExpanded ? "chevron.up" : "chevron.down")
                        .font(.caption)
                        .foregroundColor(.white.opacity(0.55))
                        .padding(.vertical, 4)
                }
            }
        }
        .frame(maxWidth: .infinity)
        .padding()
        .background(.ultraThinMaterial)
        .cornerRadius(24)
    }

    private func setsLabel(_ count: Int) -> String {
        let rem10 = count % 10
        let rem100 = count % 100
        if rem10 == 1 && rem100 != 11 { return "подход" }
        if rem10 >= 2 && rem10 <= 4 && (rem100 < 10 || rem100 >= 20) { return "подхода" }
        return "подходов"
    }
}
