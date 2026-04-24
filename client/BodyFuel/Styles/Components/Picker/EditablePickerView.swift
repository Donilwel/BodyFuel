import SwiftUI

struct EditablePickerView: View {
    let title: String
    let value: String
    let onTap: () -> Void

    var body: some View {
        HStack {
            Text(title)
                .foregroundColor(.white.opacity(0.8))
                .fixedSize(horizontal: false, vertical: true)
            
            Spacer()
            
            Button(action: onTap) {
                Text(value)
                    .lineLimit(1)
                    .foregroundColor(.white)
            }
        }
    }
}
