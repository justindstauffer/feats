# Current Status

Last updated: 2026-02-16

## Snapshot

- Branch baseline: `main` clean
- Backend and iOS security hardening wave completed and merged
- Security regression tests and CI checks are in place

## Current Priorities

1. Resume product feature delivery in small vertical slices
2. Expand backend integration coverage for group admin/comments/reactions/challenges
3. Continue iOS UX/feature work with security baseline preserved

## Active Risks

- Backend test coverage is improved but still not exhaustive across all handlers/services
- CLI iOS runtime tests can hang in restricted environments; compile checks remain reliable

## Security Baseline (Must Keep)

- WebSocket subscribe auth is server-side enforced
- Sensitive query tokens are redacted in backend logs
- `/images/:id` is object-authorized
- Device token unregister is owner-scoped
- Trusted proxies are explicit (`TRUSTED_PROXIES`, default trust none)
- PR CI runs backend tests, race, and vuln checks

## Next Suggested Feature Candidates

- Group admin UX improvements
- Goals UI completion on iOS
- Push notification deep-link handling flow
