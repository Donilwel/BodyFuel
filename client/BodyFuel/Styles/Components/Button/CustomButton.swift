import SwiftUI

struct CustomButton<T: View>: View {
    let headline: String
    let title: T
    let onTap: () -> Void
    
    init(headline: String, onTap: @escaping () -> Void, @ViewBuilder title: () -> T) {
        self.headline = headline
        self.title = title()
        self.onTap = onTap
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(headline)
                .font(.headline.bold())
                .foregroundColor(.white)
            Button(action: onTap) {
                title
                .padding()
                .glassEffect(in: .rect(cornerRadius: 12))
            }
        }
    }
}
