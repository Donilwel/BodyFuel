import SwiftUI
import WidgetKit
import PhotosUI

struct ProfileView: View {
    @ObservedObject var router = AppRouter.shared
    
    @StateObject private var viewModel = ProfileViewModel()
    
    @FocusState private var parametersFocused: ParameterField?

    private enum ParameterField: Hashable {
        case height
        case currentWeight
        case targetWeight
        case targetWorkoutsWeekly
    }
    
    private var parametersInfoCard: some View {
        InfoCard {
            ValidatedField(
                error: viewModel.isEditing ? viewModel.heightError : nil
            ) {
                EditableTextField(
                    title: "Рост",
                    value: $viewModel.height,
                    suffix: "см",
                    isEditing: viewModel.isEditing,
                    field: ParameterField.height,
                    focusedField: $parametersFocused
                )
            }
            ValidatedField(error: viewModel.isEditing ? viewModel.weightError : nil) {
                EditableTextField(
                    title: "Текущий вес",
                    value: $viewModel.weight,
                    suffix: "кг",
                    isEditing: viewModel.isEditing,
                    field: ParameterField.currentWeight,
                    focusedField: $parametersFocused
                )
            }
            ValidatedField(error: viewModel.isEditing ? viewModel.targetWeightError : nil) {
                EditableTextField(
                    title: "Целевой вес",
                    value: $viewModel.targetWeight,
                    suffix: "кг",
                    isEditing: viewModel.isEditing,
                    field: ParameterField.targetWeight,
                    focusedField: $parametersFocused
                )
            }
        }
    }
    
    private var routineInfoCard: some View {
        InfoCard {
            CustomTextView(
                title: "Калорий в день",
                value: $viewModel.targetCaloriesDaily,
                suffix: "ккал"
            )
            .onTapGesture {
                viewModel.caloriesFormState = .preview
                viewModel.showCaloriesSheet = true
            }
            
            ValidatedField(error: viewModel.isEditing ? viewModel.targetWorkoutsError : nil) {
                EditableTextField(
                    title: "Тренировок в неделю",
                    value: $viewModel.targetWorkoutsWeekly,
                    isEditing: viewModel.isEditing,
                    field: ParameterField.targetWorkoutsWeekly,
                    focusedField: $parametersFocused
                )
            }
        }
    }
    
    private var lifestyleInfoCard: some View {
        InfoCard {
            if viewModel.isEditing {
                CustomPickerField(
                    title: "Образ жизни",
                    options: Lifestyle.allCases,
                    optionTitle: \.title,
                    selection: Binding(
                        get: { Optional(viewModel.lifestyle) },
                        set: { if let v = $0 { viewModel.lifestyle = v } }
                    )
                )
                CustomPickerField(
                    title: "Спортивная подготовка",
                    options: FitnessLevel.allCases,
                    optionTitle: \.title,
                    selection: Binding(
                        get: { Optional(viewModel.fitnessLevel) },
                        set: { if let v = $0 { viewModel.fitnessLevel = v } }
                    )
                )
                CustomPickerField(
                    title: "Цель",
                    options: MainGoal.allCases,
                    optionTitle: \.title,
                    selection: Binding(
                        get: { Optional(viewModel.goal) },
                        set: { if let v = $0 { viewModel.goal = v } }
                    )
                )
            } else {
                EditablePickerView(title: "Образ жизни", value: viewModel.lifestyle.title) {}
                EditablePickerView(title: "Спортивная подготовка", value: viewModel.fitnessLevel.title) {}
                EditablePickerView(title: "Цель", value: viewModel.goal.title) {}
            }
        }
    }

    var body: some View {
        ZStack {
            AnimatedBackground()
                .ignoresSafeArea()

            ScrollView {
                VStack(spacing: -40) {
                    HStack {
                        Spacer()
                        SecondaryButton(
                            title: viewModel.isEditing ? "Сохранить" : "Изменить",
                            isLoading: false
                        ) {
                            if viewModel.isEditing {
                                Task {
                                    await viewModel.save()
                                    WidgetCenter.shared.reloadAllTimelines()
                                }
                            } else {
                                viewModel.isEditing = true
                            }
                        }
                    }
                    
                    VStack(spacing: 24) {
                        PhotosPicker(selection: $viewModel.avatarItem, matching: .images) {
                            if let data = viewModel.avatarData {
                                AvatarPickerView(data: data)
                            } else {
                                AvatarView(photoURL: viewModel.avatarUrl)
                            }
                        }
                        .onChange(of: viewModel.avatarItem) { _ in
                            Task { await viewModel.loadAvatar() }
                        }

                        parametersInfoCard
                        
                        routineInfoCard
                        
                        lifestyleInfoCard
                        
                        PrimaryButton(
                            title: "Выйти",
                            isLoading: viewModel.screenState == .loading
                        ) {
                            viewModel.logout()
                        }
                        
                        SecondaryButton(
                            title: "Удалить профиль",
                            isLoading: viewModel.screenState == .loading
                        ) {
                            Task {
                                await viewModel.deleteProfile()
                            }
                        }
                    }
                    .padding(.vertical, 40)
                }
                .padding(.horizontal)
            }
        }
        .screenLoading(viewModel.screenState == .loading && viewModel.profile == nil)
        .task { await viewModel.load() }
        .sheet(isPresented: $viewModel.showCaloriesSheet) {
            CaloriesRecalculationSheet(viewModel: viewModel)
        }
        .alert("Ошибка", isPresented: .constant(isError)) {
            Button("Отменить") {
                viewModel.screenState = .idle
            }
            Button("ОК") {
                viewModel.screenState = .idle
                viewModel.isEditing = false
            }
        } message: {
            if case let .error(message) = viewModel.screenState {
                Text(message)
            }
        }
        .onChange(of: viewModel.event) { event in
            switch event {
            case .logoutSuccess:
                router.logout()
            default:
                break
            }

            viewModel.event = .idle
        }
        .onTapGesture {
            parametersFocused = nil
        }
    }
    
