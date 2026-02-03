import Foundation

struct Goal: Codable, Identifiable, Equatable {
    let id: String
    let userId: String
    let activityTypeId: String?
    let targetCount: Int
    let period: GoalPeriod
    let currentProgress: Int
    let periodStart: Date
    let createdAt: Date
    let updatedAt: Date
    let activityType: ActivityType?

    enum CodingKeys: String, CodingKey {
        case id, period
        case userId = "user_id"
        case activityTypeId = "activity_type_id"
        case targetCount = "target_count"
        case currentProgress = "current_progress"
        case periodStart = "period_start"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
        case activityType = "activity_type"
    }

    var isAchieved: Bool {
        currentProgress >= targetCount
    }

    var progressPercentage: Double {
        guard targetCount > 0 else { return 0 }
        return min(Double(currentProgress) / Double(targetCount), 1.0)
    }
}

enum GoalPeriod: String, Codable, CaseIterable {
    case daily
    case weekly
    case monthly

    var displayName: String {
        switch self {
        case .daily: return "Daily"
        case .weekly: return "Weekly"
        case .monthly: return "Monthly"
        }
    }
}

struct CreateGoalRequest: Codable {
    let activityTypeId: String?
    let targetCount: Int
    let period: String

    enum CodingKeys: String, CodingKey {
        case activityTypeId = "activity_type_id"
        case targetCount = "target_count"
        case period
    }
}

struct UpdateGoalRequest: Codable {
    let targetCount: Int?
    let period: String?

    enum CodingKeys: String, CodingKey {
        case targetCount = "target_count"
        case period
    }
}
