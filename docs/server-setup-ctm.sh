#!/usr/bin/env bash
# Central Trade Market — server setup script
# Run as: dev_backend on 46.8.176.2
set -euo pipefail

echo "=== [CTM] Server setup: api.centraltrademarket.com ==="

ENV=/home/dev_backend/karvon/.env

# ─── 1. Update .env (only changed values) ────────────────────────────────────
echo "[1/5] Updating .env..."

# Add/update REDIS_URL
grep -q "^REDIS_URL=" "$ENV" \
  && sed -i 's|^REDIS_URL=.*|REDIS_URL=redis://redis:6379|' "$ENV" \
  || echo "REDIS_URL=redis://redis:6379" >> "$ENV"

# Remove duplicate PUBLIC_URL and set correct one
sed -i '/^PUBLIC_URL=/d' "$ENV"
echo "PUBLIC_URL=https://api.centraltrademarket.com" >> "$ENV"

# Update callback URLs
sed -i 's|MULTICARD_CALLBACK_URL=.*|MULTICARD_CALLBACK_URL=https://api.centraltrademarket.com/api/v1/payments/webhook|' "$ENV"
sed -i 's|MULTICARD_RETURN_URL=.*|MULTICARD_RETURN_URL=https://api.centraltrademarket.com/payment/result|' "$ENV"

echo "  .env updated (REDIS_URL, PUBLIC_URL, callbacks)"

# ─── 2. Create nginx HTTP config (for certbot challenge) ─────────────────────
echo "[2/5] Creating nginx HTTP config..."
sudo tee /etc/nginx/sites-available/api.centraltrademarket.com > /dev/null << 'NGEOF'
server {
    listen 80;
    server_name api.centraltrademarket.com;

    location /.well-known/acme-challenge/ {
        root /var/www/html;
    }

    location / {
        proxy_pass http://127.0.0.1:8082;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
NGEOF

sudo ln -sf /etc/nginx/sites-available/api.centraltrademarket.com \
            /etc/nginx/sites-enabled/api.centraltrademarket.com
sudo nginx -t && sudo systemctl reload nginx
echo "  Nginx HTTP config enabled"

# ─── 3. Get SSL certificate ──────────────────────────────────────────────────
echo "[3/5] Getting SSL certificate..."
echo "  DNS CHECK: api.centraltrademarket.com must point to $(curl -s ifconfig.me 2>/dev/null || echo 46.8.176.2)"
echo ""

# Check DNS resolves to this server
SERVER_IP=$(curl -s ifconfig.me 2>/dev/null || echo "")
DOMAIN_IP=$(dig +short api.centraltrademarket.com 2>/dev/null | tail -1 || true)
if [ -n "$SERVER_IP" ] && [ "$SERVER_IP" != "$DOMAIN_IP" ]; then
    echo "  ERROR: Domain points to $DOMAIN_IP, but this server is $SERVER_IP"
    echo "  Please set A record: api.centraltrademarket.com -> $SERVER_IP"
    echo "  Then re-run this script."
    exit 1
fi

sudo certbot --nginx -d api.centraltrademarket.com \
     --non-interactive --agree-tos \
     --email matyoqub18@gmail.com \
     --redirect
echo "  SSL certificate obtained!"

# ─── 4. Full nginx SSL config ─────────────────────────────────────────────────
echo "[4/5] Updating nginx full SSL config..."
sudo tee /etc/nginx/sites-available/api.centraltrademarket.com > /dev/null << 'NGEOF'
server {
    listen 80;
    server_name api.centraltrademarket.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    server_name api.centraltrademarket.com;

    ssl_certificate /etc/letsencrypt/live/api.centraltrademarket.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.centraltrademarket.com/privkey.pem;
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

    client_max_body_size 20M;

    add_header Access-Control-Allow-Origin  "*" always;
    add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, PATCH, OPTIONS" always;
    add_header Access-Control-Allow-Headers "Authorization, Content-Type, Accept-Language" always;

    location /uploads/ {
        alias /home/dev_backend/karvon/uploads/;
        expires 30d;
    }

    location /admin/ {
        proxy_pass         http://127.0.0.1:8085/;
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-Proto $scheme;
    }

    location /docs/ {
        alias /home/dev_backend/karvon/docs/;
        autoindex off;
        add_header Cache-Control "no-cache";
    }

    location / {
        if ($request_method = OPTIONS) {
            add_header Access-Control-Allow-Origin  "*";
            add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, PATCH, OPTIONS";
            add_header Access-Control-Allow-Headers "Authorization, Content-Type, Accept-Language";
            add_header Content-Length 0;
            return 204;
        }
        proxy_pass         http://127.0.0.1:8082;
        proxy_http_version 1.1;
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
        proxy_read_timeout 120s;
    }
}
NGEOF

sudo nginx -t && sudo systemctl reload nginx
echo "  Full SSL nginx config applied"

# ─── 5. Redeploy app with Redis ──────────────────────────────────────────────
echo "[5/5] Redeploying with Redis..."
cd /home/dev_backend/karvon

COMPOSE="docker-compose"
docker compose version &>/dev/null 2>&1 && COMPOSE="docker compose"

$COMPOSE up -d --no-recreate redis
$COMPOSE build app
$COMPOSE up -d --no-deps --force-recreate app

sleep 12
if curl -sf http://127.0.0.1:8082/health > /dev/null 2>&1; then
    echo "  App is healthy"
else
    echo "  WARNING: health check failed — check logs:"
    $COMPOSE logs app --tail=30
fi

echo ""
echo "=== [CTM] Setup complete! ==="
echo "  API:    https://api.centraltrademarket.com/api/v1"
echo "  Admin:  https://api.centraltrademarket.com/admin/"
echo "  Health: curl https://api.centraltrademarket.com/health"
echo ""
echo "  OTP test:"
echo "  curl -X POST https://api.centraltrademarket.com/api/v1/auth/send-otp \\"
echo "       -H 'Content-Type: application/json' \\"
echo "       -d '{\"phone\":\"+998331108811\",\"channel\":\"telegram\"}'"
