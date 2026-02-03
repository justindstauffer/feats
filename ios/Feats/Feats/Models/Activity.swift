import Foundation

struct ActivityType: Codable, Identifiable, Equatable, Hashable {
    let id: String
    let name: String
    let icon: String?
    let isSystem: Bool
    let createdBy: String?
    let createdAt: Date

    enum CodingKeys: String, CodingKey {
        case id, name, icon
        case isSystem = "is_system"
        case createdBy = "created_by"
        case createdAt = "created_at"
    }

    var displayIcon: String {
        icon ?? "figure.walk"
    }
}

struct CreateActivityRequest: Codable {
    let name: String
    let icon: String?
}
