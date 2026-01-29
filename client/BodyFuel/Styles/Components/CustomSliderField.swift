import SwiftUI

struct CustomSliderField: View {
    let title: String
    let from: Float
    let to: Float
    var step: Float = 1.0
    @Binding var value: Float
    
    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(title)
                .font(.headline.bold())
                .foregroundColor(.white)
                .fixedSize(horizontal: false, vertical: true)
            
            VStack(spacing: 6) {
                Slider(
                    value: $value,
                    in: from...to,
                    step: step
                ) {
                    Text(title)
                } minimumValueLabel: {
                    Text(Int(from).description)
                } maximumValueLabel: {
                    Text(Int(to).description)
                }
                .foregroundColor(.gray.opacity(0.6))
                
                Text(Int(value).description)
                    .foregroundColor(value == 0 ? .gray.opacity(0.6) : .black)
                    .frame(alignment: .center)
            }
            .padding()
            .glassEffect(in: .rect(cornerRadius: 12.0))
        }
    }
}
