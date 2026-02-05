#!/bin/bash
#
# Deployment script for Feats API
# Run this from your local machine to deploy to the server
#
# Usage: ./deploy.sh [server-ip-or-hostname]
#

set -e

# Configuration
SERVER=${1:-"your-server-ip"}
REMOTE_USER="feats"
REMOTE_PATH="/opt/feats-api"
BRANCH="main"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_status() { echo -e "${GREEN}[+]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[!]${NC} $1"; }
print_error() { echo -e "${RED}[-]${NC} $1"; }

if [ "$SERVER" == "your-server-ip" ]; then
    print_error "Please provide server IP or hostname"
    echo "Usage: ./deploy.sh [server-ip-or-hostname]"
    exit 1
fi

print_status "Starting deployment to $SERVER..."

# Ensure we're on the right branch and up to date
print_status "Checking local git status..."
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$CURRENT_BRANCH" != "$BRANCH" ]; then
    print_warning "You're on branch '$CURRENT_BRANCH', not '$BRANCH'"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Build locally to catch errors early
print_status "Building locally to verify..."
cd "$(dirname "$0")/.."
go build -o /dev/null ./cmd/api
print_status "Local build successful"

# Deploy via SSH
print_status "Deploying to server..."
ssh -t ${REMOTE_USER}@${SERVER} << REMOTE_SCRIPT
    set -e
    cd ${REMOTE_PATH}

    echo "Pulling latest changes..."
    git fetch origin
    git reset --hard origin/${BRANCH}

    echo "Building and restarting containers..."
    cd backend/deploy
    docker compose build --no-cache
    docker compose down
    docker compose up -d

    echo "Waiting for startup..."
    sleep 5

    echo "Checking health..."
    curl -sf http://localhost:8080/health && echo " - API is healthy!" || echo " - Health check failed!"

    echo "Recent logs:"
    docker compose logs --tail=10
REMOTE_SCRIPT

print_status "Deployment complete!"
