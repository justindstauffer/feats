import Foundation

struct Challenge: Codable, Identifiable, Equatable {
    let id: String
    let createdBy: String
    let title: String
    let description: String?
    let activityTypeId: String?
    let targetCount: Int
    let startDate: Date?
    let endDate: Date?
    let createdAt: Date
    let creator: User?
    let activityType: ActivityType?
    let participants: [ChallengeParticipant]?

    enum CodingKeys: String, CodingKey {
        case id, title, description, creator, participants
        case createdBy = "created_by"
        case activityTypeId = "activity_type_id"
        case targetCount = "target_count"
        case startDate = "start_date"
        case endDate = "end_date"
        case createdAt = "created_at"
        case activityType = "activity_type"
    }

    var isActive: Bool {
        let now = Date()
        if let start = startDate, now < start { return false }
        if let end = endDate, now > end { return false }
        return true
    }

    var isTimeBound: Bool {
        startDate != nil || endDate != nil
    }
}

struct ChallengeParticipant: Codable, Identifiable, Equatable {
    let id: String
    let challengeId: String
    let userId: String
    let progress: Int
    let completedAt: Date?
    let joinedAt: Date
    let user: User?

    enum CodingKeys: String, CodingKey {
        case id, progress, user
        case challengeId = "challenge_id"
        case userId = "user_id"
        case completedAt = "completed_at"
        case joinedAt = "joined_at"
    }

    var isCompleted: Bool {
        completedAt != nil
    }
}

struct CreateChallengeRequest: Codable {
    let title: String
    let description: String?
    let activityTypeId: String?
    let targetCount: Int
    let startDate: Date?
    let endDate: Date?

    enum CodingKeys: String, CodingKey {
        case title, description
        case activityTypeId = "activity_type_id"
        case targetCount = "target_count"
        case startDate = "start_date"
        case endDate = "end_date"
    }
}
