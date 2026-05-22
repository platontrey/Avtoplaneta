@echo off
REM Скрипт для запуска backend сервисов в продакшен режиме на Windows

echo Запуск backend сервисов в продакшен режиме...

REM Загрузка переменных окружения из .env файла
if exist .env (
    for /f "tokens=*" %%i in (.env) do set %%i
)

REM Запуск auth-service в фоне
echo Запуск auth-service...
start /B bin\auth-service.exe
set AUTH_PID=%!

REM Ожидание запуска auth-service
timeout /t 2 /nobreak > nul

REM Запуск parts-service в фоне
echo Запуск parts-service...
start /B bin\parts-service.exe
set PARTS_PID=%!

echo Сервисы запущены:
echo auth-service PID: %AUTH_PID%
echo parts-service PID: %PARTS_PID%

echo Нажмите Ctrl+C для остановки...
pause