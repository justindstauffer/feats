# iOS App Development Progress

## Overview

This document tracks the development progress of the Feats iOS app. Use this as a reference for future coding-agent sessions.

Last updated: 2026-02-21

Current status:
- iOS remains the most complete and polished client.
- Notification/deep-link and group-scoped flows are live.
- iOS behavior continues to serve as reference parity target for Android implementation decisions.

Note: canonical day-to-day agent context now lives in:
- `AGENTS.md`
- `docs/status/now.md`
- `context/project-map.yaml`

## Tech Stack

- **Platform:** iOS 17+
- **Framework:** SwiftUI
- **State Management:** @Observable (iOS 17+)
- **Networking:** URLSession with async/await
- **Real-Time:** WebSockets (implemented)
- **Secure Storage:** Keychain Services

## Project Structure

```
ios/
├── README.md                    # Setup instructions
└── Feats/
    ├── FeatsApp.swift           # App entry point
    ├── Info.plist               # App configuration
    ├── Models/
    │   ├── User.swift           # User model
    │   ├── Auth.swift           # Login/token models
    │   ├── Activity.swift       # Activity types
    │   ├── Post.swift           # Posts and images
    │   ├── Reaction.swift       # Reactions (5 types)
    │   ├── Comment.swift        # Threaded comments
    │   ├── Streak.swift         # Streak tracking
    │   ├── Challenge.swift      # Challenges
    │   ├── Goal.swift           # Personal goals
    │   ├── Group.swift          # Groups and memberships (NEW)
    │   ├── BetaInvite.swift     # Beta invite codes (NEW)
    │   └── APIResponse.swift    # API response wrappers
    ├── Services/
    │   ├── KeychainService.swift  # Secure token storage
    │   ├── APIClient.swift        # HTTP client with auth
    │   ├── AuthService.swift      # Authentication state
    │   ├── GroupService.swift     # Group state management
    │   ├── WebSocketService.swift # Real-time updates (NEW)
    │   └── AppState.swift         # App-wide state management
    ├── Views/
    │   ├── MainTabView.swift      # Tab navigation
    │   ├── Auth/
    │   │   ├── LoginView.swift
    │   │   └── RegisterView.swift   # NEW
    │   ├── Admin/                   # NEW folder
    │   │   └── BetaInvitesView.swift
    │   ├── Onboarding/            # NEW folder
    │   │   └── GroupOnboardingView.swift
    │   ├── Groups/
    │   │   ├── CreateGroupView.swift
    │   │   ├── JoinGroupView.swift
    │   │   ├── GroupSwitcherView.swift
    │   │   └── GroupInvitesView.swift  # NEW
    │   ├── Feed/
    │   │   ├── FeedView.swift
    │   │   └── PostDetailView.swift
    │   ├── Post/
    │   │   └── CreatePostView.swift
    │   ├── Profile/
    │   │   ├── ProfileView.swift
    │   │   ├── EditProfileView.swift
    │   │   └── ChangePasswordView.swift
    │   ├── Challenges/
    │   │   ├── ChallengesView.swift
    │   │   └── CreateChallengeView.swift
    │   ├── Goals/
    │   │   └── LeaderboardView.swift
    │   └── Components/
    │       ├── PostCard.swift
    │       ├── GroupHeader.swift         # Group switcher button (NEW)
    │       └── AuthenticatedImage.swift  # Image loading with auth
    ├── ViewModels/              # Empty - using @Observable in views
    └── Utilities/               # Empty - for future utilities
```

## Xcode Project Setup

The Xcode project has been created. If starting fresh:

1. Open Xcode
2. File → New → Project
3. Choose "App" under iOS
4. Configure:
   - Product Name: `Feats`
   - Team: Your Apple Developer Team
   - Organization Identifier: `com.jstauff`
   - Interface: SwiftUI
   - Language: Swift
