# Prompt For Codex: Build Android App Parity For Feats

You are in `/Users/jstauff/Documents/Development/feats-api`.
Build a production-ready Android app that matches the existing iOS Feats app behavior and backend integration.

## Mission
Create a complete Android app (`android/Feats`) with feature parity to the iOS app (`/Users/jstauff/Documents/Development/feats-api/ios/Feats`) and compatibility with the Go backend (`/Users/jstauff/Documents/Development/feats-api/backend`).

The output must be shippable for internal testing and TestFlight-equivalent Android beta testing (Google Play Internal/Closed testing).

## Ground Rules
- Do not break existing iOS or backend functionality.
- Keep API contracts backward-compatible.
- Reuse backend endpoints and event formats already in use.
- Prefer incremental commits by milestone.
- If you hit unknowns, inspect iOS implementation and mirror behavior.
- Write tests as you go (unit + integration where practical).

## Required Reading (First Step)
Read these files before coding:
- `/Users/jstauff/Documents/Development/feats-api/SPECIFICATION.md`
- `/Users/jstauff/Documents/Development/feats-api/ios-progress.md`
- `/Users/jstauff/Documents/Development/feats-api/docs/status/now.md`
- `/Users/jstauff/Documents/Development/feats-api/backend/cmd/api/bootstrap.go`
- `/Users/jstauff/Documents/Development/feats-api/backend/internal/handlers`
- `/Users/jstauff/Documents/Development/feats-api/ios/Feats/Feats/Services`
- `/Users/jstauff/Documents/Development/feats-api/ios/Feats/Feats/Views`

## Target Android Stack
Use modern Android defaults:
- Kotlin
- Jetpack Compose
- Navigation Compose
- ViewModel + StateFlow
- Coroutines
- Retrofit + OkHttp + Kotlinx Serialization or Moshi
- Room or DataStore only where needed (tokens should use encrypted storage)
- WebSocket client (OkHttp WebSocket)
- Firebase Cloud Messaging (FCM) for push on Android

## App Architecture Requirements
- `android/Feats` module structure should mirror iOS concepts:
  - `models`
  - `network`
  - `services`
  - `state`
  - `views` (screens/components)
- Implement strong typed API models matching backend JSON.
- Centralized API client with:
  - auth header injection
  - refresh-token flow
  - 401 handling
  - no-cache headers for feed freshness
- Secure token storage for refresh token (EncryptedSharedPreferences or equivalent).
- App-wide state object equivalent to iOS `AppState` and `GroupService`.

## Feature Parity Scope (Must Implement)
1. Authentication
- Login
- Register (with beta invite code if backend requires)
- Refresh token flow
- Logout
- Session persistence

2. Group-scoped app behavior
- Group onboarding (create/join)
- Current group selection/switching
- Persist last active group
- Group-scoped requests for feed/challenges/streaks/goals

3. Feed
- Paginated feed
- Pull-to-refresh
- Post card UI with activity icon/emoji
- Post detail screen
- Reactions add/remove/change
- Comments list/create
- Delete own/admin posts
- Image rendering for authenticated image endpoints

4. Create Post
- Activity type picker
- Image picker (up to backend limits)
- Submit post
- Return to feed + refresh

5. Challenges
- List active/completed
- Create challenge
- Join/leave
- Refresh + progress display

6. Profile
- View profile + stats
- Edit profile
- Change password
- Logout

7. Streaks/Leaderboard
- Render leaderboard per selected group

8. Real-time updates
- WebSocket connection lifecycle (foreground/background)
- Subscribe/unsubscribe by current group
- Trigger app refresh flags for relevant events

9. Push notifications (Android)
- FCM token registration to existing `/api/v1/devices` endpoint with `platform="android"`
- Notification tap deep-link behavior equivalent to iOS:
  - `post` / `comment` / `reaction` with `post_id` opens that post
  - `challenge` opens challenges tab/screen

## Backend Compatibility Work
If backend Android push delivery is missing or incomplete:
- Add safe backend support for Android push provider (FCM) while preserving APNs behavior.
- Keep device token registration backward-compatible.
- Add tests for new backend push behavior.
- Document new env vars in deployment docs.

If full Android push provider integration is too large for one pass, implement app-side registration and tap-routing first, then create a clearly scoped follow-up milestone for server-side FCM delivery.

## UX/Behavior Parity Rules
- Match iOS flow and semantics over visual pixel perfection.
- Maintain the same reaction/comment consistency expectations.
- Feed must reflect reaction changes (including type changes) without stale state.
- Notification tap must route to correct content regardless of current screen.

## Quality Bar
Before finishing, run and pass:
- Android unit tests
- Android UI/smoke tests where available
- Backend tests if backend changes were made: `cd backend && go test ./...`

Also provide:
- Build/run instructions in `android/README.md`
- Environment config instructions for local/dev/prod
- Known gaps list with severity

## Deliverables
1. New Android app code under `/Users/jstauff/Documents/Development/feats-api/android`
2. Minimal docs updates for setup, testing, and release
3. Test coverage for critical logic (auth, feed refresh behavior, reactions, notification routing)
4. A final parity checklist showing each iOS feature mapped to Android status

## Suggested Execution Plan
1. Scaffold Android app + core architecture + auth
2. Implement group state and base navigation shell
3. Implement feed + post detail + reactions/comments
4. Implement create post + image upload
5. Implement challenges, streaks, profile
6. Implement WebSocket integration
7. Implement push registration + notification deep linking
8. Add tests, polish, docs, and final parity audit

## Definition of Done
Done means:
- A tester can install Android app, log in, join/select group, create/react/comment on posts, and use challenges/profile features.
- Real-time updates and notification tap routing work.
- No known critical blockers for internal beta launch.
- Docs are sufficient for another engineer to run and ship the Android client.

## Output Format For Final Report
When finished, provide:
1. Summary of implemented features
2. Files/folders created and key design choices
3. Test results and commands
4. Remaining issues (if any) with concrete next steps

Start now.
