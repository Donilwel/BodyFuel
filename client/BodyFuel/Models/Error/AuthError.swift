import Foundation

enum AuthError: LocalizedError {
    case invalidCredentials
    case validation
    case userExists
    case invalidData(String)

    var errorDescription: String? {
        switch self {
        case .invalidCredentials: return "Неверный логин или пароль"
        case .validation: return "Проверьте корректность данных"
        case .userExists: return "Пользователь с таким логином уже существует"
        case .invalidData(let message): return message
        }
    }
}
