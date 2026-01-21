import Foundation

struct User: Identifiable, Codable {
    let id: UUID
    let fullName: String?
    let phone: String?
    let login: String
    let email: String?
}

