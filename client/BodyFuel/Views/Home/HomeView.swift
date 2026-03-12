import SwiftUI
import WidgetKit

struct HomeView: View {
    @StateObject private var viewModel = HomeViewModel()
    @State private var path = NavigationPath()
    
    private let sharedWidgetStorage = SharedWidgetStorage.shared

    var body: some View {
        NavigationStack(path: $path) {
            ZStack {
                AnimatedBackground()

                VStack(spacing: 20) {
                    caloriesRingBlock
                    workoutCard
                    nutritionCard
                    
                    Spacer()
                }
                .padding()
            }
            .task {
                await viewModel.load()
                WidgetCenter.shared.reloadAllTimelines()
            }
            .navigationDestination(for: String.self) { route in
                if route == "workouts" { Text("Экран тренировок") }
                if route == "nutrition" { Text("Экран питания") }
            }
        }
        .refreshable {
            Task {
                await viewModel.load()
                WidgetCenter.shared.reloadAllTimelines()
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
        Group {
            if let workout = viewModel.workout {
                VStack(alignment: .leading, spacing: 12) {
                    Text("Тренировка сегодня")
                        .font(.headline)
                        .foregroundColor(.white)

                    Text(workout.name)
                        .font(.title3.bold())
                        .foregroundColor(.white)

                    Text("\(workout.duration) мин • ~\(workout.calories) ккал")
                        .foregroundColor(.white.opacity(0.7))

                    HStack {
                        PrimaryButton(title: "Начать") {
                            path.append("workouts")
                        }
                        SecondaryButton(title: "Выбрать другую") {
                            path.append("workouts")
                        }
                    }
                }
                .cardStyle()
            }
        }
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
