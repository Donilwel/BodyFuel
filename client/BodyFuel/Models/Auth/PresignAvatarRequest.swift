import Foundation

struct PresignAvatarRequest: Encodable {
    let contentType: String
    
    private enum CodingKeys: String, CodingKey {
        case contentType = "content_type"
    }
}
