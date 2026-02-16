import Foundation

@MainActor
@Observable
class AppState {
    static let shared = AppState()

    var selectedTab: Int = 0
    var pendingPostNavigationID: String?
    var feedNeedsRefresh = false
    var challengesNeedRefresh = false
    var profileNeedsRefresh = false
    var streaksNeedRefresh = false

    private init() {
        setupWebSocketHandlers()
    }

    private func isCurrentGroup(_ groupId: String) -> Bool {
        GroupService.shared.currentGroup?.id == groupId
    }

    private func markFeedNeedsRefresh(for groupId: String) {
        guard isCurrentGroup(groupId) else { return }
        feedNeedsRefresh = true
    }

    private func markChallengesNeedRefresh(for groupId: String) {
        guard isCurrentGroup(groupId) else { return }
        challengesNeedRefresh = true
    }

    private func markStreaksNeedRefresh(for groupId: String) {
        guard isCurrentGroup(groupId) else { return }
        streaksNeedRefresh = true
    }

    private func setupWebSocketHandlers() {
        let ws = WebSocketService.shared

        // When a post is created by someone else, refresh feed
        ws.onPostCreated = { [weak self] _, groupId in
            self?.markFeedNeedsRefresh(for: groupId)
        }

        // When a post is deleted, refresh feed
        ws.onPostDeleted = { [weak self] _, groupId in
            self?.markFeedNeedsRefresh(for: groupId)
        }

        // When reactions change, refresh feed (for reaction counts)
        ws.onReactionAdded = { [weak self] _, groupId in
            self?.markFeedNeedsRefresh(for: groupId)
        }

        ws.onReactionRemoved = { [weak self] _, groupId in
            self?.markFeedNeedsRefresh(for: groupId)
        }

        // When comments are added, refresh feed (for comment counts)
        ws.onCommentCreated = { [weak self] _, groupId in
            self?.markFeedNeedsRefresh(for: groupId)
        }

        ws.onCommentDeleted = { [weak self] _, groupId in
            self?.markFeedNeedsRefresh(for: groupId)
        }

        // When challenges change, refresh challenges
        ws.onChallengeCreated = { [weak self] _, groupId in
            self?.markChallengesNeedRefresh(for: groupId)
        }

        ws.onChallengeJoined = { [weak self] _, groupId in
            self?.markChallengesNeedRefresh(for: groupId)
        }

        ws.onChallengeLeft = { [weak self] _, groupId in
            self?.markChallengesNeedRefresh(for: groupId)
        }

        // When members join/leave, could refresh member list if visible
        ws.onMemberJoined = { [weak self] _, groupId in
            // Optionally trigger a notification or refresh
            self?.markStreaksNeedRefresh(for: groupId)
        }

        ws.onMemberLeft = { [weak self] _, groupId in
            self?.markStreaksNeedRefresh(for: groupId)
        }
    }

    func refreshAllData() {
        feedNeedsRefresh = true
        challengesNeedRefresh = true
        profileNeedsRefresh = true
        streaksNeedRefresh = true
    }

    func navigateToFeed() {
        selectedTab = 0
    }

    func navigateToPost(postID: String) {
        selectedTab = 0
        pendingPostNavigationID = postID
        feedNeedsRefresh = true
    }

    func navigateToChallenges() {
        selectedTab = 1
    }

    func postCreated() {
        refreshAllData()
        navigateToFeed()
    }
}
