@echo off
echo Генерация пароля для Traefik Dashboard...
echo.
echo Для генерации пароля используйте:
echo 1. Установите apache2-utils: apt-get install apache2-utils (Linux) или choco install apache-httpd (Windows)
echo 2. Выполните: htpasswd -nbB admin YOUR_PASSWORD
echo.
echo Или используйте онлайн генератор: https://hostingcanada.org/htpasswd-generator/
echo.
echo Результат добавьте в .env файл как:
echo DASHBOARD_PASSWORD=$$2y$$05$$your_hash_here
pause
