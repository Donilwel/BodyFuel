import Foundation

enum AppError: LocalizedError, Equatable {
    case noInternet
    case serverUnavailable
    case unauthorized
    case validation(message: String)
    case notFound
    case decoding
    case encoding
    case unknown
    
    var errorDescription: String? {
        switch self {
        case .noInternet:
            return "Нет подключения к интернету"
        case .serverUnavailable:
            return "Сервер временно недоступен"
        case .unauthorized:
            return "Требуется повторный вход"
        case .validation(let message):
            return message
        case .notFound:
            return "Данные не найдены"
        case .decoding, .encoding:
            return "Ошибка обработки данных"
        case .unknown:
            return "Попробуйте еще раз"
        }
    }
}
