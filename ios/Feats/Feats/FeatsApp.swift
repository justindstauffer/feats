import SwiftUI
import UserNotifications

// MARK: - App Delegate for Push Notifications

class AppDelegate: NSObject, UIApplicationDelegate, UNUserNotificationCenterDelegate {
    func application(
        _ application: UIApplication,
        didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]? = nil
    ) -> Bool {
        UNUserNotificationCenter.current().delegate = self
        return true
    }

    func application(
        _ application: UIApplication,
        didRegisterForRemoteNotificationsWithDeviceToken deviceToken: Data
    ) {
        Task { @MainActor in
            PushNotificationService.shared.didRegisterForRemoteNotifications(deviceToken: deviceToken)
        }
    }

    func application(
        _ application: UIApplication,
        didFailToRegisterForRemoteNotificationsWithError error: Error
    ) {
        Task { @MainActor in
            PushNotificationService.shared.didFailToRegisterForRemoteNotifications(error: error)
        }
    }

    // Handle notification when app is in foreground
    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        willPresent notification: UNNotification
    ) async -> UNNotificationPresentationOptions {
        return [.banner, .sound, .badge]
    }

    // Handle notification tap
    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        didReceive response: UNNotificationResponse
    ) async {
        let userInfo = response.notification.request.content.userInfo
        await MainActor.run {
            PushNotificationService.shared.handleNotification(userInfo: userInfo)
        }
    }
}

// MARK: - Main App

@main
struct FeatsApp: App {
    @UIApplicationDelegateAdaptor(AppDelegate.self) var appDelegate
    @State private var authService = AuthService.shared
    @State private var groupService = GroupService.shared
    @Environment(\.scenePhase) private var scenePhase

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environment(authService)
                .environment(groupService)
                .preferredColorScheme(.light) // Force light mode for beta
        }
        .onChange(of: scenePhase) { _, newPhase in
            handleScenePhaseChange(newPhase)
        }
    }

    private func handleScenePhaseChange(_ phase: ScenePhase) {
        switch phase {
        case .active:
            // App came to foreground - resume WebSocket, refresh data, and clear badge
            Task {
                await WebSocketService.shared.resume()
                PushNotificationService.shared.clearBadge()
                // Refresh all data in case we missed WebSocket events while backgrounded
                AppState.shared.refreshAllData()
            }
        case .background:
            // App went to background - pause WebSocket to save battery
            WebSocketService.shared.pause()
        case .inactive:
            // Transitioning - do nothing
            break
        @unknown default:
            break
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
