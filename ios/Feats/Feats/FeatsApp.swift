import SwiftUI

@main
struct FeatsApp: App {
    @State private var authService = AuthService.shared
    @State private var groupService = GroupService.shared

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environment(authService)
                .environment(groupService)
                .preferredColorScheme(.light) // Force light mode for beta
        }
    }
}

struct ContentView: View {
    @Environment(AuthService.self) private var authService
    @Environment(GroupService.self) private var groupService
    @State private var isCheckingAuth = true

    var body: some View {
        SwiftUI.Group {
            if isCheckingAuth {
                ProgressView("Loading...")
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else if authService.isAuthenticated {
                if groupService.hasLoadedGroups && !groupService.hasGroups {
                    GroupOnboardingView()
                } else if groupService.currentGroup != nil {
                    MainTabView()
                } else {
                    ProgressView("Loading groups...")
                        .frame(maxWidth: .infinity, maxHeight: .infinity)
                }
            } else {
                LoginView()
            }
        }
        .task {
            await authService.checkAuthState()
            isCheckingAuth = false
        }
    }
}

#Preview {
    ContentView()
        .environment(AuthService.shared)
        .environment(GroupService.shared)
}
