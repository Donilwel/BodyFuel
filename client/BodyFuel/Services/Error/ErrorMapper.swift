import Foundation

struct ErrorMapper {
    static func map(_ error: Error) -> AppError {
        if let error = error as? NetworkError {
            switch error {
            case .requestFailed(let status, let message):
                switch status {
                case 401: return .unauthorized
                case 404: return .notFound
                case 400...499:
                    return .validation(message: message)
                case 500...599:
                    return .serverUnavailable
                default:
                    return .unknown
                }

            case .decodingFailed:
                return .decoding

            case .encodingFailed:
                return .encoding

            case .missingToken:
                return .unauthorized

            case .invalidURL:
                return .serverUnavailable

            case .network(let underlying):
                if let urlError = underlying as? URLError,
                   urlError.code == .notConnectedToInternet || urlError.code == .networkConnectionLost {
                    return .noInternet
                }
                return .unknown
            }
        }

        if let auth = error as? AuthError {
            switch auth {
            case .invalidCredentials:
                return .validation(message: error.localizedDescription)
            case .validation:
                return .validation(message: error.localizedDescription)
            case .userExists:
                return .validation(message: error.localizedDescription)
            case .invalidData(let message):
                return .validation(message: message)
            }
        }

        if let profile = error as? ProfileError {
            switch profile {
            case .validation:
                return .validation(message: error.localizedDescription)
            case .unauthorized:
                return .unauthorized
            case .invalidData(let message):
                return .validation(message: message)
            }
        }

        if let health = error as? HealthError {
            switch health {
            case .noPermission:
                return .validation(message: "Разрешите доступ к данным Здоровья в настройках приложения")
            case .emptyValue(let message):
                return .validation(message: message)
            }
        }

        return .unknown
    }
}
