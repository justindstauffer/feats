# Android App Development Progress

## Overview

This document tracks development progress of the Feats Android app.

Last updated: 2026-02-21

Canonical context for future agent sessions:
- `AGENTS.md`
- `docs/status/now.md`
- `docs/status/changelog.md`

## Tech Stack

- Platform: Android 10+ (API 29+)
- Language: Kotlin
- UI: Jetpack Compose + Navigation Compose
- Networking: Retrofit + OkHttp + Kotlinx Serialization
- Concurrency: Coroutines
- Realtime: OkHttp WebSocket
- Image loading: Coil
- Push client wiring: Firebase Messaging (server-side rollout deferred)

## Current Project Structure

```
android/Feats/
├── app/src/main/java/com/jstauff/feats/android/
│   ├── core/
│   │   ├── network/        # API client, DTOs, session manager
│   │   ├── realtime/       # WebSocket service
│   │   ├── state/          # App/group state stores
│   │   └── push/           # FCM service + token registrar
│   ├── ui/
│   │   ├── navigation/
│   │   ├── screens/        # auth, feed, post, challenges, leaderboard, profile, groups
│   │   └── components/
│   ├── FeatsApplication.kt
│   └── MainActivity.kt
├── app/src/test/java/com/jstauff/feats/android/
├── app/build.gradle.kts
└── settings.gradle.kts
```

## Completed Features

### Authentication and Session
- [x] Login and refresh token flow
- [x] Session bootstrap on app launch
- [x] Logout handling
- [x] Authenticated API client with bearer token header

### Group-Scoped Flow
- [x] Group onboarding (create/join)
- [x] Group switching
- [x] Group-scoped feed, post detail, challenges, streaks, profile data

### Feed and Post Detail
- [x] Feed pagination with refresh and load-more
- [x] Post detail route and data loading
- [x] Reactions add/remove
- [x] Comments list/create
- [x] Multi-image rendering in feed and post detail
- [x] Feed refresh signaling from post detail actions
- [x] Deduping during pagination merge

### Create Post
- [x] Activity type selection
- [x] Description input
- [x] Multi-photo picker
- [x] Post image upload

### Challenges
- [x] Active/completed tab filtering
- [x] Join/leave challenge
- [x] Create challenge (title, description, activity, target)
- [x] Optional start/end date input with validation

### Streaks
- [x] Group leaderboard endpoint integration
- [x] Current-user highlighting in leaderboard list

### Profile
- [x] Current user display and profile editing
- [x] Streak and goals display
- [x] Change password flow with validation
- [x] Admin beta invite management (list/create/delete)

### Realtime and Navigation
- [x] WebSocket connect/disconnect lifecycle
- [x] Group subscribe/unsubscribe handling
- [x] Refresh signals on post/reaction/comment/challenge/streak events
- [x] Notification-intent deep link routing to post/challenges

### Push (Current State)
- [x] Android device token registration client wiring
- [x] Firebase messaging service class and notification display
- [ ] End-to-end Android push delivery rollout (deferred until Firebase/Play setup is complete)

## Validation Status

- Android build and unit tests pass locally:
  - `./gradlew :app:assembleDebug :app:testDebugUnitTest`

## Known Gaps / Deferred Work

- Full Firebase/Play console production push setup and end-to-end verification
- UI visual polish pass to match SwiftUI quality more closely
- Additional instrumentation/UI tests for critical journeys

## Next Suggested Android Steps

1. Firebase/Play setup and full push validation
2. UI polish pass (typography, spacing, visual hierarchy, component consistency)
3. Expand test suite around navigation/deep-links and challenge/profile admin flows
