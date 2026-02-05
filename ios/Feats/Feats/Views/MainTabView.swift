import SwiftUI

struct MainTabView: View {
    @Bindable private var appState = AppState.shared
    @State private var groupService = GroupService.shared

    var body: some View {
        TabView(selection: $appState.selectedTab) {
            FeedView()
                .tabItem {
                    Label("Feed", systemImage: "house.fill")
                }
                .tag(0)

            ChallengesView()
                .tabItem {
                    Label("Challenges", systemImage: "flag.fill")
                }
                .tag(1)

            CreatePostView()
                .tabItem {
                    Label("Post", systemImage: "plus.circle.fill")
                }
                .tag(2)

            LeaderboardView()
                .tabItem {
                    Label("Streaks", systemImage: "flame.fill")
                }
                .tag(3)

            ProfileView()
                .tabItem {
                    Label("Profile", systemImage: "person.fill")
                }
                .tag(4)
        }
        .environment(appState)
        .environment(groupService)
    }
}

#Preview {
    MainTabView()
        .environment(AuthService.shared)
        .environment(AppState.shared)
        .environment(GroupService.shared)
}
