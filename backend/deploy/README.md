# Feats API Deployment Guide

This guide covers deploying the Feats API to a DigitalOcean droplet.

## Prerequisites

- DigitalOcean account
- Domain name with DNS access
- SSH key added to DigitalOcean

## Quick Start

### 1. Create DigitalOcean Droplet

1. Log into DigitalOcean
2. Create → Droplets
3. Choose:
   - **Image:** Ubuntu 22.04 LTS
   - **Plan:** Basic, $6/mo (1 vCPU, 1GB RAM) - upgrade later if needed
   - **Datacenter:** Choose closest to your users
   - **Authentication:** SSH keys (recommended)
4. Create Droplet and note the IP address

### 2. Configure DNS

Add an A record pointing to your droplet:
```
Type: A
Name: api (or whatever subdomain you want)
Value: [your-droplet-ip]
TTL: 3600
```

Wait for DNS propagation (can check with `dig api.yourdomain.com`)

### 3. Run Server Setup Script

SSH into your server and run the setup script:

```bash
ssh root@your-droplet-ip

# Download and run setup script
curl -sSL https://raw.githubusercontent.com/jstauff/feats-api/main/backend/deploy/setup-server.sh -o setup.sh
chmod +x setup.sh
./setup.sh
```

The script will:
- Install Docker, Nginx, Certbot
- Configure firewall (UFW)
- Set up fail2ban for security
- Obtain SSL certificate from Let's Encrypt
- Create the `feats` user for running the app

### 4. Clone and Configure

```bash
# Switch to feats user
su - feats

# Clone repository
cd /opt/feats-api
git clone https://github.com/jstauff/feats-api.git .

# Create production environment file
cp backend/.env.production.example backend/.env.production

# Generate a secure JWT secret
openssl rand -base64 32

# Edit the config
nano backend/.env.production
# Change JWT_SECRET to the generated value
```

### 5. Deploy

```bash
cd /opt/feats-api/backend/deploy
docker compose up -d
```

### 6. Create Admin User

For the first deployment, you'll need to create an admin user. You can do this by:

1. Register through the app with a beta invite code
2. Then manually update the user role in the database:

```bash
# Get a shell in the container
docker compose exec api sh

# Use sqlite to update
sqlite3 /app/data/feats.db "UPDATE users SET role='admin' WHERE email='your@email.com';"
```

Or create an invite first via the API after setting up your admin.

## Deployment Commands

### Deploy Updates

From your local machine:
```bash
cd backend/deploy
./deploy.sh your-server-ip
```

Or on the server:
```bash
su - feats
cd /opt/feats-api
./deploy.sh
```

### View Logs

```bash
cd /opt/feats-api/backend/deploy
docker compose logs -f
```

### Restart Services

```bash
docker compose restart
```

### Check Status

```bash
docker compose ps
curl https://api.yourdomain.com/health
```

## File Locations

| Item | Location |
|------|----------|
| Application | `/opt/feats-api` |
| Database | Docker volume `feats-data` |
| Uploaded files | Docker volume `feats-storage` |
| Nginx config | `/etc/nginx/sites-available/feats-api` |
| SSL certificates | `/etc/letsencrypt/live/api.yourdomain.com/` |
| Nginx logs | `/var/log/nginx/feats-api.*.log` |

## Backup

### Database Backup

```bash
# Find the volume location
docker volume inspect feats-data

# Copy the database
docker compose exec api cp /app/data/feats.db /app/data/feats.db.backup
docker cp feats-api:/app/data/feats.db.backup ./feats-backup-$(date +%Y%m%d).db
```

### Automated Backups

Add to crontab (`crontab -e` as feats user):
```
0 2 * * * docker cp feats-api:/app/data/feats.db /opt/feats-api/backups/feats-$(date +\%Y\%m\%d).db
0 3 * * 0 find /opt/feats-api/backups -mtime +30 -delete
```

## Monitoring

### Health Check

The API exposes a `/health` endpoint. You can use external monitoring services like:
- UptimeRobot (free)
- Pingdom
- DigitalOcean Monitoring

### Resource Usage

```bash
docker stats feats-api
```

## Troubleshooting

### API Not Responding

1. Check if container is running: `docker compose ps`
2. Check logs: `docker compose logs --tail=50`
3. Check nginx: `sudo nginx -t && sudo systemctl status nginx`

### SSL Certificate Issues

```bash
# Renew certificate manually
sudo certbot renew --force-renewal
sudo systemctl reload nginx
```

### Database Issues

```bash
# Check database integrity
docker compose exec api sqlite3 /app/data/feats.db "PRAGMA integrity_check;"
```

### Out of Disk Space

```bash
# Check disk usage
df -h

# Clean up Docker
docker system prune -a
```

## Security Checklist

- [x] Firewall enabled (UFW)
- [x] Fail2ban configured
- [x] SSL/TLS enabled
- [x] Rate limiting configured
- [x] Non-root user for application
- [ ] SSH key-only authentication
- [ ] Regular backups configured
- [ ] Monitoring alerts set up

### Disable Password SSH (Recommended)

```bash
sudo nano /etc/ssh/sshd_config
# Set: PasswordAuthentication no
sudo systemctl restart sshd
```

## Scaling

When you need more capacity:

1. **Vertical scaling:** Resize the droplet in DigitalOcean
2. **Database:** Migrate from SQLite to PostgreSQL
3. **Storage:** Use DigitalOcean Spaces for image storage
4. **Load balancing:** Add multiple droplets behind a load balancer
