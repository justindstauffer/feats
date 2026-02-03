import Foundation

struct Reaction: Codable, Identifiable, Equatable {
    let id: String
    let userId: String
    let postId: String
    let reactionType: ReactionType
    let createdAt: Date
    let user: User?

    enum CodingKeys: String, CodingKey {
        case id, user
        case userId = "user_id"
        case postId = "post_id"
        case reactionType = "reaction_type"
        case createdAt = "created_at"
    }
}

enum ReactionType: Int, Codable, CaseIterable {
    case like = 1
    case love = 2
    case fire = 3
    case strong = 4
    case clap = 5

    var emoji: String {
        switch self {
        case .like: return "👍"
        case .love: return "❤️"
        case .fire: return "🔥"
        case .strong: return "💪"
        case .clap: return "👏"
        }
    }
}

struct ReactionSummary: Codable, Equatable {
    let type: ReactionType
    let emoji: String
    let count: Int
}

struct ReactionsResponse: Codable, Equatable {
    let summary: [ReactionSummary]
    let reactions: [Reaction]
}

struct AddReactionRequest: Codable {
    let reactionType: Int

    enum CodingKeys: String, CodingKey {
        case reactionType = "reaction_type"
    }
}
