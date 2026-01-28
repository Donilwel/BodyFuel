import SwiftUI
import PhotosUI

struct UserParametersView: View {
    @EnvironmentObject var router: AppRouter
    @StateObject private var viewModel = UserParametersViewModel()
    
    @State private var isLifestylePickerPresented = false
    @State private var isGoalPickerPresented = false
    
    @FocusState private var parametersFocused: ParametersField?
    
    private enum ParametersField: Hashable {
        case height
        case weight
    }
    
    private var parametersFields: some View {
        VStack(spacing: 16) {
            PhotosPicker(selection: $viewModel.avatarItem, matching: .images) {
                AvatarPickerView(data: viewModel.avatarData)
            }
            .onChange(of: viewModel.avatarItem) { _ in
                Task { await viewModel.loadAvatar() }
            }
            
            ValidatedField(error: viewModel.heightError) {
                CustomTextField(
                    title: "Рост",
                    keyboardType: .numberPad,
                    field: ParametersField.height,
                    focusedField: $parametersFocused,
                    text: $viewModel.heightString.onChange {
                        viewModel.validateLive()
                    }
                )
            }
            
            ValidatedField(error: viewModel.weightError) {
                CustomTextField(
                    title: "Вес",
                    keyboardType: .numberPad,
                    field: ParametersField.weight,
                    focusedField: $parametersFocused,
                    text: $viewModel.weightString
                )
            }
            
            CustomPickerField(
                title: "Образ жизни",
                value: viewModel.lifestyle?.title ?? ""
            ) {
                isLifestylePickerPresented = true
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
        }
    }
    
    private var goalsFields: some View {
        VStack(spacing: 16) {
            CustomPickerField(
                title: "Цель",
                value: viewModel.goal?.title ?? ""
            ) {
                isGoalPickerPresented = true
            }
            .confirmationDialog(
                "Цель",
                isPresented: $isGoalPickerPresented,
                titleVisibility: .visible
            ) {
                ForEach(MainGoal.allCases) { goal in
                    Button(goal.title) {
                        viewModel.goal = goal
                    }
                }
            }
            
            if viewModel.goal != nil && viewModel.weight >= 40 && viewModel.goal != .maintain {
                let from = max(40, Float(viewModel.goal == .loseWeight
                                         ? viewModel.weight - 50
                                         : viewModel.weight))
                let to = Float(viewModel.goal == .loseWeight
                               ? viewModel.weight
                               : viewModel.weight + 50)
                
                ValidatedField(error: viewModel.targetWeightError) {
                    CustomSliderField(
                        title: "Желаемый вес",
                        from: from,
                        to: to,
                        value: $viewModel.targetWeight
                    )
                }
            }
            
            CustomSliderField(
                title: "Количество тренировок в неделю",
                from: 0,
                to: 7,
                step: 1,
                value: $viewModel.targetWorkoutsWeekly
            )
        }
    }
    
    private var parametersFormContent: some View {
        VStack(spacing: 24) {
            Text("Ответь еще на несколько вопросов для составления идеального плана")
                .font(.title2.bold())
                .foregroundColor(.white)
            
            parametersFields
        }
        .padding(24)
        .background(
            RoundedRectangle(cornerRadius: 28)
                .fill(.ultraThinMaterial)
        )
        .padding(.horizontal, 20)
        .padding(.top, 40)
    }
    
    private var goalsFormContent: some View {
        VStack(spacing: 24) {
            VStack(alignment: .leading, spacing: 24) {
                Text("Расскажи нам о своих целях")
                    .font(.title2.bold())
                    .foregroundColor(.white)
                
                goalsFields
            }
        }
        .padding(24)
        .background(
            RoundedRectangle(cornerRadius: 28)
                .fill(.ultraThinMaterial)
        )
        .padding(.horizontal, 20)
        .padding(.top, 40)
    }
    
    private var caloriesPreviewContent: some View {
        VStack(alignment: .center, spacing: 24) {
            VStack(alignment: .leading, spacing: 24) {
                Text("Давай рассчитаем твою норму калорий для достижения цели")
                    .font(.title2.bold())
                    .foregroundColor(.white)
                    .fixedSize(horizontal: false, vertical: true)
                
                Text("Убедись, что ты ответил на все вопросы, они нужны нам для расчетов")
                    .font(.headline.bold())
                    .foregroundColor(.white)
                    .fixedSize(horizontal: false, vertical: true)
                    .multilineTextAlignment(.leading)
            }
            
            PrimaryButton(
                title: "Рассчитать",
                isLoading: false
            ) {
                Task {
                    await viewModel.countRecommendedCalories()
                    viewModel.caloriesFormState = .counting
                }
            }
        }
    }
    
    private var caloriesCountContent: some View {
        VStack(alignment: .center, spacing: 8) {
            VStack(alignment: .leading, spacing: 24) {
                Text("Чтобы безопасно двигаться к цели, мы рекомендуем придерживаться")
                    .foregroundColor((.white))
                    .font(.title3.bold())
                    .fixedSize(horizontal: false, vertical: true)
                
                Text("\(Int(viewModel.targetCaloriesDaily)) ккал/день.")
                    .foregroundColor((.white))
                    .font(.title2.bold())
                    .fixedSize(horizontal: false, vertical: true)
                
                Text("Рассчитано на основе роста, веса, возраста, активности и цели")
                    .foregroundColor(.white)
                    .fixedSize(horizontal: false, vertical: true)
                    .multilineTextAlignment(.leading)
            }
            
            Spacer()
                .frame(height: 8)
            
            PrimaryButton(
                title: "Отправить",
                isLoading: viewModel.screenState == .loading
            ) {
                submit()
            }
            
            Button("Изменить норму") {
                viewModel.caloriesFormState = .editing
            }
            .padding(.horizontal)
            .frame(height: 20)
            .foregroundColor(.white.opacity(0.75))
            .fontWeight(.semibold)
            .padding()
        }
    }
    
    private var caloriesEditingContent: some View {
        VStack(spacing: 24) {
            VStack(alignment: .leading, spacing: 24) {
                CustomSliderField(
                    title: "Количество калорий в день",
                    from: viewModel.dailyEnergyExpenditure * 0.5,
                    to: viewModel.dailyEnergyExpenditure * 1.5,
                    step: 100,
                    value: $viewModel.targetCaloriesDaily
                )
                
                Text(viewModel.validateCaloriesNorm())
                    .font(.title3.bold())
                    .foregroundColor(.white)
                    .fixedSize(horizontal: false, vertical: true)
                
                Text("Твое тело тратит в среднем \(Int(viewModel.dailyEnergyExpenditure)) калорий в день. Слишком сильное отклонение от этого значения может повлиять на твое здоровье.")
                    .font(.headline.bold())
                    .foregroundColor(.white)
            }
            
            PrimaryButton(
                title: "Отправить",
                isLoading: viewModel.screenState == .loading
            ) {
                submit()
            }
        }
    }
    
    private var caloriesFormContent: some View {
        VStack(spacing: 24) {
            switch viewModel.caloriesFormState {
            case .preview:
                caloriesPreviewContent
            case .counting:
                caloriesCountContent
            case .editing:
                caloriesEditingContent
            }
        }
        .padding(24)
        .background(
            RoundedRectangle(cornerRadius: 28)
                .fill(.ultraThinMaterial)
        )
        .padding(.horizontal, 20)
        .padding(.top, 40)
    }
    
    var body: some View {
        NavigationStack {
            ZStack {
                AnimatedBackground()
                
                CustomCarousel(totalPages: 3) {
                    parametersFormContent
                    goalsFormContent
                    caloriesFormContent
                }
            }
            .alert("Что-то пошло не так", isPresented: .constant(isError)) {
                Button("OK") { viewModel.screenState = .idle }
            } message: {
                if case let .error(message) = viewModel.screenState {
                    Text(message)
                }
            }
            .onTapGesture {
                parametersFocused = nil
            }
        }
    }
    
    private func submit() {
        Task { await viewModel.submit() }
        router.currentFlow = .main
    }

    private var isError: Bool {
        if case .error = viewModel.screenState { return true }
        return false
    }
}

#Preview {
    UserParametersView()
}
struct SecondaryButton: View {
    let title: String
    let isLoading: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            if isLoading {
                ProgressView()
            } else {
                Text(title)
                    .fontWeight(.semibold)
            }
        }
        .padding(.horizontal)
        .frame(height: 20)
        .foregroundColor(.white.opacity(0.75))
        .padding()
        .disabled(isLoading)
    }
}
