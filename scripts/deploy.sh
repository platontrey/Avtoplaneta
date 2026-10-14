#!/usr/bin/env bash
set -e

# ==============================================================================
# Скрипт прямого защищенного деплоя Avtoplaneta на боевой сервер по SSH
# Использование: ./scripts/deploy.sh [SERVER_HOST] [SERVER_USER]
# ==============================================================================

SERVER_HOST="${1:-185.103.70.118}"
SERVER_USER="${2:-avtoplaneta}"
REMOTE_DIR="/opt/avtoplaneta"

echo "=== [1/4] Проверка подключения к серверу $SERVER_USER@$SERVER_HOST ==="
ssh -o ConnectTimeout=10 "$SERVER_USER@$SERVER_HOST" "echo 'Подключение успешно: \$(whoami)@\$(hostname)'"

echo "=== [2/4] Синхронизация файлов проекта (исключая секреты и acme.json) ==="
rsync -avz --delete \
  --exclude='.git' \
  --exclude='.env*' \
  --exclude='backend/.env*' \
  --exclude='docker-compose.override.yml' \
  --exclude='traefik/acme.json' \
  --exclude='traefik/acme-*.json' \
  --exclude='logs/' \
  --exclude='*.log' \
  --exclude='uploads/' \
  ./ "$SERVER_USER@$SERVER_HOST:$REMOTE_DIR/"

echo "=== [3/4] Сборка Docker образов на сервере с ограничением CPU ==="
ssh "$SERVER_USER@$SERVER_HOST" bash -s << 'EOF'
set -e
cd /opt/avtoplaneta

# Убеждаемся, что права на acme.json выставлены корректно
mkdir -p traefik
touch traefik/acme.json
chmod 600 traefik/acme.json

# Сборка образов с cpuset для исключения аппаратного сбоя MCE PCIe
docker build --cpuset-cpus="0-15" -t avtoplaneta-auth-service:latest ./backend -f ./backend/auth-service/Dockerfile
docker build --cpuset-cpus="0-15" -t avtoplaneta-parts-service:latest ./backend -f ./backend/parts-service/Dockerfile
docker build --cpuset-cpus="0-15" -t avtoplaneta-orders-service:latest ./backend -f ./backend/orders-service/Dockerfile
docker build --cpuset-cpus="0-15" -t avtoplaneta-messaging-service:latest ./backend -f ./backend/messaging-service/Dockerfile
docker build --cpuset-cpus="0-15" -t avtoplaneta-export-service:latest ./backend -f ./backend/export-service/Dockerfile
docker build --cpuset-cpus="0-15" -t avtoplaneta-frontend:latest ./frontend -f ./frontend/Dockerfile

echo "=== [4/4] Перезапуск контейнеров ==="
docker compose up -d --remove-orphans
docker image prune -f

echo "=== Статус запущенных сервисов ==="
docker compose ps
EOF

echo "=============================================================================="
echo " Деплой успешно завершен!"
echo "=============================================================================="
