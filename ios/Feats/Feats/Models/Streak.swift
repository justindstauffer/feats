import Foundation

struct Streak: Codable, Identifiable, Equatable {
    let id: String
    let userId: String
    let currentStreak: Int
    let longestStreak: Int
    let lastActivityDate: Date?
    let updatedAt: Date
    let user: User?

    enum CodingKeys: String, CodingKey {
        case id, user
        case userId = "user_id"
        case currentStreak = "current_streak"
        case longestStreak = "longest_streak"
        case lastActivityDate = "last_activity_date"
        case updatedAt = "updated_at"
    }
}
