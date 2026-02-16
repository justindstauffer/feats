# Backend Development Progress

## Overview

This document tracks the development progress of the Feats API backend. Use this as a reference for future coding-agent sessions.

Note: canonical day-to-day agent context now lives in:
- `AGENTS.md`
- `docs/status/now.md`
- `context/project-map.yaml`

## Tech Stack

- **Language:** Go 1.21+
- **Framework:** Gin (HTTP router)
- **ORM:** GORM with SQLite3
- **Authentication:** bcrypt + JWT
- **IDs:** UUID v4

## Project Structure

```
backend/
├── cmd/
│   ├── api/
│   │   ├── main.go              # Thin API entry point
│   │   ├── bootstrap.go         # Server/service wiring and router setup
│   │   └── api_integration_test.go # API integration tests
│   └── admin/
│       └── main.go              # CLI tool to create admin user
├── internal/
│   ├── config/
│   │   └── config.go            # Environment configuration
│   ├── database/
│   │   └── database.go          # DB connection, migrations, seeding
│   ├── handlers/
│   │   ├── auth.go              # Login, logout, password reset
│   │   ├── post_workflow.go     # Post side-effect workflow orchestration
│   │   ├── user.go              # User CRUD, admin functions
│   │   ├── group.go             # Group management, invites, membership
│   │   ├── post.go              # Posts and image uploads
│   │   ├── activity.go          # Activity types
│   │   ├── reaction.go          # Post reactions
│   │   ├── comment.go           # Threaded comments
│   │   ├── streak.go            # Streak tracking
│   │   ├── challenge.go         # Challenges
│   │   └── goal.go              # Personal goals
│   ├── middleware/
│   │   ├── auth.go              # JWT validation, role checking
│   │   ├── group.go             # Group membership/admin validation
│   │   ├── security.go          # Security headers
│   │   ├── ratelimit.go         # Token bucket rate limiting
│   │   ├── logger.go            # Request logging
│   │   └── cors.go              # CORS configuration
│   ├── models/
│   │   ├── user.go              # User model with security fields
│   │   ├── auth.go              # RefreshToken, PasswordHistory, ResetToken
│   │   ├── group.go             # Group, GroupMember, GroupInvite
│   │   ├── activity.go          # ActivityType with core types
│   │   ├── post.go              # Post and PostImage
│   │   ├── reaction.go          # Reaction with 5 types
│   │   ├── comment.go           # Threaded comments
│   │   ├── streak.go            # Streak tracking logic
│   │   ├── challenge.go         # Challenge and ChallengeParticipant
│   │   ├── goal.go              # Goal with daily/weekly/monthly periods
│   │   ├── device.go            # DeviceToken for push notifications
│   │   ├── audit.go             # AuditLog and RateLimit
│   │   └── models.go            # Response types and helpers
│   ├── services/
│   │   ├── auth.go              # Authentication logic, JWT, password validation
│   │   ├── beta_invite.go       # Beta invite generation/validation
│   │   ├── user.go              # User management
│   │   ├── group.go             # Group CRUD, membership, invites
│   │   ├── audit.go             # Security event logging
│   │   ├── activity.go          # Activity type management
│   │   ├── post.go              # Post CRUD, image processing
│   │   ├── reaction.go          # Reaction management
│   │   ├── comment.go           # Comment CRUD
│   │   ├── streak.go            # Streak calculation and updates
│   │   ├── challenge.go         # Challenge management
│   │   ├── goal.go              # Goal management
│   │   └── push.go              # Push notification dispatch/service layer
│   └── storage/                 # (Placeholder for S3 abstraction)
├── migrations/                  # (Placeholder for manual migrations)
├── .env.example                 # Environment variable template
├── .gitignore
├── Makefile                     # Build and run commands
├── go.mod
└── go.sum
```

## Completed Features

### Authentication & Security
- [x] JWT access tokens (15 min TTL)
- [x] Refresh token rotation with hash storage
- [x] Password hashing with bcrypt (cost 12)
- [x] Password complexity validation (12+ chars, upper, lower, digit, special)
- [x] Password history (prevents reuse of last 5)
- [x] Account lockout after 5 failed attempts
- [x] Password reset token generation
- [x] Security headers (HSTS, CSP, X-Frame-Options, etc.)
- [x] Rate limiting (login, API, uploads)
- [x] Audit logging for security events

