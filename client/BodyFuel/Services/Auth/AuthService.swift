import Foundation

protocol AuthServiceProtocol {
    func login(user: LoginPayload) async throws
    func register(user: RegisterPayload) async throws
    func sendRecoveryCode(login: String) async throws
    func confirmRecovery(code: String, newPassword: String) async throws
    func sendUserParameters() async throws
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

final class AuthService: AuthServiceProtocol {
    static let shared = AuthService()
    
    private let networkClient = NetworkClient.shared
    private let tokenStorage = TokenStorage.shared
    
    private init() {}
    
    func login(user: LoginPayload) async throws {
        let urlComponents = URLComponents(string: API.baseURLString + API.Auth.login)!
        guard let url = urlComponents.url else {
            print("[ERROR] [AuthService/login] Invalid login URL")
            throw NetworkError.invalidURL
        }
        
        do {
            let response: LoginResponseBody = try await networkClient.request(
                requiresAuthorization: false,
                url: url,
                method: .post,
                requestBody: user
            )
            
            tokenStorage.token = response.token
            
            print("[INFO] [AuthService/login]: Successfully logged in")
        } catch {
            if error.localizedDescription.contains("401") {
                throw AuthError.invalidCredentials
            } else {
                throw AuthError.invalidData(error.localizedDescription)
            }
        }
    }

    func register(user: RegisterPayload) async throws {
        let urlComponents = URLComponents(string: API.baseURLString + API.Auth.register)!
        guard let url = urlComponents.url else {
            print("[ERROR] [AuthService/register] Invalid register URL")
            throw NetworkError.invalidURL
        }
        
        let response: APIMessageResponse = try await networkClient.request(
            requiresAuthorization: false,
            url: url,
            method: .post,
            requestBody: user
        )
        
        print("[INFO] [AuthService/register]: \(response.message)")
    }

    func sendRecoveryCode(login: String) async throws { }
    func confirmRecovery(code: String, newPassword: String) async throws { }
    func sendUserParameters() async throws {}
}
