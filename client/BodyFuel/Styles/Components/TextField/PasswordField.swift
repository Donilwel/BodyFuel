import SwiftUI

struct PasswordField<Field: Hashable>: View {
    let title: String
    let field: Field
    var focusedField: FocusState<Field?>.Binding
    @Binding var text: String

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(title)
                .font(.headline.bold())
                .foregroundColor(.white)
                .fixedSize(horizontal: false, vertical: true)
            
            SecureField("", text: $text)
                .padding()
                .glassEffect(in: .rect(cornerRadius: 12.0))
                .focused(focusedField, equals: field)
        }
    }
}