### User Management
- [x] Admin and User roles
- [x] Admin can create/delete users
- [x] Profile updates (name, bio, profile picture path)
- [x] Force password change on first login (for admin-created users)

### Activity Types
- [x] 7 core activity types seeded (Gym, Hiking, Golf, Walking, Running, Cycling, Swimming)
- [x] Custom activity creation by users
- [x] Delete protection for system types and in-use types

### Posts (Feats)
- [x] Create posts with activity type and description
- [x] Edit and soft-delete posts
- [x] Image upload (up to 4 per post)
- [x] Image re-encoding to JPEG (security measure)
- [x] Image serving endpoint with auth
- [x] Pagination for post listing

### Reactions
- [x] 5 reaction types (👍 👍❤️ 🔥 💪 👏)
- [x] One reaction per user per post
- [x] Reaction summary with counts

### Comments
- [x] Threaded comments with replies
- [x] Edit and soft-delete comments
- [x] Max length validation (1000 chars)
- [x] HTML stripping

### Streaks
- [x] Automatic streak tracking on post creation
- [x] Streak reset on missed days
- [x] Longest streak tracking
- [x] Leaderboard endpoint

### Challenges
- [x] Create challenges (open or time-bound)
- [x] Optional activity type filter
- [x] Join/leave challenges
- [x] Automatic progress tracking on post creation
- [x] Completion detection

### Goals
- [x] Personal goals with daily/weekly/monthly periods
- [x] Optional activity type filter
- [x] Automatic progress tracking
- [x] Period reset logic

### Multi-Tenancy / Groups (Feb 2026)
- [x] Group model with name and description
- [x] Group membership with admin/member roles
- [x] Invite code system (XXXX-XXXX-XXXX format)
- [x] Invite expiration and max uses
- [x] Rate limiting on invite redemption (brute-force protection)
- [x] Users can belong to multiple groups
- [x] Posts, challenges, goals, streaks are group-scoped
- [x] Custom activity types are group-scoped
- [x] Leaderboard is per-group
- [x] Soft-delete membership (preserves historical posts)
- [x] Group admin can manage members and invites

### Stability + Refactor Pass (Feb 2026)
- [x] Extracted API bootstrap/wiring from `cmd/api/main.go` to `cmd/api/bootstrap.go`
- [x] Extracted post side-effect orchestration into `internal/handlers/post_workflow.go`
- [x] Added integration tests for auth invite flow and post->streak behavior (`cmd/api/api_integration_test.go`)
- [x] Added service tests for invite parsing and beta invite usage rules (`internal/services/group_test.go`, `internal/services/beta_invite_test.go`)
- [x] Fixed malformed invite-code panic risk in group invite redemption flow
- [x] Added registration rollback behavior so failed registration un-consumes beta invites
- [x] Fixed push service status-code format bug
- [x] `go test ./...` passes
- [x] `go test -race ./...` passes

## API Endpoints

### Public
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/login` | Login |
| POST | `/api/v1/auth/refresh` | Refresh tokens |
| POST | `/api/v1/auth/password/reset-request` | Request password reset |
| POST | `/api/v1/auth/password/reset` | Reset password |
| GET | `/health` | Health check |

### Protected (requires Bearer token)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/logout` | Logout |
| POST | `/api/v1/auth/password/change` | Change password |
| GET | `/api/v1/users/me` | Get current user |
| PUT | `/api/v1/users/me` | Update current user |
| GET | `/api/v1/users/:id` | Get user |
| POST | `/api/v1/devices` | Register device token |
| DELETE | `/api/v1/devices` | Unregister device token (owner-scoped) |
| POST | `/api/v1/invites/redeem` | Redeem invite code |
| GET | `/images/:id` | Serve image |

### Group Management (requires Bearer token)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/groups` | Create group |
| GET | `/api/v1/groups` | List user's groups |

