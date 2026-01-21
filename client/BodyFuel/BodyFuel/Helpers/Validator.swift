import Foundation

enum Validator {
    static func emailError(_ value: String) -> String? {
        guard !value.isEmpty else { return "Введите почту" }
        let isValid = value.range(of: "^[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,}$", options: [.regularExpression, .caseInsensitive]) != nil
        return isValid ? nil : "Некорректный email"
    }

    static func phoneError(_ value: String) -> String? {
        guard !value.isEmpty else { return "Введите телефон" }
        let isValid = value.range(of: "^[0-9+]{10,15}$", options: .regularExpression) != nil
        return isValid ? nil : "Некорректный телефон"
    }

    static func passwordError(_ value: String) -> String? {
        guard !value.isEmpty else { return "Введите пароль" }
        return value.count >= 6 ? nil : "Минимум 6 символов"
    }
}