5. Save in the `ios` folder (creates `ios/Feats.xcodeproj`)
6. Delete auto-generated `ContentView.swift` and `FeatsApp.swift`
7. Add existing files: File → Add Files to "Feats"
8. Select the `Feats` folder, ensure "Create groups" is selected
9. Build and run

## Completed Features

### Authentication
- [x] Login view with email/password
- [x] JWT access token handling
- [x] Refresh token storage in Keychain
- [x] Automatic token refresh before expiry
- [x] Logout functionality
- [x] Session persistence across app launches

### Feed
- [x] Paginated post list
- [x] Pull-to-refresh
- [x] Post card with user info, activity type, images
- [x] Post detail view
- [x] Reaction display and adding/removing
- [x] Optimistic UI updates for reactions
- [x] Switch reactions with single tap (no need to deselect first)
- [x] Comments display
- [x] Add comment functionality
- [x] Delete post (context menu in feed, toolbar button in detail view)
- [x] Authenticated image loading (images require auth token)
- [x] Auto-refresh after creating a post

### Create Post
- [x] Activity type picker (Achievement type hidden - system only)
- [x] Photo selection (up to 4)
- [x] Description input
- [x] Image upload
- [x] Auto-redirect to Feed after posting
- [x] Auto-refresh of all app data after posting

### Profile
- [x] User info display
- [x] Current streak display
- [x] Longest streak display
- [x] Goals list with progress
- [x] Edit profile (name, bio)
- [x] Change password with validation
- [x] Logout confirmation

### Challenges
- [x] Challenge list with Active/Completed tabs
- [x] Challenge card with progress
- [x] Join/leave challenge
- [x] Create challenge form
- [x] Activity type filter
- [x] Date range selection
- [x] Pull-to-refresh
- [x] Auto-refresh when navigating to tab after post creation
- [x] Progress tracking updates correctly
- [x] Completed challenges move to "Completed" tab
- [x] Auto-post to feed when challenge is completed (🏆 Achievement)
- [x] Completion date displayed on completed challenges
- [x] Green styling for completed participants

### Streaks
- [x] Leaderboard view
- [x] Rank display with medals
- [x] Current user highlighting

### App-Wide State Management
- [x] AppState singleton for cross-tab communication
- [x] Automatic data refresh flags
- [x] Tab navigation control from any view
- [x] Post creation triggers refresh of Feed, Challenges, Profile, Streaks

### Multi-Tenancy / Groups
- [x] Group model and GroupService
- [x] Onboarding flow (create/join group)
- [x] Group switcher UI in navigation bar
- [x] GroupHeader component for all main views
- [x] All content views updated for group-scoped API
- [x] Last active group persistence via UserDefaults
- [x] Automatic data refresh on group switch
- [x] Group-scoped API endpoints for posts, challenges, streaks, goals

## API Client Features

- [x] Base URL configuration (debug vs release)
- [x] JSON encoding/decoding with custom date handling
- [x] Bearer token authentication
- [x] Automatic 401 handling with token clear
- [x] Multipart form upload for images
- [x] Paginated request support
- [x] Generic request methods
- [x] Authenticated image fetching (`fetchImageData`)
- [x] Group-scoped request methods (`groupRequest`, `groupRequestPaginated`, `groupRequestMessage`, `groupUploadImage`)

## Recent Changes (Feb 2026)

### Beta Invite System
Added invite-only registration for beta testing:

**Backend:**
- `POST /auth/register` - Public registration with beta invite code
- `POST /admin/beta-invites` - Create invite codes (admin only)
- `GET /admin/beta-invites` - List all invite codes (admin only)
- `DELETE /admin/beta-invites/:id` - Delete invite code (admin only)

**iOS:**
- `RegisterView` - Registration form with invite code, email, password, name
- Updated `LoginView` with "Create Account" link
- Updated `AuthService` with `register()` method
- `BetaInvite` model
- `BetaInvitesView` - Admin UI to create, view, and share invite codes
- Admin section in `ProfileView` with link to Beta Invites

