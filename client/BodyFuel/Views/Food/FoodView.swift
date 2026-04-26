import SwiftUI

struct FoodView: View {
    @StateObject private var viewModel = FoodViewModel()
    @ObservedObject private var router = AppRouter.shared
    @State private var expandedSections: Set<MealType> = []
    @State private var showAddOptions = false

    var body: some View {
        NavigationStack {
            ZStack(alignment: .bottomTrailing) {
                AnimatedBackground()
                    .ignoresSafeArea()

                ScrollView() {
                    VStack(alignment: .center, spacing: 20) {
                        switch viewModel.screenState {
                        case .loading, .idle:
                            EmptyView()

                        case .error(let message):
                            VStack(spacing: 12) {
                                Text(message)
                                    .foregroundColor(.white.opacity(0.8))
                                    .multilineTextAlignment(.center)
                                    .frame(maxWidth: .infinity)
                                PrimaryButton(title: "Повторить") {
                                    Task { await viewModel.load() }
                                }
                            }
                            .padding(.top, 60)

                        case .loaded:
                            if let summary = viewModel.dailySummary {
                                MacroSummaryCard(summary: summary)
                            }

                            diarySection

                            recipesButton
                        }
                    }
                    .padding()
                    .padding(.bottom, 100)
                }
                
                addFoodButton
            }
            .screenLoading(viewModel.screenState == .loading)
        }
        .task {
            await viewModel.load()
        }
        .refreshable {
            await viewModel.load()
        }
        .onChange(of: router.pendingAddMeal) { pending in
            if pending {
                viewModel.showAddMeal = true
                router.pendingAddMeal = false
            }
        }
        .sheet(isPresented: $viewModel.showAddMeal) {
            AddMealView()
                .environmentObject(viewModel)
                .presentationDetents([.large])
        }
        .fullScreenCover(isPresented: $viewModel.showCamera) {
            CameraFoodView()
                .environmentObject(viewModel)
        }
        .sheet(isPresented: $viewModel.showRecipes) {
            RecipesView(recipes: viewModel.recipes, isLoading: viewModel.isLoadingRecipes) { recipe in
                // TODO: сделать промежуточный оверлей с описанием и ингредиентами
                let meal = Meal(name: recipe.name, mealType: currentMealType(), macros: recipe.macros)
                
                Task { await viewModel.saveMeal(meal) }
            }
        }
    }

    private var diarySection: some View {
        VStack(spacing: 12) {
            HStack {
                Text("Дневник питания")
                    .font(.headline)
                    .foregroundColor(.white)
                Spacer()
            }

            if viewModel.mealsByType.isEmpty {
                Text("Добавьте первый приём пищи")
                    .foregroundColor(.white.opacity(0.6))
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(.ultraThinMaterial)
                    .cornerRadius(16)
            } else {
                ForEach(viewModel.mealsByType, id: \.0) { mealType, meals in
                    MealSectionView(
                        mealType: mealType,
                        meals: meals,
                        isExpanded: expandedSections.contains(mealType),
                        onToggle: {
                            if expandedSections.contains(mealType) {
                                expandedSections.remove(mealType)
                            } else {
                                expandedSections.insert(mealType)
                            }
                        },
                        onAdd: {
                            viewModel.addMealType = mealType
                            viewModel.showAddMeal = true
                        },
                        onDelete: { meal in
                            Task { await viewModel.deleteMeal(meal) }
                        }
                    )
                }
            }
        }
    }

    private var recipesButton: some View {
        Button {
            Task { await viewModel.loadRecipes() }
        } label: {
            HStack {
                Image(systemName: "wand.and.stars")
                Text("Рекомендовать рецепты")
                Spacer()
                if viewModel.isLoadingRecipes {
                    ProgressView().tint(.white)
                } else {
                    Image(systemName: "chevron.right")
                }
            }
            .foregroundColor(.white)
            .padding()
            .background(.ultraThinMaterial)
            .cornerRadius(16)
        }
        .disabled(viewModel.isLoadingRecipes)
    }

    private var addFoodButton: some View {
        VStack(spacing: 12) {
            if showAddOptions {
                VStack(spacing: 10) {
                    Button {
                        showAddOptions = false
                        viewModel.addMealType = currentMealType()
                        viewModel.showCamera = true
                    } label: {
                        Label("Сфотографировать", systemImage: "camera.fill")
                            .font(.subheadline.weight(.semibold))
                            .foregroundColor(.white)
                            .padding(.horizontal, 16)
                            .padding(.vertical, 12)
                            .background(.ultraThinMaterial)
                            .cornerRadius(12)
                    }

                    Button {
                        showAddOptions = false
                        viewModel.addMealType = currentMealType()
                        viewModel.showAddMeal = true
                    } label: {
                        Label("Описать текстом", systemImage: "text.bubble.fill")
                            .font(.subheadline.weight(.semibold))
                            .foregroundColor(.white)
                            .padding(.horizontal, 16)
                            .padding(.vertical, 12)
                            .background(.ultraThinMaterial)
                            .cornerRadius(12)
                    }
                }
                .transition(.move(edge: .bottom).combined(with: .opacity))
            }

            Button {
                withAnimation(.spring(response: 0.35, dampingFraction: 0.7)) {
                    showAddOptions.toggle()
                }
            } label: {
                Image(systemName: showAddOptions ? "xmark" : "plus")
                    .font(.title2.weight(.semibold))
                    .foregroundColor(.white)
                    .frame(width: 56, height: 56)
                    .background(.ultraThinMaterial)
                    .clipShape(Circle())
                    .shadow(radius: 8)
                    .rotationEffect(.degrees(showAddOptions ? 45 : 0))
            }
        }
        .padding()
    }

    private func currentMealType() -> MealType {
        let hour = Calendar.current.component(.hour, from: Date())
        switch hour {
        case 6..<11: return .breakfast
        case 11..<16: return .lunch
        case 16..<20: return .dinner
        default: return .snack
        }
    }
}

#Preview {
    TabBarView()
        .environmentObject(AppRouter.shared)
        .environmentObject(WorkoutViewModel())
}
