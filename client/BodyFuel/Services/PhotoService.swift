import Foundation
import SwiftUI

protocol PhotoServiceProtocol {
    func uploadUserAvatar(data: Data) async throws -> String
}

final class PhotoService: PhotoServiceProtocol {
    static let shared = PhotoService()
    
    private let networkClient = NetworkClient.shared
    
    private init() {}
    
    func uploadUserAvatar(data: Data) async throws -> String {
        let presign = try await getPresignedURL()

        try await networkClient.uploadPhoto(data: data, to: presign.uploadURL)
        
        print("[INFO] [PhotoService/uploadUserAvatar]: Successfully uploaded avatar: \(presign.avatarURL)")
        return presign.avatarURL
    }
    
    private func getPresignedURL() async throws -> PresignAvatarResponse {
        let urlComponents = URLComponents(string: API.baseURLString + API.uploadAvatar)
        guard let urlComponents, let url = urlComponents.url else {
            print("[ERROR] [PhotoService/getPresignedURL] Invalid avatar upload URL")
            throw NetworkError.invalidURL
        }
        
        do {
            let response: PresignAvatarResponse = try await networkClient.request(
                url: url,
                method: .post,
                requestBody: PresignAvatarRequest(contentType: "image/png")
            )
            
            print("[INFO] [PhotoService/getPresignedURL]: Successfully received presigned URL")
            return response
        } catch(let error) {
            throw NetworkError.network(error)
        }
    }
}
