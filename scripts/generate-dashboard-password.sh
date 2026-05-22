#!/bin/bash

echo "Генерация пароля для Traefik Dashboard..."
echo ""
echo "Введите желаемый пароль:"
read -s PASSWORD

HASH=$(htpasswd -nbB admin "$PASSWORD" | cut -d: -f2)

echo ""
echo "Добавьте следующие строки в ваш .env файл:"
echo ""
echo "DASHBOARD_USER=admin"
echo "DASHBOARD_PASSWORD=\$$(echo $HASH | sed 's/\$/\$\$/g')"
echo ""
echo "Или используйте онлайн генератор: https://hostingcanada.org/htpasswd-generator/"
