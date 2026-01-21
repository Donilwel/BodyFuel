struct LoginResponseBody: Decodable {
    let token: String

    enum CodingKeys: String, CodingKey {
        case token = "jwt"
    }
}
