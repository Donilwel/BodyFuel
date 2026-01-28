import SwiftUI

struct CustomTextField<Field: Hashable>: View {
    let title: String
    var keyboardType: UIKeyboardType = .default
    let field: Field
    var focusedField: FocusState<Field?>.Binding
    
    @Binding var text: String

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(title)
                .font(.headline.bold())
                .foregroundColor(.white)
                .fixedSize(horizontal: false, vertical: true)
            
            TextField("", text: $text)
                .textInputAutocapitalization(.never)
                .autocorrectionDisabled()
                .keyboardType(keyboardType)
                .padding()
                .glassEffect(in: .rect(cornerRadius: 12.0))
                .frame(height: 50)
                .focused(focusedField, equals: field)
        }
        .padding(.vertical, 4)
    }
}
