# Release and Merge Runbook

Last updated: 2026-02-16

## Branching

- Create feature branch: `codex/<short-topic>`
- Keep PRs focused and test-backed

## Merge Flow

1. Push branch
2. Open PR
3. Ensure CI checks pass
4. Merge PR
5. Sync local main:
   - `git checkout main`
   - `git pull --ff-only origin main`
6. Delete merged branch (local + remote)

## Post-Merge Hygiene

- Confirm `git status` clean
- Confirm no stale `codex/*` branches
