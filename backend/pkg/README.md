# Инфраструктурная библиотека бэкенда Avtoplaneta (`backend/pkg`)

Каталог `backend/pkg/` содержит переиспользуемые Go-модули и инфраструктурные компоненты, общие для всех микросервисов проекта (`auth-service`, `parts-service`, `orders-service`, `messaging-service`, `export-service`, `template-service`).

> ⚠️ **Архитектурное правило (DRY):**  
> Категорически запрещено дублировать код этих компонентов в каталогах `internal/` микросервисов. Если функционал уже реализован и протестирован в `pkg/`, сервисы обязаны импортировать его из `avtoplaneta/pkg/<name>`.

---

## Реестр пакетов библиотеки

| Модуль | Пакет | Назначение и функциональность |
| :--- | :--- | :--- |
| **`pkg/logger`** | `logger` | Структурированное логирование на базе стандартного `log/slog`. Поддержка JSON/текстового формата, уровней (Debug, Info, Warn, Error) и автоматической связки с OpenTelemetry (`trace_id`, `span_id`). Хелпер `logger.Err(err)`. |
| **`pkg/tracing`** | `tracing` | Инициализация глобального OpenTelemetry TracerProvider, экспорт спанов в Grafana Tempo по протоколу OTLP (gRPC/HTTP), W3C Trace Context. |
| **`pkg/postgres`** | `postgres` | Пул соединений PostgreSQL (`*pgxpool.Pool`) со встроенной трассировкой SQL-запросов через `otelpgx`, проверкой доступности (Ping) и экспортом метрик пула соединений. |
| **`pkg/redisclient`** | `redisclient` | Клиент Redis (`*redis.Client`) со встроенной трассировкой команд через `otelredis` и поддержкой конфигурации через переменные окружения. |
| **`pkg/metrics`** | `metrics` | Унифицированный сбор Prometheus метрик: Gin middleware `metrics.GinMiddleware(service)` с защитой от High Cardinality (через `c.FullPath()`), эндпоинт `/metrics` и счетчики Outbox. |
| **`pkg/outbox`** | `outbox` | Универсальный движок Transactional Outbox диспетчера: мгновенное пробуждение через PostgreSQL `LISTEN/NOTIFY`, страховочный поллинг, ретраи ошибок и сбор метрик. |
| **`pkg/grpcserver`** | `grpcserver` | Каркас запуска gRPC серверов: встроенная валидация `protovalidate` (CEL), перехватчики логирования и трейсинга, стандартный `grpc.health.v1`. |
| **`pkg/grpcclient`** | `grpcclient` | Клиентский транспорт gRPC: пул соединений, таймауты, интерцепторы трейсинга OpenTelemetry и метрик. |
| **`pkg/httpserver`** | `httpserver` | Запуск и graceful shutdown HTTP-серверов (Gin) по системным сигналам с поддержкой контекста отмены (`errgroup`). |
| **`pkg/health`** | `health` | Стандартизированные Liveness (`/health/live`) и Readiness (`/health/ready`, `/health`) эндпоинты с параллельным опросом PostgreSQL, Redis и выдачей HTTP 503 при деградации. |
| **`pkg/ginmw`** | `ginmw` | Набор middleware для Gin: кросс-доменный доступ (`CORS`), проверка сессий через Traefik ForwardAuth (`RequireAuth`), ролевой доступ (`RequireRole`, `RequireAdmin`) и стандартизированные ответы (`RespondOK`, `RespondError`). |
| **`pkg/constants`** | `constants` | Общесистемные константы: маркеры остатков (`PartQuantitySold = -1`, `PartQuantityDefective = 0`), стримы Redis Streams (`events:orders`, `events:parts`), каналы LISTEN/NOTIFY, системные роли. |
| **`pkg/redisstream`** | `redisstream` | Надежный `ConsumerGroupWorker` для Redis Streams: авто-создание стрима/группы, пакетное чтение, авто-восстановление зависших сообщений из PEL (`XAutoClaim`) и отправка битых событий в DLQ. |
| **`pkg/authcontext`** | `authcontext` | Извлечение и парсинг доверенных заголовков Traefik ForwardAuth (`X-User-ID`, `X-User-Role`, `X-User-Name`, `X-User-Email`) из `context.Context` или `*gin.Context`. |
| **`pkg/httpcache`** | `httpcache` | Хелперы условного кэширования: генерация и валидация ETag, обработка заголовков `If-None-Match` с моментальным ответом `304 Not Modified`. |
| **`pkg/userdirectory`** | `userdirectory` | In-memory кэширующий клиент для разрешения имен и ролей пользователей через gRPC вызовы к `auth-service`. |

---

## Примеры быстрого подключения

### 1. Логирование (`pkg/logger`)
```go
import (
    "log/slog"
    "avtoplaneta/pkg/logger"
)

// Инициализация в main.go
log := logger.New(logger.Options{
    Level:  slog.LevelInfo,
    Format: logger.FormatJSON,
})
slog.SetDefault(log)

// Использование в коде (всегда с передачей ctx для OpenTelemetry трейсинга!)
log.InfoContext(ctx, "Пользователь успешно вошел", slog.Int64("user_id", user.ID))

if err != nil {
    log.ErrorContext(ctx, "Не удалось сохранить заказ", logger.Err(err), slog.Int64("order_id", orderID))
}
```

### 2. Подключение к PostgreSQL с трейсингом (`pkg/postgres`)
```go
import "avtoplaneta/pkg/postgres"

pool, err := postgres.NewPool(ctx, postgres.Config{
    URL:             os.Getenv("PARTS_DATABASE_URL"),
    MaxConns:        25,
    MinConns:        5,
    MaxConnLifetime: 1 * time.Hour,
    MaxConnIdleTime: 15 * time.Minute,
})
if err != nil {
    log.ErrorContext(ctx, "Ошибка подключения к БД", logger.Err(err))
    return err
}
defer pool.Close()
```

### 3. Диспетчер Transactional Outbox (`pkg/outbox`)
```go
import "avtoplaneta/pkg/outbox"

dispatcher := outbox.NewDispatcher(pool, storageAdapter, processEventFunc, outbox.Options{
    Name:         "Parts",
    Channel:      "outbox_events_channel",
    BatchSize:    50,
    PollInterval: 100 * time.Millisecond,
})

// Запуск в фоновом воркере
go func() {
    if err := dispatcher.Start(ctx); err != nil {
        log.ErrorContext(ctx, "Ошибка диспетчера outbox", logger.Err(err))
    }
}()
```

---

## Локальное тестирование и проверка пакетов

Для запуска тестов библиотеки:
```bash
cd backend
make test      # Прогон юнит-тестов всех сервисов и pkg
```
