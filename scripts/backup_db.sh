#!/usr/bin/env bash
# ==============================================================================
# Скрипт создания резервной копии базы данных PostgreSQL (без файлов и фото)
# Проект: Avtoplaneta
# ==============================================================================

set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-/opt/avtoplaneta/backups}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILENAME="avtoplaneta_db_${TIMESTAMP}.sql.gz"
BACKUP_PATH="${BACKUP_DIR}/${BACKUP_FILENAME}"

echo "=================================================================="
echo " Начинаем резервное копирование базы данных PostgreSQL..."
echo " Время запуска: $(date '+%Y-%m-%d %H:%M:%S')"
echo "=================================================================="

mkdir -p "${BACKUP_DIR}"

# Проверяем доступность docker compose
if command -v docker &>/dev/null; then
    echo "Запуск pg_dump через docker compose..."
    cd /opt/avtoplaneta 2>/dev/null || true
    docker compose exec -T postgres pg_dump -U postgres avtoplaneta | gzip > "${BACKUP_PATH}"
elif command -v pg_dump &>/dev/null; then
    echo "Запуск локального pg_dump..."
    pg_dump -U "${POSTGRES_USER:-postgres}" "${POSTGRES_DB:-avtoplaneta}" | gzip > "${BACKUP_PATH}"
else
    echo "ОШИБКА: Не найден ни 'docker', ни 'pg_dump'!" >&2
    exit 1
fi

if [ -f "${BACKUP_PATH}" ] && [ -s "${BACKUP_PATH}" ]; then
    FILE_SIZE=$(du -h "${BACKUP_PATH}" | cut -f1)
    echo "------------------------------------------------------------------"
    echo " УСПЕХ: Резервная копия создана!"
    echo " Файл:   ${BACKUP_PATH}"
    echo " Размер: ${FILE_SIZE}"
    echo "------------------------------------------------------------------"
else
    echo "ОШИБКА: Файл резервной копии не был создан или пуст!" >&2
    exit 1
fi

# Ротация: оставляем последние 14 копий
echo "Очистка старых копий (храним последние 14)..."
ls -tp "${BACKUP_DIR}"/*.sql.gz 2>/dev/null | grep -v '/$' | tail -n +15 | xargs -I {} rm -- {} 2>/dev/null || true

echo "Готово."
