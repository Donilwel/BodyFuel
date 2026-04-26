import SwiftUI

struct RecipesView: View {
    let recipes: [Recipe]
    let isLoading: Bool
    let onRecipeTap: (Recipe) -> ()
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            ZStack {
                Color.clear
                    .glassEffect(.regular.tint(AppColors.primary.opacity(0.6)).interactive(), in: .rect)
                    .ignoresSafeArea()

                if isLoading {
                    ProgressView()
                        .tint(.white)
                } else {
                    ScrollView {
                        VStack(spacing: 16) {
                            ForEach(recipes) { recipe in
                                RecipeCard(recipe: recipe) { recipe in
                                    onRecipeTap(recipe)
                                }
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
    let onTap: (Recipe) -> ()

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
        .onTapGesture {
            onTap(recipe)
        }
        .cardStyle()
    }
}
