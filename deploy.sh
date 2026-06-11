#!/usr/bin/env bash
set -euo pipefail
APP_DIR=/home/dev_backend/karvon

cd $APP_DIR

echo "[karvon] Pulling latest code..."
git pull --ff-only origin master

echo "[karvon] Building and restarting containers..."
# Prefer docker compose v2 over legacy docker-compose v1
COMPOSE="docker-compose"
docker compose version &>/dev/null 2>&1 && COMPOSE="docker compose"

# Clean up stale stopped containers from previous deploys
docker container prune -f --filter "label=com.docker.compose.project=karvon" 2>/dev/null || true

# Ensure postgres and whatsapp are running (start if not, no recreate)
$COMPOSE up -d --no-recreate postgres whatsapp 2>/dev/null || true

# Build new images (doesn't stop anything)
$COMPOSE build app admin

# Restart app and admin containers
$COMPOSE up -d --no-deps app admin

echo "[karvon] Waiting for app to be healthy..."
sleep 10
if curl -sf http://127.0.0.1:8082/health > /dev/null 2>&1; then
  echo "[karvon] OK — app is running at :8082"
else
  echo "[karvon] WARNING — health check failed, check logs:"
  $COMPOSE logs app --tail=30
fi

if curl -sf http://127.0.0.1:8085/ > /dev/null 2>&1; then
  echo "[karvon] OK — admin panel is running at :8085"
else
  echo "[karvon] WARNING — admin panel health check failed, check logs:"
  $COMPOSE logs admin --tail=20
fi
