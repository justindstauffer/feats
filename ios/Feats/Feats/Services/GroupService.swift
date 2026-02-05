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
                currentGroup = lastGroup
            } else if let firstGroup = groups.first {
                selectGroup(firstGroup)
            } else {
                currentGroup = nil
            }
        } catch {
            errorMessage = error.localizedDescription
            hasLoadedGroups = true
        }

        isLoading = false
    }

    // MARK: - Select Group

    func selectGroup(_ group: Group) {
        currentGroup = group
        UserDefaults.standard.set(group.id, forKey: lastActiveGroupKey)
        AppState.shared.refreshAllData()
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
                currentGroup = nil
                UserDefaults.standard.removeObject(forKey: lastActiveGroupKey)
            }
        }
    }

    // MARK: - Clear

    func clear() {
        groups = []
        currentGroup = nil
        hasLoadedGroups = false
        errorMessage = nil
        UserDefaults.standard.removeObject(forKey: lastActiveGroupKey)
    }
}
