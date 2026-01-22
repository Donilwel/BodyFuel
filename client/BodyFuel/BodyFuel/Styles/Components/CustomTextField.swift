import SwiftUI

struct CustomTextField: View {
    let title: String
    var keyboardType: UIKeyboardType = .default
    @Binding var text: String

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(title)
                .font(.headline.bold())
                .foregroundColor(.white)
            
            TextField("", text: $text)
                .textInputAutocapitalization(.never)
                .autocorrectionDisabled()
                .keyboardType(keyboardType)
                .padding()
                .glassEffect(in: .rect(cornerRadius: 12.0))
                .frame(height: 50)
        }
        .padding(.vertical, 4)
    }
}