**Invite Code Features:**
- Format: `XXXX-XXXX-XXXX` (alphanumeric, no ambiguous chars)
- Configurable max uses (0 = unlimited, default 1)
- Configurable expiration (default 7 days)
- Optional note field for tracking who codes are for

### Multi-Tenancy UI (Group-Based)
Added support for group-based multi-tenancy where all content is scoped to the selected group:

**New Files:**
- `Models/Group.swift` - Group, GroupMember, GroupInvite, GroupRole models
- `Services/GroupService.swift` - Singleton for managing group state
- `Views/Onboarding/GroupOnboardingView.swift` - First-time group setup
- `Views/Groups/CreateGroupView.swift` - Create new group form
- `Views/Groups/JoinGroupView.swift` - Join group via invite code
- `Views/Groups/GroupSwitcherView.swift` - Modal for switching groups
- `Views/Components/GroupHeader.swift` - Navigation bar group button

**Modified Files:**
- `APIClient.swift` - Added group-scoped request methods
- `AuthService.swift` - Load/clear groups on auth changes
- `FeatsApp.swift` - Added onboarding flow check
- `MainTabView.swift` - Added GroupService to environment
- `FeedView.swift` - Added GroupHeader, group-scoped API
- `PostDetailView.swift` - Group-scoped reactions/comments
- `ChallengesView.swift` - Added GroupHeader, group-scoped API
- `CreateChallengeView.swift` - Group-scoped API
- `LeaderboardView.swift` - Added GroupHeader, group-scoped API
- `CreatePostView.swift` - Group-scoped post/activities
- `ProfileView.swift` - Group-scoped streak/goals

**Flow:**
1. Login → Groups load automatically
2. If no groups → Onboarding shown (Create or Join)
3. After creating/joining → Main UI with group selected
4. Tap GroupHeader → Switcher modal → Select different group → Content refreshes
5. Last active group persisted in UserDefaults



### Challenge Completed Tab
- Added segmented control with "Active" and "Completed" tabs
- Active tab shows challenges that are ongoing and not yet completed by user
- Completed tab shows challenges user has finished
- Completed challenges show green "Completed" badge and completion date
- Participants who completed show with green avatars
- Shows count of how many participants have completed

### Auto-Post on Challenge Completion
- When a user completes a challenge, an automatic post is created
- Uses special "Achievement" activity type (🏆)
- Post says: "🎉 Completed the '[Challenge Name]' challenge!"
- Achievement activity type is hidden from manual post creation

### Reaction Improvements
- Optimistic UI updates - reactions update immediately before API call
- Can switch reactions with single tap (previously had to deselect first)
- Reaction summary updates correctly when adding/removing reactions
- Added `Equatable` to `ReactionSummary` for proper SwiftUI change detection

### Post Deletion
- Delete button in PostDetailView toolbar (trash icon)
- Confirmation dialog before deleting
- Only visible to post owner or admin
- Context menu delete in FeedView still works (long-press)
- Feed auto-refreshes after deletion

### AuthenticatedImage Component
Created `AuthenticatedImage.swift` to load images with Bearer token authentication. SwiftUI's built-in `AsyncImage` doesn't support custom headers, so this component:
- Uses `APIClient.fetchImageData()` to fetch with auth
- Shows loading indicator while fetching
- Shows placeholder on failure
- Uses `GeometryReader` for proper sizing

### Challenge Progress Fix
Fixed issues with challenge progress not updating:
- Backend date comparisons now use configured timezone (America/New_York)
- Challenge start/end date comparisons fixed for same-day edge cases
- Added debug logging to `UpdateProgressForActivity` in backend

### Challenge Model Fix
Removed custom `Equatable` implementation from `Challenge` struct that only compared IDs. Now uses synthesized equality that compares all fields including participants' progress, allowing SwiftUI to detect changes.

