import Foundation

struct LoginRequest: Codable {
    let email: String
    let password: String
}

struct LoginResponse: Codable {
    let tokens: TokenPair
    let user: User
}

struct TokenPair: Codable {
    let accessToken: String
    let refreshToken: String
    let expiresAt: Date

    enum CodingKeys: String, CodingKey {
        case accessToken = "access_token"
        case refreshToken = "refresh_token"
        case expiresAt = "expires_at"
    }
}

struct RefreshRequest: Codable {
    let refreshToken: String

    enum CodingKeys: String, CodingKey {
        case refreshToken = "refresh_token"
    }
}

struct ChangePasswordRequest: Codable {
    let currentPassword: String
    let newPassword: String

    enum CodingKeys: String, CodingKey {
        case currentPassword = "current_password"
        case newPassword = "new_password"
    }
}

struct RegisterRequest: Codable {
    let email: String
    let password: String
    let name: String
    let inviteCode: String

    enum CodingKeys: String, CodingKey {
        case email, password, name
        case inviteCode = "invite_code"
    }
}
