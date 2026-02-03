import Foundation

struct APIResponse<T: Codable>: Codable {
    let success: Bool
    let data: T?
    let error: APIError?
}

struct APIError: Codable, Error, LocalizedError {
    let code: String
    let message: String

    var errorDescription: String? {
        message
    }
}

struct PaginatedResponse<T: Codable>: Codable {
    let success: Bool
    let data: T?
    let pagination: Pagination?
    let error: APIError?
}

struct Pagination: Codable {
    let page: Int
    let perPage: Int
    let total: Int
    let totalPages: Int

    enum CodingKeys: String, CodingKey {
        case page, total
        case perPage = "per_page"
        case totalPages = "total_pages"
    }
}

struct MessageResponse: Codable {
    let message: String
}

// Error codes from API
enum APIErrorCode: String {
    case unauthorized = "UNAUTHORIZED"
    case forbidden = "FORBIDDEN"
    case notFound = "NOT_FOUND"
    case validationError = "VALIDATION_ERROR"
    case conflict = "CONFLICT"
    case rateLimited = "RATE_LIMITED"
    case internalError = "INTERNAL_ERROR"
    case accountLocked = "ACCOUNT_LOCKED"
    case invalidCredentials = "INVALID_CREDENTIALS"
    case tokenExpired = "TOKEN_EXPIRED"
    case tokenInvalid = "TOKEN_INVALID"
}
