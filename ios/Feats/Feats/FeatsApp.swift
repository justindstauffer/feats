import SwiftUI

@main
struct FeatsApp: App {
    @State private var authService = AuthService.shared

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environment(authService)
        }
    }
}

struct ContentView: View {
    @Environment(AuthService.self) private var authService
    @State private var isCheckingAuth = true

    var body: some View {
        Group {
            if isCheckingAuth {
                ProgressView("Loading...")
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else if authService.isAuthenticated {
                MainTabView()
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
}