    private var isError: Bool {
        if case .error = viewModel.screenState { return true }
        return false
    }
}

// MARK: - CaloriesRecalculationSheet

private struct CaloriesRecalculationSheet: View {
    @ObservedObject var viewModel: ProfileViewModel

    var body: some View {
        ZStack {
            AnimatedBackground().ignoresSafeArea()

            ScrollView {
                VStack(spacing: 24) {
                    Text("Норма калорий")
                        .font(.title2.bold())
                        .foregroundStyle(.white)
                        .frame(maxWidth: .infinity, alignment: .leading)

                    switch viewModel.caloriesFormState {
                    case .preview:
                        previewContent
                    case .counting:
                        countingContent
                    case .editing:
                        editingContent
                    }
                }
                .padding(24)
            }
        }
    }

    private var previewContent: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text("Пересчитаем вашу суточную норму калорий на основе текущих данных профиля")
                .foregroundStyle(.white.opacity(0.8))
                .fixedSize(horizontal: false, vertical: true)

            PrimaryButton(title: "Рассчитать", isLoading: false) {
                Task { await viewModel.countSheetCalories() }
            }
        }
    }

    private var countingContent: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text("Чтобы безопасно двигаться к цели, мы рекомендуем придерживаться")
                .foregroundStyle(.white)
                .font(.headline)
                .fixedSize(horizontal: false, vertical: true)

            Text("\(Int(viewModel.sheetDailyExpenditure)) ккал/день.")
                .foregroundStyle(.white)
                .font(.title2.bold())

            Text("Для этого вам ежедневно необходимо тратить примерно \(Int(viewModel.sheetDailyExpenditure - viewModel.sheetBasalMetabolicRate)) калорий")
                .foregroundStyle(.white)
                .fixedSize(horizontal: false, vertical: true)

            Text("Рассчитано на основе роста, веса, образа жизни, цели и данных приложения Здоровье")
                .foregroundStyle(.white)
                .font(.footnote)
                .fixedSize(horizontal: false, vertical: true)

            if let hint = viewModel.sheetHealthIntegrationError {
                Label(hint, systemImage: "exclamationmark.triangle")
                    .font(.footnote)
                    .foregroundStyle(.yellow.opacity(0.9))
                    .fixedSize(horizontal: false, vertical: true)
            }

            PrimaryButton(title: "Применить", isLoading: false) {
                Task { await viewModel.applySheetCalories() }
            }

            VStack(spacing: 4) {
                SecondaryButton(title: "Изменить норму") {
                    viewModel.caloriesFormState = .editing
                }
                SecondaryButton(title: "Рассчитать заново") {
                    Task { await viewModel.countSheetCalories() }
                }
            }
        }
    }

    private var editingContent: some View {
        VStack(alignment: .leading, spacing: 16) {
            CustomSliderField(
                title: "Количество калорий в день",
                from: viewModel.sheetDailyExpenditure * 0.5,
                to: viewModel.sheetDailyExpenditure * 1.5,
                step: 100,
                value: $viewModel.sheetTargetCalories
            )

            Text(viewModel.validateSheetCaloriesNorm())
                .font(.title3.bold())
                .foregroundStyle(.white)
                .fixedSize(horizontal: false, vertical: true)

            Text(viewModel.getSheetCaloriesNormHint())
                .font(.headline.bold())
                .foregroundStyle(.white)

            Text("Слишком сильный дефицит или профицит калорий может повлиять на ваше здоровье")
                .foregroundStyle(.white)
                .font(.footnote)
                .fixedSize(horizontal: false, vertical: true)

            VStack(spacing: 4) {
                PrimaryButton(title: "Применить", isLoading: false) {
                    Task { await viewModel.applySheetCalories() }
                }
                SecondaryButton(title: "Рассчитать заново") {
                    Task { await viewModel.countSheetCalories() }
                }
            }
        }
    }
}

#Preview {
    ProfileView()
}
