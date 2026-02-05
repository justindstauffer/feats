#!/bin/bash
#
# Server Setup Script for Feats API
# Run this on a fresh Ubuntu 22.04 DigitalOcean droplet
#
# Usage: curl -sSL https://raw.githubusercontent.com/jstauff/feats-api/main/backend/deploy/setup-server.sh | bash
# Or: bash setup-server.sh
#

set -e

echo "=========================================="
echo "Feats API Server Setup"
echo "=========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_status() { echo -e "${GREEN}[+]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[!]${NC} $1"; }
print_error() { echo -e "${RED}[-]${NC} $1"; }

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    print_error "Please run as root (use sudo)"
    exit 1
fi

# Get domain from user
read -p "Enter your API domain (e.g., api.feats.app): " DOMAIN
if [ -z "$DOMAIN" ]; then
    print_error "Domain is required"
    exit 1
fi

read -p "Enter your email for Let's Encrypt: " EMAIL
if [ -z "$EMAIL" ]; then
    print_error "Email is required"
    exit 1
fi

print_status "Updating system packages..."
apt-get update && apt-get upgrade -y

print_status "Installing dependencies..."
apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg \
    lsb-release \
    ufw \
    fail2ban \
    nginx \
    certbot \
    python3-certbot-nginx

print_status "Installing Docker..."
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    rm get-docker.sh
fi

print_status "Installing Docker Compose..."
apt-get install -y docker-compose-plugin

print_status "Creating feats user..."
if ! id "feats" &>/dev/null; then
    useradd -m -s /bin/bash feats
    usermod -aG docker feats
fi

print_status "Creating application directories..."
mkdir -p /opt/feats-api/{data,storage,deploy}
chown -R feats:feats /opt/feats-api

print_status "Configuring firewall..."
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow 'Nginx Full'
ufw --force enable

print_status "Configuring fail2ban..."
cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 5

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3

[nginx-http-auth]
enabled = true

[nginx-limit-req]
enabled = true
EOF

systemctl enable fail2ban
systemctl restart fail2ban

print_status "Configuring Nginx..."
cat > /etc/nginx/sites-available/feats-api << EOF
# Rate limiting zones
limit_req_zone \$binary_remote_addr zone=api_limit:10m rate=10r/s;
limit_req_zone \$binary_remote_addr zone=auth_limit:10m rate=1r/s;

upstream feats_api {
    server 127.0.0.1:8080;
    keepalive 32;
}

server {
    listen 80;
    listen [::]:80;
    server_name ${DOMAIN};

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        return 301 https://\$server_name\$request_uri;
    }
}
EOF

# Enable site
ln -sf /etc/nginx/sites-available/feats-api /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default

# Create certbot webroot
mkdir -p /var/www/certbot

# Test nginx config
nginx -t
systemctl reload nginx

print_status "Obtaining SSL certificate..."
certbot --nginx -d "$DOMAIN" --non-interactive --agree-tos -m "$EMAIL" --redirect

# Update nginx with full SSL config
cat > /etc/nginx/sites-available/feats-api << EOF
# Rate limiting zones
limit_req_zone \$binary_remote_addr zone=api_limit:10m rate=10r/s;
limit_req_zone \$binary_remote_addr zone=auth_limit:10m rate=1r/s;

upstream feats_api {
    server 127.0.0.1:8080;
    keepalive 32;
}

server {
    listen 80;
    listen [::]:80;
    server_name ${DOMAIN};

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        return 301 https://\$server_name\$request_uri;
    }
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name ${DOMAIN};

    ssl_certificate /etc/letsencrypt/live/${DOMAIN}/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/${DOMAIN}/privkey.pem;

    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:50m;
    ssl_session_tickets off;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    add_header Strict-Transport-Security "max-age=63072000" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;

    access_log /var/log/nginx/feats-api.access.log;
    error_log /var/log/nginx/feats-api.error.log;

    client_max_body_size 50M;

    location /api/v1/auth/ {
        limit_req zone=auth_limit burst=5 nodelay;
        proxy_pass http://feats_api;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header Connection "";
    }

    location /api/ {
        limit_req zone=api_limit burst=20 nodelay;
        proxy_pass http://feats_api;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header Connection "";
    }

    location /health {
        proxy_pass http://feats_api;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
    }

    location /storage/ {
        limit_req zone=api_limit burst=50 nodelay;
        proxy_pass http://feats_api;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header Connection "";
        proxy_cache_valid 200 1d;
        add_header Cache-Control "public, max-age=86400";
    }
}
EOF

nginx -t && systemctl reload nginx

print_status "Setting up automatic SSL renewal..."
echo "0 0,12 * * * root certbot renew --quiet --post-hook 'systemctl reload nginx'" > /etc/cron.d/certbot-renew

print_status "Creating deployment helper script..."
cat > /opt/feats-api/deploy.sh << 'DEPLOY_EOF'
#!/bin/bash
# Deployment script - run as feats user
set -e

cd /opt/feats-api

echo "Pulling latest code..."
git pull origin main

echo "Building and starting containers..."
cd deploy
docker compose down
docker compose build --no-cache
docker compose up -d

echo "Waiting for health check..."
sleep 5
curl -f http://localhost:8080/health || echo "Health check failed!"

echo "Deployment complete!"
docker compose logs --tail=20
DEPLOY_EOF

chmod +x /opt/feats-api/deploy.sh
chown feats:feats /opt/feats-api/deploy.sh

echo ""
echo "=========================================="
print_status "Server setup complete!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Clone your repo: cd /opt/feats-api && git clone https://github.com/jstauff/feats-api.git ."
echo "2. Create .env.production: cp backend/.env.production.example backend/.env.production"
echo "3. Edit .env.production with your JWT_SECRET (generate with: openssl rand -base64 32)"
echo "4. Deploy: cd backend/deploy && docker compose up -d"
echo ""
echo "Your API will be available at: https://${DOMAIN}"
echo ""
