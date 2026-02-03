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

    private init() {}

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