### Pull-to-Refresh Fix
Moved `.refreshable` modifier directly onto the `List` in `ChallengesView` instead of the outer `Group` to ensure it works correctly.

### Post Creation Flow
- After creating a post, user is automatically redirected to Feed tab
- All tabs (Feed, Challenges, Profile, Streaks) are flagged for refresh
- Each view checks refresh flag on `onAppear` and reloads data if needed

### Group Invites UI
- [x] GroupInvitesView for creating/managing group invites
- [x] Invite button in GroupSwitcherView for group admins
- [x] Share invite codes via copy or system share sheet
- [x] Create invites with usage limits and expiration

### Deployment (Feb 2026)
- [x] Backend deployed to DigitalOcean droplet
- [x] API available at https://feats-api.jstauff.com
- [x] SSL configured with Let's Encrypt
- [x] Docker containerized deployment
- [x] iOS production URL configured

### UI Fixes
- [x] Force light mode for beta (.preferredColorScheme(.light))
- [x] Keyboard dismissal on CreatePostView (Done button + scroll dismiss)
- [x] App icon added (trophy emoji)

### Phase 3 Service Cleanup (Feb 2026)
- [x] Refactored service-layer internals in `AuthService`, `APIClient`, `WebSocketService`
- [x] Cleaned state/helper flow in `AppState` and `GroupService`
- [x] No intended behavior changes; cleanup focused on maintainability and clearer service boundaries

### Real-Time Updates (WebSockets) - COMPLETED
- [x] Backend WebSocket hub infrastructure (`internal/websocket/hub.go`)
- [x] Backend WebSocket client handling (`internal/websocket/client.go`)
- [x] Backend event types and broadcasting (`internal/websocket/events.go`)
- [x] Backend WebSocket upgrade handler (`internal/handlers/websocket.go`)
- [x] Backend handlers broadcast events (post, reaction, comment, challenge, group)
- [x] iOS WebSocketService (`Services/WebSocketService.swift`)
- [x] iOS event type definitions and payload models
- [x] iOS AuthService connects/disconnects WebSocket on auth changes
- [x] iOS AppState handles WebSocket events for auto-refresh
- [x] Views refresh immediately when WebSocket events received (onChange handlers)
- [x] App lifecycle handling (pause on background, resume on foreground)
- [x] Group subscription management (subscribe/unsubscribe on group switch)
- [x] Connection state tracking with reconnection logic
- [x] Nginx configured for WebSocket proxying

**WebSocket Endpoint:** `wss://feats-api.jstauff.com/ws?token=<jwt>`

**Nginx Config Required:**
```nginx
location /ws {
    proxy_pass http://127.0.0.1:8080;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_read_timeout 86400;
}
```

---

## WebSocket Implementation Details

### Overview
Real-time updates via WebSocket connections. When someone posts, reacts, comments, or joins a challenge, all connected clients in the same group see updates immediately.

### Backend Implementation

**Files Created:**
| File | Description |
|------|-------------|
| `internal/websocket/hub.go` | Central hub managing all connections and group subscriptions |
| `internal/websocket/client.go` | Individual WebSocket client with read/write pumps |
| `internal/websocket/events.go` | Event types and payload structures |
| `internal/handlers/websocket.go` | HTTP upgrade handler with JWT auth via query param |

**Files Modified:**
| File | Changes |
|------|---------|
| `cmd/api/main.go` | Initialize hub, add `/ws` route |
| `internal/handlers/post.go` | Broadcast `post.created`, `post.deleted` |
| `internal/handlers/reaction.go` | Broadcast `reaction.added`, `reaction.removed` |
| `internal/handlers/comment.go` | Broadcast `comment.created`, `comment.deleted` |
| `internal/handlers/challenge.go` | Broadcast `challenge.created`, `challenge.joined`, `challenge.left` |
| `internal/handlers/group.go` | Broadcast `member.joined`, `member.left` |
| `go.mod` | Added `github.com/gorilla/websocket` dependency |

