import Foundation

protocol AuthServiceProtocol {
    func login(user: LoginPayload) async throws
    func register(user: RegisterPayload) async throws
    func sendRecoveryCode(login: String) async throws
    func confirmRecovery(code: String, newPassword: String) async throws
    func sendUserParameters() async throws
}

final class AuthService: AuthServiceProtocol {
    static let shared = AuthService()
    
    private let networkClient = NetworkClient.shared
    private let tokenStorage = TokenStorage.shared
    
    private init() {}
    
    func login(user: LoginPayload) async throws {
        do {
            let urlComponents = URLComponents(string: API.baseURLString + API.Auth.login)
            guard let urlComponents, let url = urlComponents.url else {
                print("[ERROR] [AuthService/login] Invalid login URL")
                throw NetworkError.invalidURL
            }
            
            let response: LoginResponseBody = try await networkClient.request(
                requiresAuthorization: false,
                url: url,
                method: .post,
                requestBody: user
            )
            
            tokenStorage.token = response.token
            
            print("[INFO] [AuthService/login]: Successfully logged in, token: \(response.token)")
        } catch {
            print("[ERROR] [AuthService/login]: \(error.localizedDescription)")
            throw mapToAuthError(error)
        }
    }

    func register(user: RegisterPayload) async throws {
        do {
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
        } catch {
            print("[ERROR] [AuthService/register]: \(error.localizedDescription)")
            throw mapToAuthError(error)
        }
    }

    func sendRecoveryCode(login: String) async throws { }
    func confirmRecovery(code: String, newPassword: String) async throws { }
    func sendUserParameters() async throws {}
}

extension AuthService {
    private func mapToAuthError(_ error: Error) -> AuthError {
        guard let networkError = error as? NetworkError else {
            return .invalidData(error.localizedDescription)
        }
        
        switch networkError {
        case .requestFailed(let statusCode, let message):
            switch statusCode {
            case 400:
                return .validation
            case 401:
                return .invalidCredentials
            case 409:
                return .userExists
            default:
                return .invalidData(message)
            }
            
        default:
            return .invalidData(error.localizedDescription)
        }
    }
}
