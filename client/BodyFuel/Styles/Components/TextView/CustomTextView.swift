import SwiftUI

struct CustomTextView<Value: LosslessStringConvertible>: View {
    let title: String
    @Binding var value: Value
    var suffix: String?
    
    var body: some View {
        HStack {
            Text(title)
                .foregroundColor(.white.opacity(0.8))
            
            Spacer()
            
            Text(String(value) + (suffix ?? ""))
                .foregroundColor(.white)
        }
    }
    
    private func getKeyboardType() -> UIKeyboardType {
        if Value.self == Int.self || Value.self == Double.self || Value.self == Float.self {
            return .decimalPad
        } else if Value.self == String.self {
            return .default
        }
        return .default
    }
}