**Event Types:**
- `post.created` / `post.deleted`
- `reaction.added` / `reaction.removed`
- `comment.created` / `comment.deleted`
- `challenge.created` / `challenge.joined` / `challenge.left`
- `member.joined` / `member.left`

### iOS Implementation

**Files Created:**
| File | Description |
|------|-------------|
| `Services/WebSocketService.swift` | WebSocket client with event handling, reconnection, lifecycle |

**Files Modified:**
| File | Changes |
|------|---------|
| `FeatsApp.swift` | Handle ScenePhase for pause/resume WebSocket |
| `AuthService.swift` | Connect WebSocket on login, disconnect on logout |
| `GroupService.swift` | Call `switchToGroup()` when changing groups |
| `AppState.swift` | Set up WebSocket event handlers for auto-refresh |
| `APIClient.swift` | Added `webSocketURL()` method |
| `FeedView.swift` | Added `onChange` to refresh on flag change |
| `ChallengesView.swift` | Added `onChange` to refresh on flag change |
| `LeaderboardView.swift` | Added `onChange` to refresh on flag change |

**Connection States:**
```swift
enum WebSocketConnectionState: Equatable {
    case disconnected
    case connecting
    case connected
    case reconnecting(attempt: Int, maxAttempts: Int)
    case failed
}
```

**Key Methods:**
- `connect()` - Connect and subscribe to current group
- `disconnect()` - Clean disconnect (logout)
- `pause()` - Pause when app backgrounds
- `resume()` - Resume when app foregrounds
- `switchToGroup(_ groupId)` - Unsubscribe old, subscribe new
- `retryConnection()` - Manual retry after failure

**Reconnection Behavior:**
| Scenario | Behavior |
|----------|----------|
| Connection drops | Auto-reconnect with exponential backoff (2s, 4s, 8s, 16s, 30s max) |
| 5 consecutive failures | Stops, state = `.failed`, user can call `retryConnection()` |
| App backgrounds | Pauses cleanly (no reconnect attempts wasted) |
| App foregrounds | Resumes immediately, resets retry counter |

---

## Next Feature: Push Notifications (APNs)

### Overview
Push notifications alert users when something happens while the app is closed or backgrounded. Works alongside WebSockets - WebSockets for live updates when app is open, push notifications for background alerts.

### What Triggers Notifications
- Someone posts in your group
- Someone reacts to your post
- Someone comments on your post
- Someone joins a challenge you're in
- Someone completes a challenge you're in
- New member joins your group
- Challenge you're in is about to end (reminder)

### Apple Developer Setup

1. **Create APNs Key** (Apple Developer Portal):
   - Certificates, Identifiers & Profiles → Keys → +
   - Enable "Apple Push Notifications service (APNs)"
   - Download the `.p8` file (save securely - can only download once)
   - Note the Key ID and Team ID

2. **Enable Push in App ID**:
   - Identifiers → Select Feats app ID
   - Enable "Push Notifications" capability

3. **Add to Xcode**:
   - Target → Signing & Capabilities → + Capability → Push Notifications

### Backend Implementation

**New Files:**
- `internal/push/apns.go` - APNs client for sending notifications
- `internal/push/notifications.go` - Notification templates and logic
- `internal/models/device.go` - Device token storage (already exists partially)

**APNs Client:**
```go
type APNsClient struct {
    keyID      string
    teamID     string
    privateKey *ecdsa.PrivateKey
    bundleID   string
    production bool
}

func (c *APNsClient) Send(token string, notification Notification) error
```

**Notification Types:**
```go
type NotificationType string

const (
    NotifyNewPost       NotificationType = "new_post"
    NotifyReaction      NotificationType = "reaction"
    NotifyComment       NotificationType = "comment"
    NotifyChallengeJoin NotificationType = "challenge_join"
    NotifyMemberJoined  NotificationType = "member_joined"
)
```

