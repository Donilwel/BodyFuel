import SwiftUI
import Kingfisher

struct AvatarView: View {
    let photoURL: String
    
    var body: some View {
        KFImage(URL(string: photoURL))
            .placeholder {
                Image(systemName: "person.crop.circle.fill")
                    .resizable()
                    .foregroundColor(.white.opacity(0.6))
            }
            .resizable()
            .scaledToFill()
            .frame(width: 120, height: 120)
            .clipShape(Circle())
            .glassEffect(in: .circle)
    }
}
