import SwiftUI

struct CustomPickerField: View {
    let title: String
    let value: String
    let onTap: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(title)
                .font(.headline.bold())
                .foregroundColor(.white)
                .fixedSize(horizontal: false, vertical: true)
            Button(action: onTap) {
                HStack {
                    Text(value)
                        .foregroundColor(.black)
                    
                    Spacer()
                    
                    Image(systemName: "chevron.down")
                        .foregroundColor(.gray.opacity(0.6))
                }
                .padding()
                .glassEffect(in: .rect(cornerRadius: 12))
            }
        }
    }
}
