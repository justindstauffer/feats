# iOS App Development Progress

## Overview

This document tracks the development progress of the Feats iOS app. Use this as a reference for future Claude Code sessions.

## Tech Stack

- **Platform:** iOS 17+
- **Framework:** SwiftUI
- **State Management:** @Observable (iOS 17+)
- **Networking:** URLSession with async/await
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
    │   └── APIResponse.swift    # API response wrappers
    ├── Services/
    │   ├── KeychainService.swift  # Secure token storage
    │   ├── APIClient.swift        # HTTP client with auth
    │   ├── AuthService.swift      # Authentication state
    │   └── AppState.swift         # App-wide state management
    ├── Views/
    │   ├── MainTabView.swift      # Tab navigation
    │   ├── Auth/
    │   │   └── LoginView.swift
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

## API Client Features

- [x] Base URL configuration (debug vs release)
- [x] JSON encoding/decoding with custom date handling
- [x] Bearer token authentication
- [x] Automatic 401 handling with token clear
- [x] Multipart form upload for images
- [x] Paginated request support
- [x] Generic request methods
- [x] Authenticated image fetching (`fetchImageData`)

## Recent Changes (Feb 2026)

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
- No sensitive data in UserDefaults
- ATS configured to allow localhost only in development
- Password validation enforces same rules as backend
- Images are fetched with authentication tokens
- Post deletion restricted to owner or admin

## Testing the App

1. Start the backend:
   ```bash
   cd backend && make dev
   ```

2. Open Xcode project and run on simulator

3. Login with your admin credentials:
   - Email: justindstauffer@gmail.com
   - Password: (your password)

4. Create a post to test the flow

5. Verify:
   - Images load correctly in feed
   - Challenge progress updates after posting
   - Pull-to-refresh works on all tabs
   - Creating a post redirects to Feed with updated data
   - Reactions can be switched with single tap
   - Completing a challenge creates an Achievement post
   - Completed challenges appear in Completed tab
   - Posts can be deleted from detail view

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
