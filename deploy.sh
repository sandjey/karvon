#!/usr/bin/env bash
set -euo pipefail

# Prevent two simultaneous deploys (CI + manual overlap causes "removal already in progress")
exec 9>/tmp/ctm_deploy.lock
flock -w 600 9 || { echo "[ctm] ERROR: deploy lock timeout вЂ” another deploy has been running > 10 min"; exit 1; }

APP_DIR=/home/dev_backend/karvon

cd $APP_DIR

echo "[ctm] Pulling latest code..."
git fetch origin master
git reset --hard FETCH_HEAD

echo "[ctm] Building and restarting containers..."
COMPOSE="docker-compose"
docker compose version &>/dev/null 2>&1 && COMPOSE="docker compose"

# Ensure postgres and whatsapp are running (start if not, no recreate)
$COMPOSE up -d --no-recreate postgres whatsapp 2>/dev/null || true

# Build new images
$COMPOSE build app admin

# Let compose manage the full lifecycle вЂ” it stops, removes, and starts containers itself.
# Do NOT manually docker stop/rm before this: compose tracks containers by ID in image
# labels, and removing them externally causes "No such container" on recreate.
$COMPOSE up -d --no-deps --force-recreate app admin

echo "[ctm] Waiting for app to be healthy..."
sleep 15
if curl -sf http://127.0.0.1:8082/health > /dev/null 2>&1; then
  echo "[ctm] OK вЂ” app is running at :8082"
else
  echo "[ctm] WARNING вЂ” health check failed, check logs:"
  $COMPOSE logs app --tail=30
fi

if curl -sf http://127.0.0.1:8085/ > /dev/null 2>&1; then
  echo "[ctm] OK вЂ” admin panel is running at :8085"
else
  echo "[ctm] WARNING вЂ” admin panel health check failed, check logs:"
  $COMPOSE logs admin --tail=20
fi
