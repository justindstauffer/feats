# ADR-002: Security Baseline (Feb 2026)

Date: 2026-02-16
Status: Accepted

## Context

Recent hardening work addressed token exposure, websocket authorization, object-level auth, device token ownership, and proxy trust.

## Decision

Adopt and maintain the following baseline:

- WebSocket dynamic subscribe is server-authorized by group membership
- Sensitive query params are redacted in request logs
- `/images/:id` enforces object-level authorization (group member or global admin)
- Device token unregister is scoped to authenticated owner
- Trusted proxies are explicit via `TRUSTED_PROXIES`; default trust-none
- Backend PRs run `go test`, `go test -race`, and `govulncheck`

## Consequences

- Lower risk of cross-tenant data exposure
- Lower risk of credential/token leakage via logs
- Better confidence through automated regression checks

## Revisit If

- Auth model changes (e.g., mTLS between services, signed ws tickets)
- Logging/observability stack changes with different redaction model
- Deployment topology changes affecting client IP trust model
