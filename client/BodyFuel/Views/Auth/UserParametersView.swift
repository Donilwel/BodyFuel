import SwiftUI
import PhotosUI

struct UserParametersView: View {
    @StateObject private var viewModel = UserParametersViewModel()
    
    @State private var isLifestylePickerPresented = false
    @State private var isGoalPickerPresented = false
    
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
                    text: $viewModel.heightString.onChange {
                        viewModel.validateLive()
                    }
                )
            }
            
            ValidatedField(error: viewModel.weightError) {
                CustomTextField(title: "Вес", keyboardType: .numberPad, text: $viewModel.weightString)
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
                titleVisibility: .visible
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
                ForEach(Goal.allCases) { goal in
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
                title: "Количество шагов в день",
                from: 0,
                to: 30000,
                step: 1000,
                value: $viewModel.targetStepsDaily
            )
            
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
            
            PrimaryButton(
                title: "Отправить",
                isLoading: viewModel.screenState == .loading
            ) {
                Task { await viewModel.submit() }
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
        ZStack {
            AnimatedBackground()

            CustomCarousel {
                parametersFormContent
            } secondView: {
                goalsFormContent
            }
        }
        .alert("Что-то пошло не так", isPresented: .constant(isError)) {
            Button("OK") { viewModel.screenState = .idle }
        } message: {
            if case let .error(message) = viewModel.screenState {
                Text(message)
            }
        }
    }

    private var isError: Bool {
        if case .error = viewModel.screenState { return true }
        return false
    }
}

#Preview {
    UserParametersView()
}
