import Foundation

@MainActor
@Observable
final class GroupService {
    static let shared = GroupService()

    private(set) var groups: [Group] = []
    private(set) var currentGroup: Group?
    private(set) var isLoading = false
    private(set) var hasLoadedGroups = false
    var errorMessage: String?

    private let apiClient = APIClient.shared
    private let lastActiveGroupKey = "lastActiveGroupId"

    var hasGroups: Bool {
        !groups.isEmpty
    }

    private init() {}

    private func setCurrentGroup(_ group: Group?, refreshData: Bool) {
        currentGroup = group

        if let group {
            UserDefaults.standard.set(group.id, forKey: lastActiveGroupKey)
            WebSocketService.shared.switchToGroup(group.id)
            if refreshData {
                AppState.shared.refreshAllData()
            }
        } else {
            UserDefaults.standard.removeObject(forKey: lastActiveGroupKey)
            WebSocketService.shared.switchToGroup(nil)
        }
    }

    // MARK: - Load Groups

    func loadGroups() async {
        isLoading = true
        errorMessage = nil

        do {
            groups = try await apiClient.request(endpoint: "/groups")
            hasLoadedGroups = true

            // Restore last active group or select first
            if let lastGroupId = UserDefaults.standard.string(forKey: lastActiveGroupKey),
               let lastGroup = groups.first(where: { $0.id == lastGroupId }) {
                setCurrentGroup(lastGroup, refreshData: false)
            } else if let firstGroup = groups.first {
                selectGroup(firstGroup)
            } else {
                setCurrentGroup(nil, refreshData: false)
            }
        } catch {
            errorMessage = error.localizedDescription
            hasLoadedGroups = true
        }

        isLoading = false
    }

    // MARK: - Select Group

    func selectGroup(_ group: Group) {
        setCurrentGroup(group, refreshData: true)
    }

    // MARK: - Create Group

    func createGroup(name: String, description: String?) async throws -> Group {
        let request = CreateGroupRequest(name: name, description: description)
        let group: Group = try await apiClient.request(
            endpoint: "/groups",
            method: .post,
            body: request
        )

        groups.append(group)
        selectGroup(group)
        return group
    }

    // MARK: - Join Group

    func joinGroup(inviteCode: String) async throws -> Group {
        let request = RedeemInviteRequest(code: inviteCode)
        let group: Group = try await apiClient.request(
            endpoint: "/invites/redeem",
            method: .post,
            body: request
        )

        groups.append(group)
        selectGroup(group)
        return group
    }

    // MARK: - Leave Group

    func leaveGroup(_ group: Group) async throws {
        _ = try await apiClient.requestMessage(
            endpoint: "/groups/\(group.id)/leave",
            method: .post
        )

        groups.removeAll { $0.id == group.id }

        // If we left the current group, switch to another
        if currentGroup?.id == group.id {
            if let firstGroup = groups.first {
                selectGroup(firstGroup)
            } else {
                setCurrentGroup(nil, refreshData: false)
            }
        }
    }

    // MARK: - Clear

    func clear() {
        groups = []
        hasLoadedGroups = false
        errorMessage = nil
        setCurrentGroup(nil, refreshData: false)
    }
}
