import Foundation

struct ErrorMapper {
    static func map(_ error: Error) -> AppError {
        if let error = error as? NetworkError {
            switch error {
            case .network:
                return .noInternet
                
            case .requestFailed(let status, _):
                switch status {
                case 401: return .unauthorized
                case 404: return .notFound
                case 500...599: return .serverUnavailable
                default: return .unknown
                }
                
            case .decodingFailed:
                return .decoding
                
            case .encodingFailed:
                return .encoding
                
            case .missingToken:
                return .unauthorized
                
            case .invalidURL:
                return .serverUnavailable
                
            default:
                return .unknown
            }
        }
        
        if let auth = error as? AuthError {
            switch auth {
            case .invalidCredentials:
                return .validation(message: "Неверный логин или пароль")
            case .validation:
                return .validation(message: "Проверьте введённые данные")
            case .userExists:
                return .validation(message: "Пользователь уже существует")
            case .invalidData(let message):
                return .validation(message: message)
            }
        }
        
//        if let profile = error as? ProfileError {
//            switch profile {
//            case .uploadFailed:
//                return .serverUnavailable
//            case .invalidData:
//                return .validation(message: "Некорректные данные профиля")
//            case .unknown:
//                return .unknown
//            }
//        }
        
        return .unknown
    }
}