### Group-Scoped Routes (requires group membership)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/groups/:gid` | Get group |
| POST | `/api/v1/groups/:gid/leave` | Leave group |
| GET | `/api/v1/groups/:gid/members` | List members |
| GET | `/api/v1/groups/:gid/users/:id/streak` | Get user streak |
| GET | `/api/v1/groups/:gid/users/:id/goals` | Get user goals |
| GET | `/api/v1/groups/:gid/activities` | List activities |
| POST | `/api/v1/groups/:gid/activities` | Create activity |
| DELETE | `/api/v1/groups/:gid/activities/:id` | Delete activity |
| GET | `/api/v1/groups/:gid/posts` | List posts |
| POST | `/api/v1/groups/:gid/posts` | Create post |
| GET | `/api/v1/groups/:gid/posts/:id` | Get post |
| PUT | `/api/v1/groups/:gid/posts/:id` | Update post |
| DELETE | `/api/v1/groups/:gid/posts/:id` | Delete post |
| POST | `/api/v1/groups/:gid/posts/:id/images` | Upload image |
| DELETE | `/api/v1/groups/:gid/posts/:id/images/:image_id` | Delete image |
| GET | `/api/v1/groups/:gid/posts/:id/reactions` | Get reactions |
| POST | `/api/v1/groups/:gid/posts/:id/reactions` | Add reaction |
| DELETE | `/api/v1/groups/:gid/posts/:id/reactions` | Remove reaction |
| GET | `/api/v1/groups/:gid/posts/:id/comments` | Get comments |
| POST | `/api/v1/groups/:gid/posts/:id/comments` | Create comment |
| PUT | `/api/v1/groups/:gid/comments/:id` | Update comment |
| DELETE | `/api/v1/groups/:gid/comments/:id` | Delete comment |
| GET | `/api/v1/groups/:gid/streaks/leaderboard` | Get leaderboard |
| GET | `/api/v1/groups/:gid/challenges` | List challenges |
| POST | `/api/v1/groups/:gid/challenges` | Create challenge |
| GET | `/api/v1/groups/:gid/challenges/:id` | Get challenge |
| POST | `/api/v1/groups/:gid/challenges/:id/join` | Join challenge |
| DELETE | `/api/v1/groups/:gid/challenges/:id/leave` | Leave challenge |
| DELETE | `/api/v1/groups/:gid/challenges/:id` | Delete challenge |
| POST | `/api/v1/groups/:gid/goals` | Create goal |
| PUT | `/api/v1/groups/:gid/goals/:id` | Update goal |
| DELETE | `/api/v1/groups/:gid/goals/:id` | Delete goal |

### Group Admin Routes (requires group admin role)
| Method | Endpoint | Description |
|--------|----------|-------------|
| PUT | `/api/v1/groups/:gid` | Update group |
| DELETE | `/api/v1/groups/:gid` | Delete group |
| PUT | `/api/v1/groups/:gid/members/:uid` | Update member role |
| DELETE | `/api/v1/groups/:gid/members/:uid` | Remove member |
| POST | `/api/v1/groups/:gid/invites` | Create invite |
| GET | `/api/v1/groups/:gid/invites` | List invites |
| DELETE | `/api/v1/groups/:gid/invites/:iid` | Revoke invite |

### Admin Only
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/admin/users` | Create user |
| GET | `/api/v1/admin/users` | List users |
| DELETE | `/api/v1/admin/users/:id` | Delete user |
| GET | `/api/v1/admin/audit-logs` | Get audit logs |

## Getting Started

```bash
cd backend

# Install dependencies
make deps

# Copy environment template
cp .env.example .env

# Edit .env and set JWT_SECRET (required, min 32 chars)
# Generate one with: make generate-secret

# Create the initial admin user
make create-admin

