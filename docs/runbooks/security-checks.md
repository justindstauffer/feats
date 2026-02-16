# Security Checks Runbook

Last updated: 2026-02-16

## Backend Required Checks

- `cd backend && go test ./...`
- `cd backend && go test -race ./...`
- `cd backend && govulncheck ./...` (installed in CI)

## Manual Spot Checks

- Confirm no token/device-token values are logged
- Confirm object-level auth paths fail closed (403/404 as expected)
- Confirm websocket subscription controls deny unauthorized subscriptions
- Confirm `TRUSTED_PROXIES` is explicit in deployment env

## If Security-Sensitive Code Changed

Update:
- `docs/status/now.md`
- `docs/status/changelog.md`
- ADR under `docs/architecture/` when policy/architecture changes
