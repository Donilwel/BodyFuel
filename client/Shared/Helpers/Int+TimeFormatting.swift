import Foundation

extension Int {
    var formattedTime: String {
        let formatter = DateComponentsFormatter()
        
        formatter.calendar = Calendar.current
        formatter.calendar?.locale = Locale(identifier: "ru_RU")
        
        formatter.unitsStyle = .short
        
        formatter.allowedUnits = [.hour, .minute]
        
        formatter.zeroFormattingBehavior = .default
        
        return formatter.string(from: TimeInterval(self)) ?? "0 сек"
    }
}
