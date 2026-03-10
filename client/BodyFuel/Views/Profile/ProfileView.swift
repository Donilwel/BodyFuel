import SwiftUI

struct ProfileView: View {
    @ObservedObject var router = AppRouter.shared
    
    @StateObject private var viewModel = ProfileViewModel()
    
    @FocusState private var parametersFocused: ParameterField?
    
    @State private var isLifestylePickerPresented = false
    @State private var isGoalPickerPresented = false
    
    private enum ParameterField: Hashable {
        case height
        case currentWeight
        case targetWeight
        case targetWorkoutsWeekly
    }
    
    private var parametersInfoCard: some View {
        InfoCard {
            ValidatedField(
                error: viewModel.heightError
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
            ValidatedField(error: viewModel.weightError) {
                EditableTextField(
                    title: "Текущий вес",
                    value: $viewModel.weight,
                    suffix: "кг",
                    isEditing: viewModel.isEditing,
                    field: ParameterField.currentWeight,
                    focusedField: $parametersFocused
                )
            }
            ValidatedField(error: viewModel.targetWeightError) {
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
            ValidatedField(error: viewModel.targetWorkoutsError) {
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
            EditablePickerView(
                title: "Образ жизни",
                value: viewModel.lifestyle.title
            ) {
                if viewModel.isEditing {
                    isLifestylePickerPresented = true
                }
            }
            .confirmationDialog(
                "Образ жизни",
                isPresented: $isLifestylePickerPresented,
                titleVisibility: .hidden
            ) {
                ForEach(Lifestyle.allCases) { lifestyle in
                    Button(lifestyle.title) {
                        viewModel.lifestyle = lifestyle
                    }
                }
            }
            
            EditablePickerView(
                title: "Цель",
                value: viewModel.goal.title
            ) {
                if viewModel.isEditing {
                    isGoalPickerPresented = true
                }
            }
            .confirmationDialog(
                "Цель",
                isPresented: $isGoalPickerPresented,
                titleVisibility: .hidden
            ) {
                ForEach(MainGoal.allCases) { goal in
                    Button(goal.title) {
                        viewModel.goal = goal
                    }
                }
            }
        }
    }

    var body: some View {
        ZStack {
            AnimatedBackground()

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
                                }
                            } else {
                                viewModel.isEditing = true
                            }
                        }
                    }
                    
                    VStack(spacing: 24) {
                        AvatarView(photoURL: viewModel.avatarUrl)
                        
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
            }
        }
        .task { await viewModel.load() }
        .alert("Что-то пошло не так", isPresented: .constant(isError)) {
            Button("OK") {
                viewModel.screenState = .idle
            }
        } message: {
            if case let .error(message) = viewModel.screenState {
                Text(message)
            }
        }
        .onChange(of: viewModel.event) { event in
            switch event {
            case .logoutSuccess:
                router.selectedTab = .auth
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

#Preview {
    ProfileView()
}
