import SwiftUI

struct RecipesView: View {
    let recipes: [Recipe]
    let isLoading: Bool
    let onRecipeTap: (Recipe) -> ()
    @Environment(\.dismiss) private var dismiss
    @State private var showFeedback = false

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
                                    dismiss()
                                }
                            }
                        }
                        .padding()
                    }
                }
            }
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Закрыть") { dismiss() }
                        .foregroundColor(.white)
                }
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button {
                        showFeedback = true
                    } label: {
                        Image(systemName: "bubble.and.pencil")
                            .foregroundColor(.white)
                    }
                }
            }
            .toolbarBackground(.hidden, for: .navigationBar)
            .sheet(isPresented: $showFeedback) {
                FeedbackSheet(title: "Отзыв о рецептах")
                    .presentationDetents([.medium, .large])
                    .presentationDragIndicator(.visible)
            }
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
