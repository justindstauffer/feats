# Android App (Feats)

This folder contains the Android client for Feats.

## Current Status
- Functional parity is largely in place with core iOS flows.
- Build/test currently pass locally via Gradle.
- Push rollout is intentionally deferred until Firebase + Play setup is finalized.

Implemented areas:
- Auth/session bootstrap and login
- Group onboarding and group switching
- Feed, post detail, reactions, comments
- Create post with image upload
- Challenges list/join/leave/create
- Streak leaderboard
- Profile editing, password change, admin beta invites
- WebSocket refresh signaling
- Notification-intent deep-link routing hooks

## Setup
1. Open `android/Feats` in Android Studio (latest stable).
2. Let Gradle sync.
3. Ensure JDK 17 is selected for Gradle.
4. Configure API endpoints in `app/build.gradle.kts` (`API_BASE_URL`, `WS_BASE_URL`) for your environment.
5. Run app on emulator/device with Android 10+ (API 29+).

## Build and Test

From `android/Feats`:

```bash
./gradlew :app:assembleDebug :app:testDebugUnitTest
```

## Notes

- Push notifications:
  - App-side FCM wiring and device registration code exists.
  - End-to-end production rollout is deferred until Firebase and Google Play console setup is complete.
- Use `android-progress.md` and `docs/status/now.md` for current cross-platform context.
