import SwiftUI
import WidgetKit

struct HomeView: View {
    @EnvironmentObject var workoutViewModel: WorkoutViewModel
    @ObservedObject private var router = AppRouter.shared
    @ObservedObject private var nutritionStore = NutritionStore.shared
    @StateObject private var viewModel = HomeViewModel()
    @State private var path = NavigationPath()
    
    private let sharedWidgetStorage = SharedWidgetStorage.shared

    var body: some View {
        NavigationStack(path: $path) {
            ZStack {
                AnimatedBackground()
                    .ignoresSafeArea()

                ScrollView {
                    VStack(spacing: 20) {
                        if nutritionStore.isDataStale {
                            staleDataBanner
                        }
                        caloriesRingBlock
                        workoutCard
                        nutritionCard

                        Spacer()
                    }
                }
            }
            .screenLoading(viewModel.state == .loading)
            .task {
                await withTaskGroup(of: Void.self) { group in
                    group.addTask { try? await viewModel.load() }
                    group.addTask { await workoutViewModel.load() }
                }
                Task { await HealthKitService.shared.startBackgroundObservers() }
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
                    NutritionStore.shared.invalidate()
                    UserStore.shared.invalidateProfile()
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
            .sheet(isPresented: $workoutViewModel.showWorkoutFilter) {
                WorkoutFilterView(viewModel: workoutViewModel)
            }
            .alert("Нет доступа к Здоровью", isPresented: $workoutViewModel.showHealthPermissionAlert) {
                Button("Открыть настройки") {
                    if let url = URL(string: UIApplication.openSettingsURLString) {
                        UIApplication.shared.open(url)
                    }
                }
                Button("Отмена", role: .cancel) {}
            } message: {
                Text("Для полноценной работы тренировки необходимо разрешение на доступ к Здоровью. Без него тренировку начать не получится.")
            }
        }
    }
    
    private var staleDataBanner: some View {
        HStack(spacing: 6) {
            Image(systemName: "clock.arrow.circlepath")
                .font(.caption)
            Text("Данные могут быть устаревшими")
                .font(.caption)
        }
        .padding(.top, 6)
        .foregroundStyle(.white.opacity(0.65))
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    private var caloriesRingBlock: some View {
        CaloriesRingProgressView(
            consumed: viewModel.stats?.caloriesConsumed ?? 0,
            goal: viewModel.goals?.calories ?? 0,
            burned: viewModel.stats?.caloriesBurned ?? 0,
            basalMetabolicRate: max(viewModel.basalMetabolicRate ?? 1500, 1)
        )
    }

    private var workoutCard: some View {
        VStack(spacing: 6) {
            if workoutViewModel.isWorkoutStale {
                HStack(spacing: 6) {
                    Image(systemName: "clock.arrow.circlepath").font(.caption)
                    Text("Сохраненная тренировка")
                        .font(.caption)
                }
                .foregroundStyle(.white.opacity(0.65))
                .frame(maxWidth: .infinity, alignment: .leading)
            }
            WorkoutCardView(
                workout: workoutViewModel.recommendedWorkout,
                isChanging: workoutViewModel.screenState == .loading,
                startAction: workoutViewModel.startWorkout,
                changeAction: workoutViewModel.changeWorkout
            )
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
                    guard let userId = UserSessionManager.shared.currentUserId,
                          UserSessionManager.shared.authToken(for: userId) != nil else {
                        router.logout()
                        return
                    }
                    router.selectedTab = .food
                    router.pendingAddMeal = true
                }
            }
        }
        .cardStyle()
    }

}

#Preview {
//    HomeView()
    TabBarView()
        .environmentObject(AppRouter.shared)
        .environmentObject(WorkoutViewModel())
}
