import SwiftUI

struct FoodView: View {
    @StateObject private var viewModel = FoodViewModel()
    @State private var expandedSections: Set<MealType> = []
    @State private var showAddOptions = false

    var body: some View {
        NavigationStack {
            ZStack(alignment: .bottomTrailing) {
                AnimatedBackground()
                    .ignoresSafeArea()

                ScrollView {
                    VStack(spacing: 20) {
//                        Text("Питание")
//                            .foregroundColor(.white)
//                            .font(.largeTitle.bold())
                        
                        switch viewModel.screenState {
                        case .loading:
                            ProgressView()
                                .tint(.white)
                                .padding(.top, 60)

                        case .error(let message):
                            VStack(spacing: 12) {
                                Text(message)
                                    .foregroundColor(.white.opacity(0.8))
                                    .multilineTextAlignment(.center)
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
            .navigationTitle("Питание")
            .navigationBarTitleDisplayMode(.large)
            .toolbarBackground(.hidden, for: .navigationBar)
        }
        .task {
            await viewModel.load()
        }
        .sheet(isPresented: $viewModel.showAddMeal) {
            AddMealView()
                .environmentObject(viewModel)
        }
        .fullScreenCover(isPresented: $viewModel.showCamera) {
            CameraFoodView()
                .environmentObject(viewModel)
        }
        .sheet(isPresented: $viewModel.showRecipes) {
            RecipesView(recipes: viewModel.recipes, isLoading: viewModel.isLoadingRecipes)
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
                        viewModel.addMealType = currentBestMealType()
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
                        viewModel.addMealType = currentBestMealType()
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
                    .background(AppColors.gradient)
                    .clipShape(Circle())
                    .shadow(radius: 8)
                    .rotationEffect(.degrees(showAddOptions ? 45 : 0))
            }
        }
        .padding()
    }

    private func currentBestMealType() -> MealType {
        let hour = Calendar.current.component(.hour, from: Date())
        switch hour {
        case 6..<11: return .breakfast
        case 11..<16: return .lunch
        case 16..<20: return .dinner
        default: return .snack
        }
    }
}

// MARK: - Macro Summary Card

struct MacroSummaryCard: View {
    let summary: NutritionDailySummary

    var body: some View {
        VStack(spacing: 16) {
            HStack(spacing: 20) {
                ZStack {
                    MacroRingView(summary: summary)
                        .frame(width: 140, height: 140)

                    VStack(spacing: 2) {
                        Text("\(summary.remainingCalories)")
                            .font(.title2.bold())
                            .foregroundColor(.white)
                        Text("ккал осталось")
                            .font(.caption)
                            .foregroundColor(.white.opacity(0.7))
                            .multilineTextAlignment(.center)
                    }
                }

                VStack(alignment: .leading, spacing: 10) {
                    MacroLegendRow(label: "Белки", value: summary.consumed.protein, goal: summary.goal.protein, color: .blue)
                    MacroLegendRow(label: "Жиры", value: summary.consumed.fat, goal: summary.goal.fat, color: .orange)
                    MacroLegendRow(label: "Углеводы", value: summary.consumed.carbs, goal: summary.goal.carbs, color: .green)
                }

                Spacer()
            }

            HStack {
                VStack {
                    Text("\(summary.consumed.calories)")
                        .font(.subheadline.bold())
                        .foregroundColor(.white)
                    Text("потреблено")
                        .font(.caption)
                        .foregroundColor(.white.opacity(0.7))
                }
                Spacer()
                VStack {
                    Text("\(summary.goal.calories)")
                        .font(.subheadline.bold())
                        .foregroundColor(.white)
                    Text("цель")
                        .font(.caption)
                        .foregroundColor(.white.opacity(0.7))
                }
                Spacer()
                VStack {
                    Text("\(summary.burned)")
                        .font(.subheadline.bold())
                        .foregroundColor(.white)
                    Text("сожжено")
                        .font(.caption)
                        .foregroundColor(.white.opacity(0.7))
                }
            }
        }
        .padding()
        .background(.ultraThinMaterial)
        .cornerRadius(24)
    }
}

struct MacroLegendRow: View {
    let label: String
    let value: Double
    let goal: Double
    let color: Color

    var body: some View {
        HStack(spacing: 8) {
            Circle()
                .fill(color)
                .frame(width: 8, height: 8)
            VStack(alignment: .leading, spacing: 2) {
                Text(label)
                    .font(.caption)
                    .foregroundColor(.white.opacity(0.7))
                Text("\(Int(value)) / \(Int(goal)) г")
                    .font(.caption.bold())
                    .foregroundColor(.white)
            }
        }
    }
}

// MARK: - Macro Ring

struct MacroRingView: View {
    let summary: NutritionDailySummary

    private let ringWidth: CGFloat = 12

    var body: some View {
        GeometryReader { geometry in
            let size = min(geometry.size.width, geometry.size.height)
            let radius = size / 2

            ZStack {
                Circle()
                    .stroke(Color.white.opacity(0.15), lineWidth: ringWidth)

                MacroArcShape(startFraction: 0, endFraction: summary.proteinProgress * 0.333)
                    .stroke(Color.blue, style: StrokeStyle(lineWidth: ringWidth, lineCap: .round))

                MacroArcShape(startFraction: 0.333, endFraction: 0.333 + summary.fatProgress * 0.333)
                    .stroke(Color.orange, style: StrokeStyle(lineWidth: ringWidth, lineCap: .round))

                MacroArcShape(startFraction: 0.667, endFraction: 0.667 + summary.carbsProgress * 0.333)
                    .stroke(Color.green, style: StrokeStyle(lineWidth: ringWidth, lineCap: .round))
            }
            .padding(ringWidth / 2)
        }
    }
}

struct MacroArcShape: Shape {
    var startFraction: Double
    var endFraction: Double

    func path(in rect: CGRect) -> Path {
        var path = Path()
        let center = CGPoint(x: rect.midX, y: rect.midY)
        let radius = min(rect.width, rect.height) / 2
        let startAngle = Angle(degrees: -90 + startFraction * 360)
        let endAngle = Angle(degrees: -90 + endFraction * 360)
        path.addArc(center: center, radius: radius, startAngle: startAngle, endAngle: endAngle, clockwise: false)
        return path
    }
}

// MARK: - Meal Section

struct MealSectionView: View {
    let mealType: MealType
    let meals: [Meal]
    let isExpanded: Bool
    let onToggle: () -> Void
    let onAdd: () -> Void

    private var totalCalories: Int {
        meals.reduce(0) { $0 + $1.macros.calories }
    }

    var body: some View {
        VStack(spacing: 0) {
            Button(action: onToggle) {
                HStack {
                    Image(systemName: mealType.iconName)
                        .foregroundColor(.white.opacity(0.8))
                        .frame(width: 24)
                    Text(mealType.displayName)
                        .font(.subheadline.weight(.semibold))
                        .foregroundColor(.white)
                    Spacer()
                    Text("\(totalCalories) ккал")
                        .font(.subheadline)
                        .foregroundColor(.white.opacity(0.8))
                    Image(systemName: isExpanded ? "chevron.up" : "chevron.down")
                        .font(.caption)
                        .foregroundColor(.white.opacity(0.6))
                        .padding(.leading, 4)
                }
                .padding()
            }

            if isExpanded {
                VStack(spacing: 0) {
                    Divider().background(Color.white.opacity(0.2))

                    ForEach(meals) { meal in
                        MealRowView(meal: meal)
                        if meal.id != meals.last?.id {
                            Divider()
                                .background(Color.white.opacity(0.1))
                                .padding(.horizontal)
                        }
                    }

                    Divider().background(Color.white.opacity(0.2))

                    Button(action: onAdd) {
                        HStack {
                            Image(systemName: "plus.circle")
                            Text("Добавить блюдо")
                        }
                        .font(.subheadline)
                        .foregroundColor(.white.opacity(0.7))
                        .padding()
                    }
                }
            }
        }
        .background(.ultraThinMaterial)
        .cornerRadius(16)
        .animation(.easeInOut(duration: 0.25), value: isExpanded)
    }
}

struct MealRowView: View {
    let meal: Meal
    @State private var showMacros = false

    var body: some View {
        VStack(spacing: 0) {
            Button {
                withAnimation(.easeInOut(duration: 0.2)) {
                    showMacros.toggle()
                }
            } label: {
                HStack {
                    VStack(alignment: .leading, spacing: 2) {
                        Text(meal.name)
                            .font(.subheadline)
                            .foregroundColor(.white)
                            .multilineTextAlignment(.leading)
                        Text(meal.time, style: .time)
                            .font(.caption)
                            .foregroundColor(.white.opacity(0.5))
                    }
                    Spacer()
                    Text("\(meal.macros.calories) ккал")
                        .font(.subheadline)
                        .foregroundColor(.white.opacity(0.9))
                }
                .padding(.horizontal)
                .padding(.vertical, 12)
            }

            if showMacros {
                MacroRowView(macros: meal.macros)
                    .padding(.horizontal)
                    .padding(.bottom, 12)
            }
        }
    }
}

struct MacroRowView: View {
    let macros: MacroNutrients

    var body: some View {
        HStack(spacing: 16) {
            MacroChip(label: "Б", value: macros.protein, color: .blue)
            MacroChip(label: "Ж", value: macros.fat, color: .orange)
            MacroChip(label: "У", value: macros.carbs, color: .green)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }
}

struct MacroChip: View {
    let label: String
    let value: Double
    let color: Color

    var body: some View {
        HStack(spacing: 4) {
            Text(label)
                .font(.caption.bold())
                .foregroundColor(color)
            Text("\(Int(value)) г")
                .font(.caption)
                .foregroundColor(.white.opacity(0.8))
        }
        .padding(.horizontal, 8)
        .padding(.vertical, 4)
        .background(color.opacity(0.15))
        .cornerRadius(8)
    }
}

// MARK: - Recipes View

struct RecipesView: View {
    let recipes: [Recipe]
    let isLoading: Bool
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            ZStack {
                AnimatedBackground()
                    .ignoresSafeArea()

                if isLoading {
                    ProgressView()
                        .tint(.white)
                } else {
                    ScrollView {
                        VStack(spacing: 16) {
                            ForEach(recipes) { recipe in
                                RecipeCard(recipe: recipe)
                            }
                        }
                        .padding()
                    }
                }
            }
            .navigationTitle("Рекомендуемые рецепты")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Закрыть") { dismiss() }
                        .foregroundColor(.white)
                }
            }
            .toolbarBackground(.hidden, for: .navigationBar)
        }
    }
}

struct RecipeCard: View {
    let recipe: Recipe

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text(recipe.name)
                    .font(.headline)
                    .foregroundColor(.white)
                Spacer()
                Label("\(recipe.preparationTime) мин", systemImage: "clock")
                    .font(.caption)
                    .foregroundColor(.white.opacity(0.7))
            }

            Text(recipe.description)
                .font(.subheadline)
                .foregroundColor(.white.opacity(0.8))
                .fixedSize(horizontal: false, vertical: true)

            MacroRowView(macros: recipe.macros)
        }
        .cardStyle()
    }
}

#Preview {
    FoodView()
}
