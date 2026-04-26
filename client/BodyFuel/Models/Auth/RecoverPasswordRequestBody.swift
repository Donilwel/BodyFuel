import Foundation

struct RecoverPasswordRequestBody: Encodable {
    let email: String
}

struct ResetPasswordRequestBody: Encodable {
    let email: String
    let code: String
    let newPassword: String

    enum CodingKeys: String, CodingKey {
        case email
        case code
        case newPassword = "new_password"
    }
}
