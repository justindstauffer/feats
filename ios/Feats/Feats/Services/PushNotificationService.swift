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
        debugLog("Requesting push notification permission")
        let center = UNUserNotificationCenter.current()

        // Check current status first
        let settings = await center.notificationSettings()
        debugLog("Current authorization status: \(settings.authorizationStatus.rawValue)")

        do {
            let granted = try await center.requestAuthorization(options: [.alert, .sound, .badge])
            debugLog("Permission granted: \(granted)")
            self.isAuthorized = granted

            if granted {
                registerForRemoteNotifications()
            }

            return granted
        } catch {
            debugLog("Push notification permission error")
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
        debugLog("Received APNs device token")

        // Send token to backend
        Task {
            await sendTokenToBackend(tokenString)
        }
    }

    func didFailToRegisterForRemoteNotifications(error: Error) {
        debugLog("Failed to register for remote notifications")
    }

    // MARK: - Backend Registration

    private func sendTokenToBackend(_ token: String) async {
        guard AuthService.shared.isAuthenticated else {
            debugLog("Skipping token registration while unauthenticated")
            return
        }

        do {
            let request = RegisterDeviceRequest(token: token, platform: "ios")
            let _: MessageResponse = try await apiClient.request(
                endpoint: "/devices",
                method: .post,
                body: request
            )
            debugLog("Device token registered with backend")
        } catch {
            debugLog("Failed to register device token")
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
            debugLog("Device token unregistered from backend")
        } catch {
            debugLog("Failed to unregister device token")
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
                if let postID = userInfo["post_id"] as? String {
                    AppState.shared.navigateToPost(postID: postID)
                    debugLog("Navigate to post \(postID)")
                }
            case "comment":
                if let postID = userInfo["post_id"] as? String {
                    AppState.shared.navigateToPost(postID: postID)
                    debugLog("Navigate to post comments \(postID)")
                }
            case "reaction":
                if let postID = userInfo["post_id"] as? String {
                    AppState.shared.navigateToPost(postID: postID)
                    debugLog("Navigate to post \(postID)")
                }
            case "challenge":
                if userInfo["challenge_id"] as? String != nil {
                    AppState.shared.navigateToChallenges()
                    debugLog("Navigate to challenge")
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

    private func debugLog(_ message: String) {
        #if DEBUG
        print("PushNotificationService: \(message)")
        #endif
    }
}

// MARK: - Request Models

struct RegisterDeviceRequest: Codable {
    let token: String
    let platform: String
}
