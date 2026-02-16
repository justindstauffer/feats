# Development Runbook

Last updated: 2026-02-16

## Local Backend

1. `cd backend`
2. `make deps`
3. `cp .env.example .env` (first time)
4. Set `JWT_SECRET` in `.env`
5. `make dev`

## Local iOS

From repo root:

1. `make ios-build`
2. `make ios-test` (unit target)
3. `make ios-destinations` if simulator name changed

## Verification Before PR

- `cd backend && go test ./...`
- `cd backend && go test -race ./...`
- `make ios-build` (if iOS files changed)