**Environment Variables:**
```
APNS_KEY_PATH=/path/to/AuthKey_XXXXXX.p8
APNS_KEY_ID=XXXXXXXXXX
APNS_TEAM_ID=XXXXXXXXXX
APNS_BUNDLE_ID=com.jstauff.Feats
APNS_PRODUCTION=false  # true for App Store builds
```

### iOS Implementation

**Request Permission** (on first launch or login):
```swift
func requestNotificationPermission() async -> Bool {
    let center = UNUserNotificationCenter.current()
    do {
        let granted = try await center.requestAuthorization(options: [.alert, .badge, .sound])
        if granted {
            await MainActor.run {
                UIApplication.shared.registerForRemoteNotifications()
            }
        }
        return granted
    } catch {
        return false
    }
}
```

**Register Device Token:**
```swift
// In AppDelegate or FeatsApp
func application(_ application: UIApplication,
                 didRegisterForRemoteNotificationsWithDeviceToken deviceToken: Data) {
    let token = deviceToken.map { String(format: "%02.2hhx", $0) }.joined()
    Task {
        try? await APIClient.shared.request(
            endpoint: "/devices",
            method: .post,
            body: ["token": token, "platform": "ios"]
        )
    }
}
```

**Handle Notifications:**
```swift
// Tap notification → open relevant screen
func userNotificationCenter(_ center: UNUserNotificationCenter,
                            didReceive response: UNNotificationResponse) async {
    let userInfo = response.notification.request.content.userInfo
    // Navigate to post, challenge, etc. based on notification type
}
```

### Files to Create

**Backend:**
| File | Description |
|------|-------------|
| `internal/push/apns.go` | APNs HTTP/2 client |
| `internal/push/notifications.go` | Notification content builders |

**iOS:**
| File | Description |
|------|-------------|
| `Services/NotificationService.swift` | Permission, registration, handling |

### Files to Modify

**Backend:**
- `internal/handlers/post.go` - Send push after post created
- `internal/handlers/reaction.go` - Send push to post owner
- `internal/handlers/comment.go` - Send push to post owner
- `internal/handlers/challenge.go` - Send push to participants
- `internal/handlers/group.go` - Send push on member joined
- `.env.production` - Add APNs configuration

**iOS:**
- `FeatsApp.swift` - Set up notification delegate
- `AuthService.swift` - Register device on login, unregister on logout

### Implementation Order

1. **Apple Developer Setup** - Create key, enable capabilities
2. **Backend APNs Client** - HTTP/2 client with JWT auth
3. **Backend Device Storage** - Store/remove device tokens
4. **iOS Permission & Registration** - Request permission, send token
5. **Backend Notification Triggers** - Send on post/reaction/comment
6. **iOS Deep Linking** - Tap notification → navigate to content
7. **Testing** - TestFlight with production APNs

---

## Future Session: Features & Improvements

### Challenge Participants View
- [ ] Add a tab or separate view inside challenges to show all participants
- [ ] Display participant progress, join date, completion status
- [ ] Consider showing leaderboard within each challenge

### Post Deletion - Image Cleanup
- [ ] Verify that images attached to deleted posts are removed from server storage
- [ ] Check backend `DeletePost` to ensure image files are cleaned up
- [ ] Consider soft-delete behavior for images

### Challenge Progress on Post Deletion
- [ ] Decide how to handle challenge progress when a counted post is deleted
- [ ] Options:
  - Decrement progress when post is deleted
  - Keep progress as-is (already counted)
  - Only allow deletion if it won't affect completed challenges
- [ ] Implement chosen behavior in backend
- [ ] Update iOS to refresh challenge data after post deletion

### Group Management (Future)
- [ ] Badge counts per group in switcher
- [ ] Group admin settings (create invites, manage members)
- [ ] Push notifications showing group context
- [ ] Group avatars/images
- [ ] Leave group confirmation

