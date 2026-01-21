import Foundation

protocol AuthServiceProtocol {
    func login(login: String, password: String) async throws -> User
    func register(user: RegisterPayload) async throws -> User
    func sendRecoveryCode(login: String) async throws
    func confirmRecovery(code: String, newPassword: String) async throws
}

enum AuthError: LocalizedError {
    case invalidCredentials
    case invalidData(String)

    var errorDescription: String? {
        switch self {
        case .invalidCredentials: return "Неверный логин или пароль"
        case .invalidData(let message): return message
        }
    }
}

struct RegisterPayload: Encodable {
    let name: String
    let surname: String
    let phone: String
    let login: String
    let email: String
    let password: String
}

final class AuthService: AuthServiceProtocol {
    static let shared = AuthService()
    
    private init() {}
    
    func login(login: String, password: String) async throws -> User {
        guard password == "123456" else { throw AuthError.invalidCredentials }
        return User(id: .init(), name: "test", surname: "test", phone: nil, login: login, email: nil)
    }

    func register(user: RegisterPayload) async throws -> User {
        return User(id: .init(), name: user.name, surname: user.surname, phone: user.phone, login: user.login, email: user.email)
    }

    func sendRecoveryCode(login: String) async throws { }
    func confirmRecovery(code: String, newPassword: String) async throws { }
}
