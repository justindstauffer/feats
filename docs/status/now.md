# Current Status

Last updated: 2026-02-21

## Snapshot

- Branch baseline: `main` clean
- Backend + iOS security hardening baseline remains in place
- Android client moved from scaffold to broad functional parity
- Android build/tests pass locally via Gradle
- Android push infrastructure is partially wired in app code, but full Firebase/Play rollout is intentionally deferred

## Platform Status

### Backend (Go API)
- Group-scoped multi-tenant API is stable and in use by both mobile clients
- Security baseline controls are implemented and should be preserved
- Push service changes for broader platform support were added during Android work; full operational rollout depends on final Firebase deployment decisions

### iOS (SwiftUI)
- Primary product surface is feature-complete for current scope
- Notifications and deep-link routing are functioning
- Serves as behavior reference for Android parity decisions

### Android (Kotlin + Compose)
- Implemented:
  - Auth/session bootstrap
  - Group onboarding/switching
  - Feed + post detail + reactions + comments
  - Create post with photo upload
  - Challenges list/join/leave/create (with optional date inputs)
  - Streak leaderboard
  - Profile edit, change password, admin beta invites
  - WebSocket refresh signaling and notification-intent navigation hooks
- Build state:
  - `./gradlew :app:assembleDebug :app:testDebugUnitTest` passes locally

## Current Priorities

1. Preserve stability while converting Android functional parity into polished production UX
2. Keep backend/iOS behavior aligned with Android as parity evolves
3. Complete Android push rollout later (Firebase + Play setup window)

## Active Risks

- Android visual polish and UX consistency lag behind iOS despite functional parity
- Android push end-to-end is not production-validated yet (deferred intentionally)
- Backend test coverage is solid in core areas but still not exhaustive for all edge-path handlers

## Security Baseline (Must Keep)

- WebSocket subscribe auth is server-side enforced
- Sensitive query tokens are redacted in backend logs
- `/images/:id` is object-authorized
- Device token unregister is owner-scoped
- Trusted proxies are explicit (`TRUSTED_PROXIES`, default trust none)
- PR CI runs backend tests, race, and vuln checks

## Primary Context Files

- `backend-progress.md`
- `ios-progress.md`
- `android-progress.md`
- `docs/status/changelog.md`
