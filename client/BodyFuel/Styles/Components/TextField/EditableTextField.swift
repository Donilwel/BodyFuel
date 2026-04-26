import SwiftUI

struct EditableTextField<Field: Hashable, Value: LosslessStringConvertible>: View {
    let title: String
    @Binding var value: Value
    var suffix: String?
    var isEditing: Bool
    let field: Field
    var focusedField: FocusState<Field?>.Binding
    
    @State private var textValue: String = ""
    
    var body: some View {
        HStack {
            Text(title)
                .foregroundColor(.white.opacity(0.8))
            
            Spacer()
            
            if isEditing {
                TextField("", text: $textValue)
                    .multilineTextAlignment(.trailing)
                    .focused(focusedField, equals: field)
                    .keyboardType(getKeyboardType())
                    .textInputAutocapitalization(.never)
                    .autocorrectionDisabled()
                    .foregroundColor(.white)
                    .onAppear {
                        textValue = String(value)
                    }
                    .onChange(of: textValue) { newValue in
                        if let converted = Value(newValue) {
                            value = converted
                        }
                    }
            } else {
                Text(suffix == nil ? String(value) : "\(value) \(suffix!)")
                    .foregroundColor(.white)
            }
        }
    }
    
    private func getKeyboardType() -> UIKeyboardType {
        if Value.self == Int.self || Value.self == Double.self || Value.self == Float.self {
            return .decimalPad
        } else if Value.self == String.self {
            return .default
        }
        return .default
    }
}

#Preview {
    ProfileView()
}
