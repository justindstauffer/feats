# AGENTS

This file is the canonical quick-start for coding agents in this repository.

Last updated: 2026-02-16

## Context Order (Read In This Order)

1. `AGENTS.md` (this file)
2. `docs/status/now.md`
3. `context/project-map.yaml`
4. Relevant runbook in `docs/runbooks/`
5. Relevant ADR in `docs/architecture/`
6. Historical detail: `docs/status/changelog.md`

Legacy narrative docs still exist for deep history:
- `backend-progress.md`
- `ios-progress.md`

## Repo Layout

- Backend: `backend/` (Go + Gin + GORM)
- iOS: `ios/Feats/` (SwiftUI)
- Root iOS convenience commands: `Makefile`
- Backend Makefile: `backend/Makefile`

## Core Commands

Backend:
- `cd backend && go test ./...`
- `cd backend && go test -race ./...`
- `cd backend && make dev`

iOS:
- `make ios-build`
- `make ios-test`
- `make ios-destinations`

## Git Workflow

- Branch prefix: `codex/`
- Keep changes scoped and mergeable.
- Prefer small PRs with tests.
- After merge: sync `main`, delete local and remote feature branches.

## Security Guardrails

- Do not log tokens, device tokens, or raw authenticated payloads.
- Do not add query-param auth without explicit redaction strategy.
- Keep object-level authorization checks in service-layer reads.
- Keep trusted proxies explicit via `TRUSTED_PROXIES` (secure default is trust none).

## Required Updates When Behavior Changes

When you change security, auth, routing, or deployment-critical behavior:
- Update `docs/status/now.md`
- Add entry to `docs/status/changelog.md`
- Add or update ADR in `docs/architecture/` if design-level decision changed