# Run in development mode
make dev
```

Server runs at `http://localhost:8080`

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make build` | Build binary |
| `make run` | Build and run |
| `make dev` | Run in debug mode |
| `make deps` | Install dependencies |
| `make test` | Run tests |
| `make create-admin` | Create admin user interactively |
| `make db-reset` | Delete database file |
| `make generate-secret` | Generate a secure JWT secret |

Root repo Makefile commands:

| Command | Description |
|---------|-------------|
| `make ios-build` | Build iOS app + test targets (`build-for-testing`) |
| `make ios-test` | Run iOS unit tests target (`FeatsTests`) |
| `make ios-destinations` | Show available Xcode destinations |

## TODO / Not Yet Implemented

### Backend
- [ ] Push notifications (APNs integration)
- [ ] Profile picture upload endpoint
- [ ] Image compression for large uploads
- [ ] S3 storage backend
- [ ] Email service integration (optional)
- [ ] Background job for hard-deleting old soft-deleted content
- [ ] Background job for cleaning old audit logs
- [x] Unit tests (core service coverage started)
- [x] Integration tests (API integration coverage started)
- [ ] Expand unit test coverage across all services/handlers
- [ ] Expand integration coverage for group admin, comments, reactions, and challenges

### Future Security Enhancements (Low Priority)
- [ ] JWT issuer/audience validation - Add `iss` and `aud` claims for defense-in-depth
- [ ] JTI blocklist for immediate token revocation - Currently access tokens valid until expiration even after password change
- [ ] Redis-based distributed rate limiting - Current in-memory rate limiting doesn't work across multiple instances
- [ ] Request body size limits middleware
- [ ] CSRF tokens (if browser-facing features added)

### iOS App
- [ ] Not started - see `ios/` folder

## Environment Variables

See `.env.example` for full list. Key variables:

| Variable | Required | Description |
|----------|----------|-------------|
| `JWT_SECRET` | Yes | Min 32 chars, used for signing JWTs |
| `DATABASE_PATH` | No | Default: `./feats.db` |
| `PORT` | No | Default: `8080` |
| `GIN_MODE` | No | `debug` or `release` |
| `BCRYPT_COST` | No | Default: `12` |
| `ALLOWED_ORIGINS` | No | Comma-separated list of allowed CORS origins (empty = deny all) |
| `SESSION_INACTIVE_TTL` | No | Default: `720h` (30 days) - Session inactivity timeout |

## Database

SQLite database auto-created on first run. Tables are auto-migrated via GORM.

Core activity types are seeded automatically:
- Gym 🏋️
- Hiking 🥾
- Golf ⛳
- Walking 🚶
- Running 🏃
- Cycling 🚴
- Swimming 🏊

## Security Notes

- All passwords hashed with bcrypt (cost 12)
- JWT access tokens are short-lived (15 min)
- Refresh tokens stored as SHA-256 hashes
- Refresh token rotation on each use
- Account lockout after 5 failed login attempts
- All security events logged to audit_logs table
- Images re-encoded to JPEG to strip metadata
- Rate limiting on all endpoints
- Security headers on all responses

### Security Fixes Implemented (Feb 2026)

| Fix | Description |
|-----|-------------|
| CORS Configuration | Added `ALLOWED_ORIGINS` env var; denies all browser origins by default |
| Path Traversal Protection | `ServeImage` validates paths stay within storage directory |
| Magic Byte Validation | Image uploads validated by file signature, not just extension |
| Session Inactivity Timeout | Sessions expire after `SESSION_INACTIVE_TTL` of inactivity |
| Audit Log Sanitization | Fixed `maskEmail` panic, added input sanitization for log entries |
| HTML Sanitization | User input escaped with `html.EscapeString` after tag stripping |
| UserHandler Dependency | Fixed nil `AuthService` in user creation flow |

### Multi-Tenancy Security (Feb 2026)

| Feature | Description |
|---------|-------------|
| Group Isolation | All content scoped to groups; cross-group access prevented |
| Reaction/Comment Group Validation | Services validate post belongs to group before allowing reactions/comments |
| Invite Code Brute-Force Protection | 12-char codes (A-Z, 2-9 charset), 5 attempts/min rate limit |
| Invite Expiration | Codes expire after configurable time |
| Invite Use Limits | Max uses per invite code |
| Membership Validation | Middleware checks group membership/admin for all group routes |
| Soft Delete Membership | Historical posts preserved when users leave |

### Security Baseline (Feb 2026)

| Area | Baseline |
|------|----------|
| WebSocket subscriptions | Dynamic `subscribe` requests are server-authorized against active group membership |
| Token logging safety | Sensitive query params (`token`, `access_token`, `refresh_token`, `authorization`) are redacted in request logs |
| Image object authorization | `/images/:id` requires object-level authorization (group member or global admin) before file serve |
| Device token ownership | Device token unregister is scoped to authenticated owner (`user_id + token`) |
| Proxy/IP trust | Gin trusted proxies are explicitly configured via `TRUSTED_PROXIES`; default is trust none |
| Security regression guardrails | CI runs `go test ./...`, `go test -race ./...`, and `govulncheck ./...` on backend PRs |

Operational invariants for future work:
- Never introduce new auth tokens in URL query params without explicit redaction/masking strategy.
- Keep object-level authorization checks in service-layer read paths (not only at route-level middleware).
- Preserve explicit trusted-proxy configuration; do not rely on framework defaults for client IP trust.
- Update security regression tests whenever auth/authz behavior changes.

## Reference Documents

- `SPECIFICATION.md` - Full project specification with security requirements
