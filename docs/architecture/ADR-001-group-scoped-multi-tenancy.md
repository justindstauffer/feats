# ADR-001: Group-Scoped Multi-Tenancy

Date: 2026-02-16
Status: Accepted

## Context

The product supports multiple user groups where posts, challenges, goals, streaks, and custom activities must be isolated by group.

## Decision

All content-domain routes and data access are group-scoped.

- API route pattern: `/api/v1/groups/:gid/...`
- Membership middleware enforces access (`member` or `admin`)
- Admin middleware enforces elevated actions for group management
- Services validate group relationships for domain actions

## Consequences

- Stronger tenant isolation by default
- More explicit route and service signatures
- Slightly higher implementation overhead for new endpoints

## Revisit If

- Product moves to organization/workspace hierarchy beyond one group dimension
- Need for cross-group federation views with explicit policy controls
