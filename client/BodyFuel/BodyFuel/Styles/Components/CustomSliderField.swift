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
            
            VStack(spacing: 6) {
                Slider(
                    value: $value,
                    in: from...to,
                    step: step
                ) {
                    Text(title)
                } minimumValueLabel: {
                    Text(from.formatted())
                } maximumValueLabel: {
                    Text(to.formatted())
                }
                .foregroundColor(.gray.opacity(0.6))
                
                Text("\(value.formatted())")
                    .foregroundColor(value == 0 ? .gray.opacity(0.6) : .black)
                    .frame(alignment: .center)
            }
            .padding()
            .glassEffect(in: .rect(cornerRadius: 12.0))
        }
    }
}
