# Feats API Deployment Guide

## Prerequisites

- Docker installed on your local machine
- SSH access to the production server
- The `.env.production` file configured on the server at `/opt/feats-api/.env.production`

## Build and Deploy

### 1. Build the Docker Image (on your Mac)

```bash
cd /Users/jstauff/Documents/Development/feats-api/backend

# Build for AMD64 (server architecture)
docker build --platform linux/amd64 -t feats-api:latest .
```

### 2. Save and Transfer to Server

```bash
# Save the image to a compressed tar file
docker save feats-api:latest | gzip > feats-api.tar.gz

# Transfer to server (replace YOUR_SERVER with your server IP or hostname)
scp feats-api.tar.gz root@YOUR_SERVER:/opt/feats-api/
```

### 3. Deploy on Server

SSH into your server:
```bash
ssh root@YOUR_SERVER
```

Load the image and restart the container:
```bash
cd /opt/feats-api

# Load the new image
docker load < feats-api.tar.gz

# Stop and remove the old container
docker stop feats-api
docker rm feats-api

# Start the new container
docker run -d --name feats-api \
  -p 8080:8080 \
  -v /opt/feats-api/data:/app/data \
  -v /opt/feats-api/storage:/app/storage \
  --env-file /opt/feats-api/.env.production \
  feats-api:latest
```

### 4. Verify Deployment

Check that the container is running:
```bash
docker ps
```

Check the logs for any errors:
```bash
docker logs feats-api
```

Test the health endpoint:
```bash
curl http://localhost:8080/health
```

You should see: `{"status":"ok"}`

## Quick Deploy Script

For convenience, you can run this one-liner on the server after uploading the tar file:

```bash
cd /opt/feats-api && docker load < feats-api.tar.gz && docker stop feats-api && docker rm feats-api && docker run -d --name feats-api -p 8080:8080 -v /opt/feats-api/data:/app/data -v /opt/feats-api/storage:/app/storage --env-file /opt/feats-api/.env.production feats-api:latest && docker logs -f feats-api
```

## Troubleshooting

### View container logs
```bash
docker logs feats-api
docker logs -f feats-api  # Follow logs in real-time
```

### Check container status
```bash
docker ps -a
```

### Restart container
```bash
docker restart feats-api
```

### Check environment variables
```bash
docker inspect feats-api | grep -A 30 "Env"
```

### Enter container shell
```bash
docker exec -it feats-api /bin/sh
```

## Rollback

If you need to rollback to a previous version, keep old tar files with version names:
```bash
# Before deploying, save current version
docker save feats-api:latest | gzip > feats-api-backup-$(date +%Y%m%d).tar.gz

# To rollback, load the backup
docker load < feats-api-backup-YYYYMMDD.tar.gz
# Then restart as above
```
