#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# Скрипт создания нового микросервиса по стандарту Avtoplaneta
# Использование:
#   ./scripts/new-service.sh <service_name> [http_port] [grpc_port]
# Пример:
#   ./scripts/new-service.sh billing 8086 9086
# ==============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_ROOT="$(cd "${BACKEND_DIR}/.." && pwd)"

if [ $# -lt 1 ]; then
    echo "Использование: $0 <service_name> [http_port] [grpc_port]"
    echo "Пример: $0 notifications 8086 9086"
    exit 1
fi

RAW_NAME="$1"
# Если не оканчивается на -service, добавляем суффикс
if [[ "${RAW_NAME}" != *-service ]]; then
    SERVICE_NAME="${RAW_NAME}-service"
    BASE_NAME="${RAW_NAME}"
else
    SERVICE_NAME="${RAW_NAME}"
    BASE_NAME="${RAW_NAME%-service}"
fi

HTTP_PORT="${2:-8090}"
GRPC_PORT="${3:-9090}"

# Приведение к PascalCase для Go структур: billing -> Billing
PASCAL_NAME="$(echo "${BASE_NAME}" | sed -r 's/(^|-)([a-z])/\U\2/g')"
SNAKE_NAME="${SERVICE_NAME//-/_}"

TEMPLATE_DIR="${BACKEND_DIR}/template-service"
TARGET_DIR="${BACKEND_DIR}/${SERVICE_NAME}"

echo "=========================================================="
echo " Создание микросервиса: ${SERVICE_NAME}"
echo " Базовое имя:          ${BASE_NAME} (${PASCAL_NAME})"
echo " HTTP порт:             ${HTTP_PORT}"
echo " gRPC порт:             ${GRPC_PORT}"
echo " Целевая директория:    ${TARGET_DIR}"
echo "=========================================================="

if [ ! -d "${TEMPLATE_DIR}" ]; then
    echo "Ошибка: шаблон ${TEMPLATE_DIR} не найден!"
    exit 1
fi

if [ -d "${TARGET_DIR}" ]; then
    echo "Ошибка: директория ${TARGET_DIR} уже существует!"
    exit 1
fi

# 1. Копирование шаблона
echo "[1/6] Копирование эталонного шаблона..."
cp -r "${TEMPLATE_DIR}" "${TARGET_DIR}"

# 2. Замена текстовых вхождений в коде
echo "[2/6] Адаптация кода и конфигураций под ${SERVICE_NAME}..."

# Поиск и замена внутри целевой папки
find "${TARGET_DIR}" -type f -name "*.go" -o -name "*.mod" -o -name "*.yaml" -o -name "Dockerfile" -o -name "*.md" | while read -r file; do
    sed -i "s|template-service|${SERVICE_NAME}|g" "${file}"
    sed -i "s|template_service|${SNAKE_NAME}|g" "${file}"
    sed -i "s|Template|${PASCAL_NAME}|g" "${file}"
done

# Обновление дефолтных портов в config.go
sed -i "s|\"PORT\", \"8090\"|\"PORT\", \"${HTTP_PORT}\"|g" "${TARGET_DIR}/internal/config/config.go"
sed -i "s|\"GRPC_PORT\", \"9090\"|\"GRPC_PORT\", \"${GRPC_PORT}\"|g" "${TARGET_DIR}/internal/config/config.go"
sed -i "s|template_db|${SNAKE_NAME}_db|g" "${TARGET_DIR}/internal/config/config.go"

# Обновление портов в Dockerfile
sed -i "s|EXPOSE 8090 9090|EXPOSE ${HTTP_PORT} ${GRPC_PORT}|g" "${TARGET_DIR}/Dockerfile"
sed -i "s|localhost:8090/health|localhost:${HTTP_PORT}/health|g" "${TARGET_DIR}/Dockerfile"

# 3. Регистрация в go.work
echo "[3/6] Регистрация в корневом go.work..."
GOWORK_FILE="${REPO_ROOT}/go.work"
if [ -f "${GOWORK_FILE}" ]; then
    if ! grep -q "./backend/${SERVICE_NAME}" "${GOWORK_FILE}"; then
        sed -i "/\.\/backend\/tools\/vehicle-catalog-sync/i \\\t./backend/${SERVICE_NAME}" "${GOWORK_FILE}"
        echo " -> Добавлено в go.work: ./backend/${SERVICE_NAME}"
    fi
fi

# 4. Регистрация в backend/Makefile
echo "[4/6] Регистрация таргетов в backend/Makefile..."
MAKEFILE="${BACKEND_DIR}/Makefile"
if [ -f "${MAKEFILE}" ]; then
    # Добавляем в build-services
    if ! grep -q "cd ${SERVICE_NAME} && go build" "${MAKEFILE}"; then
        sed -i "/build-services:/a \\\tcd ${SERVICE_NAME} \&\& go build -o /dev/null ./..." "${MAKEFILE}"
    fi
    # Добавляем в unit-тесты
    if ! grep -q "cd ${SERVICE_NAME} && go test" "${MAKEFILE}"; then
        sed -i "/# Запуск всех юнит-тестов/a \\\tcd ${SERVICE_NAME} \&\& go test ./..." "${MAKEFILE}"
    fi
fi

# 5. Синхронизация зависимостей и генерация
echo "[5/6] Запуск go mod tidy и проверка компиляции..."
(
    cd "${TARGET_DIR}"
    go mod tidy
    sqlc generate
    go test ./...
)

echo "[6/6] Готово! Сервис успешно создан и протестирован."
echo ""
echo "=========================================================="
echo " Рекомендуемая секция для docker-compose.yml:"
echo "=========================================================="
cat <<EOF
  ${SERVICE_NAME}:
    image: avtoplaneta-${SERVICE_NAME}:latest
    build:
      context: ./backend
      dockerfile: ${SERVICE_NAME}/Dockerfile
    container_name: avtoplaneta-${BASE_NAME}
    restart: unless-stopped
    networks:
      - avtoplaneta_network
    expose:
      - "${HTTP_PORT}"
      - "${GRPC_PORT}"
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.${BASE_NAME}.rule=(Host(\`\${PUBLIC_DOMAIN:-backend-server.ru}\`) || Host(\`www.\${PUBLIC_DOMAIN:-backend-server.ru}\`)) && PathPrefix(\`/api/v1/${BASE_NAME}\`)"
      - "traefik.http.routers.${BASE_NAME}.entrypoints=websecure"
      - "traefik.http.routers.${BASE_NAME}.tls=true"
      - "traefik.http.routers.${BASE_NAME}.tls.certresolver=letsencrypt"
      - "traefik.http.routers.${BASE_NAME}.middlewares=compress-res@file,auth-forward@file"
      - "traefik.http.services.${BASE_NAME}.loadbalancer.server.port=${HTTP_PORT}"
    environment:
      - NODE_ENV=production
      - PORT=${HTTP_PORT}
      - GRPC_PORT=${GRPC_PORT}
      - DATABASE_URL=\${${BASE_NAME^^}_DATABASE_URL:-\${DATABASE_URL}}
      - REDIS_URL=redis:6379
      - GOMAXPROCS=4
    depends_on:
      - postgres
      - redis
EOF
echo "=========================================================="
