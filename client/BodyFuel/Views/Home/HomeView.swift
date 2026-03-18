import SwiftUI
import WidgetKit

struct HomeView: View {
    @EnvironmentObject var workoutViewModel: WorkoutViewModel
    @StateObject private var viewModel = HomeViewModel()
    @State private var path = NavigationPath()
    
    private let sharedWidgetStorage = SharedWidgetStorage.shared

    var body: some View {
        NavigationStack(path: $path) {
            ZStack {
                AnimatedBackground()
                
                ScrollView {
                    VStack(spacing: 20) {
                        caloriesRingBlock
                        workoutCard
                        nutritionCard
                        
                        Spacer()
                    }
                    .padding()
                }
            }
            .task {
                await viewModel.load()
                await workoutViewModel.load()
                WidgetCenter.shared.reloadAllTimelines()
            }
            .navigationDestination(for: String.self) { route in
                if route == "workouts" { Text("Экран тренировок") }
                if route == "nutrition" { Text("Экран питания") }
            }
            .navigationDestination(isPresented: $workoutViewModel.isWorkoutActive) {
                WorkoutExecutionView()
                    .environmentObject(workoutViewModel)
            }
            .navigationDestination(isPresented: $workoutViewModel.showWorkoutSummary) {
                WorkoutSummaryView()
                    .environmentObject(workoutViewModel)
            }
            .refreshable {
                Task {
                    await viewModel.load()
                    WidgetCenter.shared.reloadAllTimelines()
                }
            }
            .onChange(of: workoutViewModel.shouldStartFromDeepLink) { newValue in
                if newValue {
                    workoutViewModel.shouldStartFromDeepLink = false
                    workoutViewModel.startWorkout()
                }
            }
        }
    }
    
    private var caloriesRingBlock: some View {
        VStack {
            if let stats = viewModel.stats,
                let goals = viewModel.goals,
                let basalMetabolicRate = viewModel.basalMetabolicRate {
                CaloriesRingProgressView(
                    consumed: stats.caloriesConsumed,
                    goal: goals.calories,
                    burned: stats.caloriesBurned,
                    basalMetabolicRate: basalMetabolicRate
                )
            }
        }
    }

    private var workoutCard: some View {
        WorkoutCardView(
            workout: workoutViewModel.recommendedWorkout,
            startAction: workoutViewModel.startWorkout,
            changeAction: workoutViewModel.changeWorkout
        )
    }

    private var nutritionCard: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Питание сегодня")
                .font(.headline)
                .foregroundColor(.white)
            
            ForEach(viewModel.meals) { meal in
                HStack {
                    Text(meal.title)
                    Spacer()
                    Text("\(meal.calories) ккал")
                }
                .foregroundColor(.white.opacity(0.8))
            }
            
            HStack {
                PrimaryButton(title: "Добавить приём пищи") {
                    path.append("nutrition")
                }
            }
        }
        .cardStyle()
    }

}

#Preview {
    TabBarView()
}
