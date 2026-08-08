# Feats API Deployment Guide

> ⚠️ **Use `deploy.sh` / docker-compose — do NOT `docker run` with bind mounts.**
> The production database lives in the **external named volume `feats-data`**
> (see `docker-compose.yml`), not in `/opt/feats-api/data`. An earlier version of
> this guide used `docker run -v /opt/feats-api/data:/app/data`, which points the
> app at an *empty* directory while the real data sits untouched in the named
> volume — the app comes up looking like all data was lost. Always deploy through
> `docker compose` so the correct volumes are attached.

## Prerequisites

- SSH access to the production server (user `feats`)
- `/opt/feats-api/.env.production` configured on the server
- Secrets present on the host (mounted read-only by `docker-compose.yml`):
  - `/home/feats/secrets/apns.p8` — APNs key (iOS push)
  - `/home/feats/secrets/feats-fcm.json` — Firebase service-account key (Android push)

## Standard deploy (from your Mac)

```bash
cd backend/deploy
./deploy.sh <server-ip-or-hostname>
```

This pushes nothing itself — it SSHes in, runs `git reset --hard origin/main`,
then `docker compose build --no-cache && down && up`. Because `feats-data` and
`feats-storage` are `external: true`, `down`/`up` never removes them, so the
database and uploads are preserved.

Make sure your work is pushed to `origin/main` first (`git push origin main`).

## Enabling Android push (FCM) — one-time server setup

The FCM migration needs the service-account key on the host and an env var:

1. Copy the Firebase service-account JSON to the server:
   ```bash
   scp feats-fcm.json feats@<server>:/home/feats/secrets/feats-fcm.json
   ```
2. Add this line to `/opt/feats-api/.env.production` on the server
   (**absolute, in-container path** — Go does not expand `~`):
   ```
   FCM_CREDENTIALS_PATH=/app/secrets/feats-fcm.json
   ```
3. Deploy as above. On startup the logs should show `FCM notifications enabled`.

## Verify

```bash
ssh feats@<server> 'cd /opt/feats-api/backend/deploy && docker compose logs --tail=30'
curl -sf https://feats-api.jstauff.com/health
```

Look for `APNs notifications enabled` and `FCM notifications enabled`.

## Rollback

`deploy.sh` uses `git reset --hard origin/main`, so to roll back, move `main`
back (revert the bad commit and push) and re-run `deploy.sh`. The external
volumes keep the data across rollbacks.
