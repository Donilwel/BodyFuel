import SwiftUI

struct WorkoutFilterView: View {
    @ObservedObject var viewModel: WorkoutViewModel
    @Environment(\.dismiss) private var dismiss

    @State private var selectedPlace: WorkoutPlace? = nil
    @State private var selectedType: ExerciseType? = nil
    @State private var selectedLevel: WorkoutLevel? = nil

    private let chipColumns = [GridItem(.adaptive(minimum: 100), spacing: 8)]

    var body: some View {
        ZStack {
            Color.clear
                .glassEffect(.regular.tint(AppColors.primary.opacity(0.6)).interactive(), in: .rect)
                .ignoresSafeArea()
            
            VStack(spacing: 0) {
                ScrollView {
                    VStack(alignment: .leading, spacing: 28) {
                        Text("Параметры тренировки")
                            .font(.title2.bold())
                            .foregroundColor(.white)
                        
                        filterSection(title: "Место") {
                            chipsGrid(
                                options: WorkoutPlace.allCases,
                                selected: $selectedPlace,
                                label: { $0.rawValue }
                            )
                        }
                        
                        filterSection(title: "Тип") {
                            chipsGrid(
                                options: ExerciseType.allCases,
                                selected: $selectedType,
                                label: { $0.rawValue }
                            )
                        }
                        
                        filterSection(title: "Уровень") {
                            chipsGrid(
                                options: WorkoutLevel.allCases,
                                selected: $selectedLevel,
                                label: { $0.rawValue }
                            )
                        }
                    }
                    .padding(24)
                }
                
                VStack(spacing: 12) {
                    PrimaryButton(title: "Сгенерировать") {
                        Task {
                            await viewModel.generateWithFilters(
                                place: selectedPlace,
                                type: selectedType,
                                level: selectedLevel
                            )
                        }
                    }
                    SecondaryButton(title: "Отмена") {
                        dismiss()
                    }
                }
                .padding(.horizontal, 24)
                .padding(.bottom, 32)
                .padding(.top, 12)
            }
            .presentationDetents([.medium, .large])
            .presentationDragIndicator(.visible)
        }
    }

    private func filterSection<Content: View>(title: String, @ViewBuilder content: () -> Content) -> some View {
        VStack(alignment: .leading, spacing: 12) {
            Text(title)
                .font(.headline.bold())
                .foregroundColor(.white)
            content()
        }
    }

    private func chipsGrid<T: Hashable>(
        options: [T],
        selected: Binding<T?>,
        label: @escaping (T) -> String
    ) -> some View {
        LazyVGrid(columns: chipColumns, alignment: .leading, spacing: 8) {
            ForEach(options, id: \.self) { option in
                let isSelected = selected.wrappedValue == option
                Button(label(option)) {
                    selected.wrappedValue = isSelected ? nil : option
                }
                .font(.subheadline.weight(.medium))
                .foregroundColor(isSelected ? .black : .white)
                .frame(maxWidth: .infinity)
                .padding(.vertical, 10)
                .background(
                    RoundedRectangle(cornerRadius: 20)
                        .fill(isSelected ? Color.white : Color.white.opacity(0.15))
                )
            }
        }
    }
}
