import SwiftUI

struct PhoneTextField<Field: Hashable>: View {
    let field: Field
    var focusedField: FocusState<Field?>.Binding
    @Binding var text: String

    @State private var displayText: String = ""

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Телефон")
                .font(.headline.bold())
                .foregroundColor(.white)
                .fixedSize(horizontal: false, vertical: true)

            TextField("+7 (___) ___-__-__", text: $displayText)
                .typesettingLanguage(Locale.Language(identifier: "en"))
                .textInputAutocapitalization(.never)
                .autocorrectionDisabled()
                .keyboardType(.phonePad)
                .padding()
                .glassEffect(in: .rect(cornerRadius: 12.0))
                .frame(height: 50)
                .focused(focusedField, equals: field)
                .onChange(of: displayText) { newValue in
                    let formatted = Self.format(newValue)
                    if formatted != newValue {
                        displayText = formatted
                    }
                    text = formatted
                }
                .onAppear {
                    if !text.isEmpty {
                        displayText = text
                    }
                }
        }
        .padding(.vertical, 4)
    }

    static func format(_ input: String) -> String {
        var digits = input.filter { $0.isNumber }

        if digits.hasPrefix("8") {
            digits = "7" + digits.dropFirst()
        }

        digits = String(digits.prefix(11))

        guard !digits.isEmpty else { return "" }

        guard digits.hasPrefix("7") else {
            return String(digits.prefix(1))
        }

        var result = "+7"
        let rest = String(digits.dropFirst())

        if rest.isEmpty { return result }

        result += " ("
        let part1 = String(rest.prefix(3))
        result += part1
        if rest.count <= 3 { return result }

        result += ") "
        let part2 = String(rest.dropFirst(3).prefix(3))
        result += part2
        if rest.count <= 6 { return result }

        result += "-"
        let part3 = String(rest.dropFirst(6).prefix(2))
        result += part3
        if rest.count <= 8 { return result }

        result += "-"
        let part4 = String(rest.dropFirst(8).prefix(2))
        result += part4

        return result
    }
}
