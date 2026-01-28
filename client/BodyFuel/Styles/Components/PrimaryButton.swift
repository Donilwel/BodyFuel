import SwiftUI

struct PrimaryButton: View {
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
        .foregroundColor(.white)
        .padding()
        .glassEffect(.regular.tint(AppColors.primary).interactive(), in: .rect(cornerRadius: 12))
        .disabled(isLoading)
    }
}

#Preview {
    AuthView()
}
