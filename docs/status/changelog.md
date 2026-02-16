# Changelog (Engineering Context)

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
