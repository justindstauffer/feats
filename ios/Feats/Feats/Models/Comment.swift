import Foundation

struct Comment: Codable, Identifiable, Equatable {
    let id: String
    let postId: String
    let userId: String
    let parentId: String?
    let content: String
    let createdAt: Date
    let updatedAt: Date
    let user: User?
    let replies: [Comment]?

    enum CodingKeys: String, CodingKey {
        case id, content, user, replies
        case postId = "post_id"
        case userId = "user_id"
        case parentId = "parent_id"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }

    static func == (lhs: Comment, rhs: Comment) -> Bool {
        lhs.id == rhs.id
    }
}

struct CreateCommentRequest: Codable {
    let content: String
    let parentId: String?

    enum CodingKeys: String, CodingKey {
        case content
        case parentId = "parent_id"
    }
}

struct UpdateCommentRequest: Codable {
    let content: String
}
