import Foundation
import UserNotifications
import UIKit

@MainActor
@Observable
final class PushNotificationService {
    static let shared = PushNotificationService()

    var isAuthorized = false
    var deviceToken: String?

    private let apiClient = APIClient.shared

    private init() {}

    // MARK: - Request Permission

    func requestPermission() async -> Bool {
        print("🔔 Requesting push notification permission...")
        let center = UNUserNotificationCenter.current()

        // Check current status first
        let settings = await center.notificationSettings()
        print("🔔 Current authorization status: \(settings.authorizationStatus.rawValue)")

        do {
            let granted = try await center.requestAuthorization(options: [.alert, .sound, .badge])
            print("🔔 Permission granted: \(granted)")
            self.isAuthorized = granted

            if granted {
                registerForRemoteNotifications()
            }

            return granted
        } catch {
            print("🔔 Push notification permission error: \(error)")
            return false
        }
    }

    func checkPermissionStatus() async {
        let center = UNUserNotificationCenter.current()
        let settings = await center.notificationSettings()

        self.isAuthorized = settings.authorizationStatus == .authorized

        if isAuthorized {
            registerForRemoteNotifications()
        }
    }

    // MARK: - Register for Remote Notifications

    private func registerForRemoteNotifications() {
        UIApplication.shared.registerForRemoteNotifications()
    }

    // MARK: - Handle Device Token

    func didRegisterForRemoteNotifications(deviceToken: Data) {
        let tokenString = deviceToken.map { String(format: "%02.2hhx", $0) }.joined()
        self.deviceToken = tokenString
        print("Device token: \(tokenString)")

        // Send token to backend
        Task {
            await sendTokenToBackend(tokenString)
        }
    }

    func didFailToRegisterForRemoteNotifications(error: Error) {
        print("Failed to register for remote notifications: \(error)")
    }

    // MARK: - Backend Registration

    private func sendTokenToBackend(_ token: String) async {
        guard AuthService.shared.isAuthenticated else {
            print("Not authenticated, skipping token registration")
            return
        }

        do {
            let request = RegisterDeviceRequest(token: token, platform: "ios")
            let _: MessageResponse = try await apiClient.request(
                endpoint: "/devices",
                method: .post,
                body: request
            )
            print("Device token registered with backend")
        } catch {
            print("Failed to register device token: \(error)")
        }
    }

    func unregisterToken() async {
        guard let token = deviceToken else { return }

        do {
            let request = RegisterDeviceRequest(token: token, platform: "ios")
            let _: MessageResponse = try await apiClient.request(
                endpoint: "/devices",
                method: .delete,
                body: request
            )
            print("Device token unregistered from backend")
        } catch {
            print("Failed to unregister device token: \(error)")
        }
    }

    // MARK: - Re-register on Login

    func onUserLogin() async {
        if let token = deviceToken {
            await sendTokenToBackend(token)
        }
    }

    // MARK: - Handle Notifications

    func handleNotification(userInfo: [AnyHashable: Any]) {
        // Parse notification data
        if let type = userInfo["type"] as? String {
            switch type {
            case "post":
                if let postId = userInfo["post_id"] as? String {
                    // Navigate to post
                    print("Navigate to post: \(postId)")
                }
            case "comment":
                if let postId = userInfo["post_id"] as? String {
                    // Navigate to post with comments
                    print("Navigate to post comments: \(postId)")
                }
            case "reaction":
                if let postId = userInfo["post_id"] as? String {
                    // Navigate to post
                    print("Navigate to post: \(postId)")
                }
            case "challenge":
                if let challengeId = userInfo["challenge_id"] as? String {
                    // Navigate to challenge
                    print("Navigate to challenge: \(challengeId)")
                }
            default:
                break
            }
        }
    }

    // MARK: - Clear Badge

    func clearBadge() {
        UNUserNotificationCenter.current().setBadgeCount(0)
    }
}

// MARK: - Request Models

struct RegisterDeviceRequest: Codable {
    let token: String
    let platform: String
}
