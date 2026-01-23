import SwiftUI

struct AvatarPickerView: View {
    let data: Data?

    var body: some View {
        Group {
            if let data, let image = UIImage(data: data) {
                Image(uiImage: image)
                .resizable()
                .scaledToFill()
            } else {
                Image(systemName: "person.crop.circle.fill")
                .resizable()
                .foregroundColor(.gray)
                .padding(20)
            }
        }
        .frame(width: 96, height: 96)
        .glassEffect()
        .clipShape(Circle())
    }
}
