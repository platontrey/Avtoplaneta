#!/bin/bash

# Скрипт для запуска backend сервисов в продакшен режиме

echo "Запуск backend сервисов в продакшен режиме..."

# Загрузка переменных окружения
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

# Запуск auth-service в фоне
echo "Запуск auth-service..."
./bin/auth-service &
AUTH_PID=$!

# Ожидание запуска auth-service
sleep 2

# Запуск parts-service в фоне
echo "Запуск parts-service..."
./bin/parts-service &
PARTS_PID=$!

echo "Сервисы запущены:"
echo "auth-service PID: $AUTH_PID"
echo "parts-service PID: $PARTS_PID"

# Функция для остановки сервисов
cleanup() {
    echo "Остановка сервисов..."
    kill $AUTH_PID $PARTS_PID 2>/dev/null
    exit 0
}

# Обработка сигналов для корректного завершения
trap cleanup SIGINT SIGTERM

# Ожидание завершения
wait