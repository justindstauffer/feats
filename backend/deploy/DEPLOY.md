# Feats API Deployment Guide

This documents how the production server is **actually** deployed: build the image
locally, transfer it, and run it with `docker run` against the existing named
volumes. This box is **not** a git checkout at `/opt/feats-api`, so the git-based
`deploy.sh` does not work here (kept only for reference / a future setup).

> ⚠️ **Read the gotchas at the bottom before your first deploy.** The database
> lives in an external named volume and the env file has quirks that have caused
> real outages.

## Server facts (prod: 45.55.34.183, Ubuntu)

- Container name: `feats-api`, runs as `appuser` (uid 1000).
- **Database:** external named volume **`feats-data`** → `/app/data/feats.db`.
  Uploads: named volume `feats-storage` → `/app/storage`.
  (There are also empty `deploy_feats-data` / `deploy_feats-storage` volumes from a
  past compose run — **never** attach those.)
- Env file: `/opt/feats-api/feats/backend/.env.production`.
- Secrets on host, mounted read-only:
  - `/home/feats/secrets/apns.p8` (iOS push)
  - `/home/feats/secrets/feats-fcm.json` (Android push)

## 1. Build the image (on your Mac)

Single-arch amd64, no attestation manifest (so `docker load` on the server is
unambiguous):

```bash
cd backend
docker buildx build --platform linux/amd64 --provenance=false --sbom=false \
  -t feats-api:latest --load .
docker image inspect feats-api:latest --format 'arch={{.Architecture}} os={{.Os}}'  # expect amd64 linux
```

## 2. Save and transfer

```bash
docker save feats-api:latest | gzip > /tmp/feats-api.tar.gz
scp /tmp/feats-api.tar.gz feats@45.55.34.183:/tmp/feats-api.tar.gz
```

## 3. Deploy (on the server)

```bash
ssh feats@45.55.34.183
docker tag feats-api:latest feats-api:rollback   # rollback point (skip if no current image)
docker load < /tmp/feats-api.tar.gz
docker stop feats-api && docker rm feats-api
docker run -d --name feats-api --restart unless-stopped \
  -p 127.0.0.1:8080:8080 \
  -v feats-data:/app/data \
  -v feats-storage:/app/storage \
  -v /home/feats/secrets/apns.p8:/app/secrets/apns.p8:ro \
  -v /home/feats/secrets/feats-fcm.json:/app/secrets/feats-fcm.json:ro \
  --env-file /opt/feats-api/feats/backend/.env.production \
  -e GIN_MODE=release -e DATABASE_PATH=/app/data/feats.db -e STORAGE_PATH=/app/storage \
  feats-api:latest
sleep 4
docker logs feats-api 2>&1 | grep -iE "notifications|FCM|APNs"
curl -sf http://localhost:8080/health && echo " HEALTH OK"
```

Expected: `APNs notifications enabled`, `FCM notifications enabled`, `HEALTH OK`.
(nginx terminates TLS and proxies to `127.0.0.1:8080`, so the container binds
localhost only.)

## Rollback

```bash
docker stop feats-api && docker rm feats-api
# rerun the docker run block above, but end with:  feats-api:rollback
```

The external volumes are untouched by rollback, so data is preserved.

## Gotchas (each of these has bitten us)

1. **`docker run --env-file` does NOT strip quotes** (docker compose does). If a
   value in `.env.production` is wrapped in quotes — e.g.
   `FCM_CREDENTIALS_PATH="/app/secrets/feats-fcm.json"` — the quotes become part
   of the value and the file "can't be found." Keep path/secret values
   **unquoted**, or override with `-e` on the `docker run` line.
2. **Use `docker run` with the named volume, not a bind mount.** `-v
   /opt/feats-api/data:/app/data` points at an empty host dir while the real DB
   sits in the `feats-data` volume — the app comes up looking like all data was
   lost. Always `-v feats-data:/app/data`.
3. **Build single-arch with `--provenance=false`.** A default buildx build makes a
   manifest list with attestations; `docker save`/`load` of that can resolve the
   wrong (or no) architecture on the server.
4. **`FCM_CREDENTIALS_PATH` must be absolute** — Go's `os.Getenv` does not expand
   `~`.

See [prod-deployment-topology] notes for the volume history.
