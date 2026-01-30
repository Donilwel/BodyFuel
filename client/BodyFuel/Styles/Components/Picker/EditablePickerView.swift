import SwiftUI

struct EditablePickerView: View {
    let title: String
    let value: String
    let onTap: () -> Void

    var body: some View {
        HStack {
            Text(title)
                .font(.headline.bold())
                .foregroundColor(.white)
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
