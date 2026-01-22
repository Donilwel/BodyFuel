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
        .frame(height: 48)
        .background(AppColors.primary)
        .foregroundColor(.white)
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .disabled(isLoading)
    }
}
