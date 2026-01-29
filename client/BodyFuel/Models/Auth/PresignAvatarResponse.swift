import Foundation

struct PresignAvatarResponse: Decodable {
    let uploadURL: URL
    let objectKey: String
    let avatarURL: String
    
    private enum CodingKeys: String, CodingKey {
        case uploadURL = "upload_url"
        case objectKey = "object_key"
        case avatarURL = "avatar_url"
    }
}
