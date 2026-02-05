import Foundation

struct BetaInvite: Codable, Identifiable {
    let id: String
    let code: String
    let createdBy: String
    let expiresAt: Date
    let maxUses: Int
    let useCount: Int
    let note: String?
    let createdAt: Date
    let creator: User?

    enum CodingKeys: String, CodingKey {
        case id, code, note, creator
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

struct CreateBetaInviteRequest: Codable {
    var maxUses: Int = 1
    var expiresIn: Int = 168 // 7 days in hours
    var note: String?

    enum CodingKeys: String, CodingKey {
        case note
        case maxUses = "max_uses"
        case expiresIn = "expires_in"
    }
}
