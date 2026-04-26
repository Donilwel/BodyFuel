import SwiftUI

struct WeightInputSheet: View {
    @ObservedObject var viewModel: StatsViewModel
    @Environment(\.dismiss) private var dismiss

    @State private var weightText = ""
    @State private var isLoading = false
    @State private var errorMessage: String?
    @FocusState private var weightFocused: Bool

    var body: some View {
        ZStack {
            Color.clear
                .glassEffect(.regular.tint(AppColors.primary.opacity(0.6)).interactive(), in: .rect)
                .ignoresSafeArea()

            VStack(spacing: 24) {
                Text("Добавить вес")
                    .font(.title2.bold())
                    .foregroundStyle(.white)

                VStack(alignment: .leading, spacing: 8) {
                    Text("Текущий вес (кг)")
                        .font(.subheadline)
                        .foregroundStyle(.white.opacity(0.8))

                    TextField("Например: 72.5", text: $weightText)
                        .keyboardType(.decimalPad)
                        .focused($weightFocused)
                        .padding()
                        .background(.ultraThinMaterial)
                        .clipShape(RoundedRectangle(cornerRadius: 16))
                        .foregroundStyle(.white)
                        .onChange(of: weightText) { newValue in
                            errorMessage = Validator.gramsAmount(newValue)
                        }
                        .toolbar {
                            ToolbarItemGroup(placement: .keyboard) {
                                Spacer()
                                Button("Готово") { weightFocused = false }
                                    .foregroundStyle(.white)
                            }
                        }
                }

                if let error = errorMessage {
                    Text(error)
                        .font(.caption)
                        .foregroundStyle(.red)
                }

                PrimaryButton(title: "Сохранить", isLoading: isLoading) {
                    save()
                }
                .disabled(errorMessage != nil)
                .frame(maxWidth: .infinity)

                SecondaryButton(title: "Отмена") {
                    dismiss()
                }
            }
            .padding(24)
        }
        .presentationDetents([.height(320)])
        .presentationCornerRadius(28)
        .onAppear {
            weightFocused = true
        }
        .onTapGesture {
            weightFocused = false
        }
    }

    private func save() {
        let normalized = weightText.replacingOccurrences(of: ",", with: ".")
        guard let value = Double(normalized), value > 0, value < 500 else {
            errorMessage = "Введите корректный вес"
            return
        }
        isLoading = true
        errorMessage = nil
        Task {
            do {
                try await viewModel.addWeight(value)
                dismiss()
            } catch {
                errorMessage = "Не удалось сохранить вес"
            }
            isLoading = false
        }
    }
}
