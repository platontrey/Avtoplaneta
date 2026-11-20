#!/bin/bash
set -euo pipefail

# Скрипт создания нового микросервиса из эталонного шаблона template-service
# Использование:
#   make new-service NAME=billing PORT=8086 GRPC_PORT=9086

NAME="${1:-}"
PORT="${2:-}"
GRPC_PORT="${3:-}"

if [ -z "${NAME}" ] || [ -z "${PORT}" ] || [ -z "${GRPC_PORT}" ]; then
    echo "Ошибка: Не указаны обязательные аргументы."
    echo "Использование: make new-service NAME=<имя> PORT=<http_порт> GRPC_PORT=<grpc_порт>"
    echo "Пример: make new-service NAME=billing PORT=8086 GRPC_PORT=9086"
    exit 1
fi

SERVICE_NAME="${NAME}-service"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
TEMPLATE_DIR="${BACKEND_DIR}/template-service"
TARGET_DIR="${BACKEND_DIR}/${SERVICE_NAME}"

echo "=========================================================="
echo "Создание микросервиса: ${SERVICE_NAME}"
echo "HTTP Port: ${PORT}, gRPC Port: ${GRPC_PORT}"
echo "Целевая директория: ${TARGET_DIR}"
echo "=========================================================="

if [ -d "${TARGET_DIR}" ]; then
    echo "Ошибка: Директория ${TARGET_DIR} уже существует."
    exit 1
fi

if [ ! -d "${TEMPLATE_DIR}" ]; then
    echo "Ошибка: Шаблонный сервис ${TEMPLATE_DIR} не найден."
    exit 1
fi

# 1. Копирование шаблона
echo "[1/6] Копирование файлов из template-service..."
cp -r "${TEMPLATE_DIR}" "${TARGET_DIR}"

# Очистка локальных бинарников и кэша шаблона, если они были скопированы
rm -rf "${TARGET_DIR}/bin" "${TARGET_DIR}/${SERVICE_NAME}" "${TARGET_DIR}/template-service"

# 2. Замена плейсхолдеров в файлах сервиса
echo "[2/6] Замена параметров и конфигураций..."

# Преобразование в UpperCamelCase для Go (например, billing-service -> BillingService)
# и простая очистка имени (billing -> billing)
PASCAL_NAME="$(echo "${NAME}" | awk -F'-' '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) tolower(substr($i,2))}1' OFS="")"
DB_NAME="$(echo "${NAME}" | tr '-' '_')_db"

find "${TARGET_DIR}" -type f \( -name "*.go" -o -name "*.yaml" -o -name "*.yml" -o -name "*.mod" -o -name "Dockerfile" -o -name "*.md" -o -name "*.sql" \) | while read -r file; do
    sed -i "s/template-service/${SERVICE_NAME}/g" "${file}"
    sed -i "s/template_service/${NAME}_service/g" "${file}"
    sed -i "s/template_db/${DB_NAME}/g" "${file}"
    sed -i "s/TemplateService/${PASCAL_NAME}Service/g" "${file}"
    sed -i "s/template/${NAME}/g" "${file}"
    sed -i "s/8086/${PORT}/g" "${file}"
    sed -i "s/9086/${GRPC_PORT}/g" "${file}"
done

# 3. Регистрация в docker-compose.yml
echo "[3/6] Проверка docker-compose.yml..."
COMPOSE_FILE="${BACKEND_DIR}/../docker-compose.yml"
if [ -f "${COMPOSE_FILE}" ]; then
    if grep -q "${SERVICE_NAME}:" "${COMPOSE_FILE}"; then
        echo "Сервис ${SERVICE_NAME} уже зарегистрирован в docker-compose.yml."
    else
        cat <<EOF >> "${COMPOSE_FILE}"

  ${SERVICE_NAME}:
    build:
      context: ./backend
      dockerfile: ${SERVICE_NAME}/Dockerfile
    container_name: avtoplaneta-${SERVICE_NAME}
    restart: unless-stopped
    ports:
      - "${PORT}:${PORT}"
      - "${GRPC_PORT}:${GRPC_PORT}"
    environment:
      - PORT=${PORT}
      - GRPC_PORT=${GRPC_PORT}
      - DATABASE_URL=postgres://avtoplaneta:avtoplaneta@postgres:5432/${DB_NAME}?sslmode=disable
      - REDIS_URL=redis:6379
    networks:
      - avtoplaneta-network
    depends_on:
      - postgres
      - redis
EOF
        echo "Сервис ${SERVICE_NAME} успешно добавлен в docker-compose.yml."
    fi
fi

# 4. Регистрация в backend/Makefile
echo "[4/6] Регистрация таргетов в backend/Makefile..."
MAKEFILE="${BACKEND_DIR}/Makefile"
if [ -f "${MAKEFILE}" ]; then
    # Добавляем в build-services
    if ! grep -q "cd ${SERVICE_NAME} && go build" "${MAKEFILE}"; then
        sed -i "/build-services:/a \\\tcd ${SERVICE_NAME} \&\& go build -tags=\"sonic avx\" -o /dev/null ./..." "${MAKEFILE}"
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

echo "=========================================================="
echo "Микросервис ${SERVICE_NAME} успешно создан!"
echo "Путь: ${TARGET_DIR}"
echo "Для запуска проверки выполните: cd backend && make build-services"
echo "=========================================================="
