import Foundation

struct User: Codable, Identifiable, Equatable {
    let id: String
    let email: String
    let name: String
    let profilePicture: String?
    let bio: String?
    let role: UserRole
    let createdAt: Date
    let updatedAt: Date

    enum CodingKeys: String, CodingKey {
        case id, email, name, bio, role
        case profilePicture = "profile_picture"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }
}

enum UserRole: String, Codable {
    case admin
    case user

    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        let rawValue = try container.decode(String.self)
        self = UserRole(rawValue: rawValue) ?? .user
    }
}

struct UpdateUserRequest: Codable {
    var name: String?
    var bio: String?
    var profilePicture: String?

    enum CodingKeys: String, CodingKey {
        case name, bio
        case profilePicture = "profile_picture"
    }
}
