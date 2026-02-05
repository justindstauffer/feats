import Foundation

@MainActor
@Observable
final class AuthService {
    static let shared = AuthService()

    private(set) var currentUser: User?
    private(set) var isAuthenticated = false
    private(set) var isLoading = false

    private let apiClient = APIClient.shared
    private let keychain = KeychainService.shared

    private init() {}

    // MARK: - Authentication State

    func checkAuthState() async {
        // Try to restore session from keychain
        guard let _ = try? keychain.getRefreshToken() else {
            isAuthenticated = false
            return
        }

        // Try to get current user to validate token
        do {
            let user: User = try await apiClient.request(endpoint: "/users/me")
            self.currentUser = user
            self.isAuthenticated = true

            // Load user's groups
            await GroupService.shared.loadGroups()
        } catch {
            // Token invalid, clear everything
            logout()
        }
    }

    // MARK: - Register

    func register(email: String, password: String, name: String, inviteCode: String) async throws {
        isLoading = true
        defer { isLoading = false }

        let request = RegisterRequest(email: email, password: password, name: name, inviteCode: inviteCode)

        let response: LoginResponse = try await apiClient.request(
            endpoint: "/auth/register",
            method: .post,
            body: request,
            authenticated: false
        )

        // Save tokens
        apiClient.setAccessToken(response.tokens.accessToken, expiresAt: response.tokens.expiresAt)
        try keychain.saveRefreshToken(response.tokens.refreshToken)

        // Set user
        self.currentUser = response.user
        self.isAuthenticated = true

        // Load user's groups
        await GroupService.shared.loadGroups()
    }

    // MARK: - Login

    func login(email: String, password: String) async throws {
        isLoading = true
        defer { isLoading = false }

        let request = LoginRequest(email: email, password: password)

        let response: LoginResponse = try await apiClient.request(
            endpoint: "/auth/login",
            method: .post,
            body: request,
            authenticated: false
        )

        // Save tokens
        apiClient.setAccessToken(response.tokens.accessToken, expiresAt: response.tokens.expiresAt)
        try keychain.saveRefreshToken(response.tokens.refreshToken)

        // Set user
        self.currentUser = response.user
        self.isAuthenticated = true

        // Load user's groups
        await GroupService.shared.loadGroups()
    }

    // MARK: - Logout

    func logout() {
        Task {
            // Try to logout on server (ignore errors)
            try? await apiClient.requestMessage(endpoint: "/auth/logout", method: .post)
        }

        apiClient.clearTokens()
        currentUser = nil
        isAuthenticated = false

        // Clear groups
        GroupService.shared.clear()
    }

    // MARK: - Password

    func changePassword(currentPassword: String, newPassword: String) async throws {
        let request = ChangePasswordRequest(
            currentPassword: currentPassword,
            newPassword: newPassword
        )

        _ = try await apiClient.requestMessage(
            endpoint: "/auth/password/change",
            method: .post,
            body: request
        )

        // After password change, need to re-login
        logout()
    }

    // MARK: - Profile

    func updateProfile(name: String? = nil, bio: String? = nil) async throws {
        let request = UpdateUserRequest(name: name, bio: bio)

        let user: User = try await apiClient.request(
            endpoint: "/users/me",
            method: .put,
            body: request
        )

        self.currentUser = user
    }

    func refreshCurrentUser() async throws {
        let user: User = try await apiClient.request(endpoint: "/users/me")
        self.currentUser = user
    }
}
