import SwiftUI

struct PasswordField: View {
    let title: String
    @Binding var text: String

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(title)
                .font(.headline.bold())
                .foregroundColor(.white)
            
            SecureField("", text: $text)
                .padding()
                .glassEffect(in: .rect(cornerRadius: 12.0))
        }
    }
}
