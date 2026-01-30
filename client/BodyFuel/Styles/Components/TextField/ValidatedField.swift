import SwiftUI

struct ValidatedField<Field: View>: View {
    let error: String?
    let field: Field

    init(
        error: String?,
        @ViewBuilder field: () -> Field
    ) {
        self.error = error
        self.field = field()
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            field
//                .overlay(
//                    RoundedRectangle(cornerRadius: 12)
//                        .stroke(error != nil ? Color.red : .clear, lineWidth: 1)
//                )

            if let error {
                Text(error)
                    .font(.caption)
                    .foregroundColor(.red)
                    .transition(.opacity)
            }
        }
    }
}
