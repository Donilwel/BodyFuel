import Foundation

enum Validator {
    static func emailError(_ value: String) -> String? {
        guard !value.isEmpty else { return "Введите почту" }
        let isValid = value.range(of: "^[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,}$", options: [.regularExpression, .caseInsensitive]) != nil
        return isValid ? nil : "Некорректный email"
    }

    static func phoneError(_ value: String) -> String? {
        let digits = value.filter { $0.isNumber }
        guard !digits.isEmpty else { return "Введите телефон" }
        guard digits.count == 11 && digits.hasPrefix("7") else { return "Некорректный телефон" }
        return nil
    }

    static func passwordError(_ value: String) -> String? {
        guard !value.isEmpty else { return "Введите пароль" }
        return value.count >= 6 ? nil : "Минимум 6 символов"
    }

    static func macroGrams(_ text: String) -> String? {
        guard !text.isEmpty else { return nil }
        let normalized = text.replacingOccurrences(of: ",", with: ".")
        guard let value = Double(normalized) else { return "Введите число" }
        guard value >= 0 else { return "Значение не может быть отрицательным" }
        guard value <= 1000 else { return "Слишком большое значение" }
        return nil
    }

    static func gramsAmount(_ text: String) -> String? {
        guard !text.isEmpty else { return nil }
        let normalized = text.replacingOccurrences(of: ",", with: ".")
        guard let value = Double(normalized) else { return "Введите число" }
        guard value > 0 else { return "Введите количество больше нуля" }
        return nil
    }
}
