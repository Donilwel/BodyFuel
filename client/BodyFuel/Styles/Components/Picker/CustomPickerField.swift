import SwiftUI

struct CustomPickerField<Option: Identifiable & Equatable>: View {
    let title: String
    let options: [Option]
    let optionTitle: (Option) -> String
    @Binding var selection: Option?

    @State private var isExpanded = false

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(title)
                .font(.headline.bold())
                .foregroundColor(.white)
                .fixedSize(horizontal: false, vertical: true)

            VStack(spacing: 0) {
                Button {
                    withAnimation(.spring(duration: 0.25)) {
                        isExpanded.toggle()
                    }
                } label: {
                    HStack {
                        Text(selection.map(optionTitle) ?? "Выберите...")
                            .foregroundColor(selection != nil ? .black : .gray.opacity(0.6))
                        Spacer()
                        Image(systemName: "chevron.down")
                            .foregroundColor(.gray.opacity(0.6))
                            .rotationEffect(.degrees(isExpanded ? 180 : 0))
                            .animation(.spring(duration: 0.25), value: isExpanded)
                    }
                    .padding()
                }

                if isExpanded {
                    Divider()
                        .padding(.horizontal, 12)

                    ForEach(Array(options.enumerated()), id: \.offset) { index, option in
                        Button {
                            withAnimation(.spring(duration: 0.25)) {
                                selection = option
                                isExpanded = false
                            }
                        } label: {
                            HStack {
                                Text(optionTitle(option))
                                    .foregroundColor(.black)
                                Spacer()
                                if selection == option {
                                    Image(systemName: "checkmark")
                                        .foregroundColor(AppColors.primary)
                                        .font(.footnote.bold())
                                }
                            }
                            .padding(.horizontal)
                            .padding(.vertical, 12)
                        }

                        if index < options.count - 1 {
                            Divider()
                                .padding(.horizontal, 12)
                        }
                    }
                }
            }
            .glassEffect(in: .rect(cornerRadius: 12))
        }
    }
}
