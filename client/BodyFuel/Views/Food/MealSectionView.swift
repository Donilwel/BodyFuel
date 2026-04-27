import SwiftUI

struct MealSectionView: View {
    let mealType: MealType
    let meals: [Meal]
    let isExpanded: Bool
    let onToggle: () -> Void
    let onAdd: () -> Void
    let onDelete: (Meal) -> Void

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
                        MealRowView(meal: meal, onDelete: { onDelete(meal) })
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
    let onDelete: () -> Void
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
            .contextMenu {
                Button(role: .destructive, action: onDelete) {
                    Label("Удалить", systemImage: "trash")
                }
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
            MacroChip(label: "Ж", value: macros.fat, color: .purple)
            MacroChip(label: "У", value: macros.carbs, color: .indigo)
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
