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

struct ValidatedField<Field: View>: View {
    let error: String?
    let field: Field

    init(
        error: String?,
        @ViewBuilder field: () -> Field
    ) {
        self.error = error
        self.field = field()
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            field
                .overlay(
                    RoundedRectangle(cornerRadius: 12)
                        .stroke(error != nil ? Color.red : .clear, lineWidth: 1)
                )

            if let error {
                Text(error)
                    .font(.caption)
                    .foregroundColor(.red)
                    .transition(.opacity)
            }
        }
    }
}

struct AuthTextField: View {
    let title: String
    let keyboardType: UIKeyboardType
    @Binding var text: String

    var body: some View {
        TextField(title, text: $text)
            .textInputAutocapitalization(.never)
            .autocorrectionDisabled()
            .keyboardType(keyboardType)
            .padding()
            .glassEffect(in: .rect(cornerRadius: 12.0))
    }
}

struct PasswordField: View {
    let title: String
    @Binding var text: String

    var body: some View {
        SecureField(title, text: $text)
            .padding()
            .glassEffect(in: .rect(cornerRadius: 12.0))
    }
}

#Preview {
    AuthView()
}
