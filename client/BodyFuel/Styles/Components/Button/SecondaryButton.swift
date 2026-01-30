import SwiftUI

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
