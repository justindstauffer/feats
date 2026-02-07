import Foundation

@MainActor
@Observable
class AppState {
    static let shared = AppState()

    var selectedTab: Int = 0
    var feedNeedsRefresh = false
    var challengesNeedRefresh = false
    var profileNeedsRefresh = false
    var streaksNeedRefresh = false

    private init() {
        setupWebSocketHandlers()
    }

    private func setupWebSocketHandlers() {
        let ws = WebSocketService.shared
        let groupService = GroupService.shared

        // When a post is created by someone else, refresh feed
        ws.onPostCreated = { [weak self] payload, groupId in
            guard groupService.currentGroup?.id == groupId else { return }
            self?.feedNeedsRefresh = true
        }

        // When a post is deleted, refresh feed
        ws.onPostDeleted = { [weak self] _, groupId in
            guard groupService.currentGroup?.id == groupId else { return }
            self?.feedNeedsRefresh = true
        }

        // When reactions change, refresh feed (for reaction counts)
        ws.onReactionAdded = { [weak self] _, groupId in
            guard groupService.currentGroup?.id == groupId else { return }
            self?.feedNeedsRefresh = true
        }

        ws.onReactionRemoved = { [weak self] _, groupId in
            guard groupService.currentGroup?.id == groupId else { return }
            self?.feedNeedsRefresh = true
        }

        // When comments are added, refresh feed (for comment counts)
        ws.onCommentCreated = { [weak self] _, groupId in
            guard groupService.currentGroup?.id == groupId else { return }
            self?.feedNeedsRefresh = true
        }

        ws.onCommentDeleted = { [weak self] _, groupId in
            guard groupService.currentGroup?.id == groupId else { return }
            self?.feedNeedsRefresh = true
        }

        // When challenges change, refresh challenges
        ws.onChallengeCreated = { [weak self] _, groupId in
            guard groupService.currentGroup?.id == groupId else { return }
            self?.challengesNeedRefresh = true
        }

        ws.onChallengeJoined = { [weak self] _, groupId in
            guard groupService.currentGroup?.id == groupId else { return }
            self?.challengesNeedRefresh = true
        }

        ws.onChallengeLeft = { [weak self] _, groupId in
            guard groupService.currentGroup?.id == groupId else { return }
            self?.challengesNeedRefresh = true
        }

        // When members join/leave, could refresh member list if visible
        ws.onMemberJoined = { [weak self] _, groupId in
            guard groupService.currentGroup?.id == groupId else { return }
            // Optionally trigger a notification or refresh
            self?.streaksNeedRefresh = true
        }

        ws.onMemberLeft = { [weak self] _, groupId in
            guard groupService.currentGroup?.id == groupId else { return }
            self?.streaksNeedRefresh = true
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

    func postCreated() {
        refreshAllData()
        navigateToFeed()
    }
}
