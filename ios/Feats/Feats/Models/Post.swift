import Foundation

struct Post: Codable, Identifiable, Equatable {
    let id: String
    let userId: String
    let activityTypeId: String
    let description: String?
    let createdAt: Date
    let updatedAt: Date
    let user: User?
    let activityType: ActivityType?
    let images: [PostImage]?
    let reactions: [Reaction]?
    let commentCount: Int?

    enum CodingKeys: String, CodingKey {
        case id, description, user, images, reactions
        case userId = "user_id"
        case activityTypeId = "activity_type_id"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
        case activityType = "activity_type"
        case commentCount = "comment_count"
    }

    static func == (lhs: Post, rhs: Post) -> Bool {
        let lhsReactionSignature = lhs.reactions?.map { "\($0.userId):\($0.reactionType.rawValue)" } ?? []
        let rhsReactionSignature = rhs.reactions?.map { "\($0.userId):\($0.reactionType.rawValue)" } ?? []
        let lhsImageSignature = lhs.images?.map(\.id) ?? []
        let rhsImageSignature = rhs.images?.map(\.id) ?? []

        return lhs.id == rhs.id &&
            lhs.description == rhs.description &&
            lhs.commentCount == rhs.commentCount &&
            lhsReactionSignature == rhsReactionSignature &&
            lhsImageSignature == rhsImageSignature
    }
}

struct PostImage: Codable, Identifiable, Equatable {
    let id: String
    let postId: String
    let displayOrder: Int
    let createdAt: Date

    enum CodingKeys: String, CodingKey {
        case id
        case postId = "post_id"
        case displayOrder = "display_order"
        case createdAt = "created_at"
    }
}

struct CreatePostRequest: Codable {
    let activityTypeId: String
    let description: String?

    enum CodingKeys: String, CodingKey {
        case activityTypeId = "activity_type_id"
        case description
    }
}

struct UpdatePostRequest: Codable {
    let description: String?
}
