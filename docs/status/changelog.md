# Changelog (Engineering Context)

## 2026-02-21

- Added and stabilized major Android parity work:
  - Group onboarding/switching
  - Feed + post detail + reactions + comments
  - Create post with image upload
  - Challenges active/completed flows and challenge creation
  - Streak leaderboard
  - Profile editing, password change, and admin beta-invite management
  - Realtime refresh wiring and notification-intent navigation hooks
- Hardened Android feed/detail behavior:
  - Pagination dedupe improvements
  - Multi-image rendering in feed and post detail
  - Additional refresh signaling paths for post/reaction updates
- Confirmed Android build and test success:
  - `./gradlew :app:assembleDebug :app:testDebugUnitTest`
- Added dedicated Android progress context doc:
  - `android-progress.md`
- Updated status docs to represent current cross-platform state and deferred Android push rollout timing

## 2026-02-16

- Merged backend security hardening:
  - WebSocket subscribe authorization guard
  - Query token redaction in request logger
  - Image object-level authorization for `/images/:id`
  - Device token unregister ownership enforcement
  - Trusted proxy configuration via `TRUSTED_PROXIES`
- Added backend security regression and CI gate workflow:
  - `go test ./...`
  - `go test -race ./...`
  - `govulncheck ./...`
- Merged iOS service logging hardening:
  - Removed sensitive logging patterns
  - Debug-gated service logs for API/WebSocket/push paths
- Added structured context system for agents:
  - `AGENTS.md`
  - `docs/architecture/`
  - `docs/runbooks/`
  - `docs/status/`
  - `context/project-map.yaml`
