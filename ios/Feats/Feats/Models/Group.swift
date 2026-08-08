import Foundation

struct Group: Codable, Identifiable, Equatable {
    let id: String
    let name: String
    let description: String?
    let createdBy: String
    let createdAt: Date
    let updatedAt: Date
    let members: [GroupMember]?

    enum CodingKeys: String, CodingKey {
        case id, name, description, members
        case createdBy = "created_by"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }

    static func == (lhs: Group, rhs: Group) -> Bool {
        lhs.id == rhs.id
    }
}

struct GroupMember: Codable, Identifiable, Equatable {
    let id: String
    let groupId: String
    let userId: String
    let role: GroupRole
    let joinedAt: Date
    let user: User?

    enum CodingKeys: String, CodingKey {
        case id, role, user
        case groupId = "group_id"
        case userId = "user_id"
        case joinedAt = "joined_at"
    }
}

enum GroupRole: String, Codable {
    case admin
    case member
}

struct GroupInvite: Codable, Identifiable {
    let id: String
    let groupId: String
    let code: String
    let createdBy: String
    let expiresAt: Date
    let maxUses: Int
    let useCount: Int
    let createdAt: Date

    enum CodingKeys: String, CodingKey {
        case id, code
        case groupId = "group_id"
        case createdBy = "created_by"
        case expiresAt = "expires_at"
        case maxUses = "max_uses"
        case useCount = "use_count"
        case createdAt = "created_at"
    }

    var isExpired: Bool {
        Date() > expiresAt
    }

    var hasUsesRemaining: Bool {
        maxUses == 0 || useCount < maxUses
    }

    var isValid: Bool {
        !isExpired && hasUsesRemaining
    }

    var usesDescription: String {
        if maxUses == 0 {
            return "Unlimited uses (\(useCount) used)"
        }
        return "\(useCount)/\(maxUses) uses"
    }
}

struct CreateGroupInviteRequest: Codable {
    var maxUses: Int = 1
    var expiresIn: Int = 168 // 7 days in hours

    enum CodingKeys: String, CodingKey {
        case maxUses = "max_uses"
        case expiresIn = "expires_in"
    }
}

struct CreateGroupRequest: Codable {
    let name: String
    let description: String?
}

struct UpdateGroupRequest: Codable {
    let name: String?
    let description: String?
}

struct UpdateMemberRequest: Codable {
    let role: String
}

struct RedeemInviteRequest: Codable {
    let code: String
}
