import Foundation

struct User: Identifiable, Codable {
    let id: UUID
    let name: String?
    let surname: String?
    let phone: String?
    let login: String
    let email: String?
}