### Other Future Features
- [ ] Goals creation/editing UI
- [ ] Reply to comments
- [ ] Edit post
- [ ] Delete comment
- [ ] Push notifications
- [ ] Image caching
- [ ] Offline queue for failed requests
- [ ] Deep linking
- [ ] Share sheet for posts
- [ ] Profile picture upload
- [ ] Certificate pinning (for production)
- [ ] Biometric authentication
- [ ] Widget extension
- [ ] Apple Watch app

## Environment Configuration

Update `APIClient.swift` for production:

```swift
#if DEBUG
private let baseURL = "http://localhost:8080/api/v1"
private let imageBaseURL = "http://localhost:8080"
#else
private let baseURL = "https://your-domain.com/api/v1"
private let imageBaseURL = "https://your-domain.com"
#endif
```

## Backend Configuration

The backend `.env` file should have:
```
TIMEZONE=America/New_York
```

This ensures date comparisons for challenges work correctly with the iOS app's local timezone.

## Security Notes

- Access tokens stored in memory only (cleared on app termination)
- Refresh tokens stored in Keychain with `kSecAttrAccessibleWhenUnlockedThisDeviceOnly`
- No sensitive data in UserDefaults (only last active group ID)
- ATS configured to allow localhost only in development
- Password validation enforces same rules as backend
- Images are fetched with authentication tokens
- Post deletion restricted to owner or admin
- All content API calls are scoped to user's groups (enforced by backend)

## Testing the App

1. Start the backend:
   ```bash
   cd backend && make dev
   ```

2. Build iOS targets from repo root (fast compile/link health check):
   ```bash
   make ios-build
   ```

3. Run iOS unit tests from repo root:
   ```bash
   make ios-test
   ```

4. If destination names change after an Xcode update, list valid simulator names:
   ```bash
   make ios-destinations
   ```

5. Open Xcode project and run on simulator if you want full interactive debugging/UI testing

6. Login with your admin credentials:
   - Email: justindstauffer@gmail.com
   - Password: (your password)

7. Create a post to test the flow

8. Verify:
   - Images load correctly in feed
   - Challenge progress updates after posting
   - Pull-to-refresh works on all tabs
   - Creating a post redirects to Feed with updated data
   - Reactions can be switched with single tap
   - Completing a challenge creates an Achievement post
   - Completed challenges appear in Completed tab
   - Posts can be deleted from detail view
   - Group onboarding shows for new users
   - Can create and join groups
   - Switching groups refreshes all content
   - Last active group persists across app launches

## Known Issues (Resolved)

- ~~Images may not load if backend URL is incorrect~~ - Fixed with AuthenticatedImage
- ~~Challenge progress not updating~~ - Fixed with timezone and date comparison fixes
- ~~Pull-to-refresh not working on Challenges~~ - Fixed by moving .refreshable to List
- ~~Data not syncing across tabs after post creation~~ - Fixed with AppState
- ~~Reaction emoji not disappearing when removed~~ - Fixed with optimistic updates and Equatable
- ~~Had to tap twice to change reactions~~ - Fixed with single-tap reaction switching

## Remaining Known Issues

- First launch may show briefly logged in before checking auth state
- Date formatting may vary based on device locale
- CLI `xcodebuild test` may hang in restricted/sandboxed environments during simulator attach; use local Xcode run or `make ios-build` for reliable compile validation

## Security Baseline (Feb 2026)

- Service logging is debug-gated (`#if DEBUG`) for network, websocket, and push-service internals.
- Device tokens are not printed in app logs.
- Avoid logging raw response bodies from authenticated API calls.
- Avoid logging identifiers that unnecessarily increase data exposure surface in telemetry/log sinks.

Operational invariants for future work:
- Keep sensitive operational logs behind debug guards.
- Do not add token/device-token values to logs.
- Prefer coarse-grained status logs over payload dumps in networking code.
