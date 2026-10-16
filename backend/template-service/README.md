# Эталонный микросервис Avtoplaneta (template-service)

Данный микросервис является **инженерным стандартом и эталоном чистой архитектуры** в монорепозитории Avtoplaneta. Он соответствует правилам [AGENTS.md](../../AGENTS.md) и используется для создания новых сервисов платформы.

---

## Архитектура слоев (Clean Architecture)

```
template-service/
├── Dockerfile                  # Оптимизированный мультистейдж Dockerfile с кешированием Go
├── sqlc.yaml                   # Конфигурация компилятора SQL -> Go
├── db/
│   ├── migrations/             # DDL миграции (000001_*.up.sql / .down.sql) для sqlc
│   ├── queries/                # Типизированные SQL-запросы (*.sql)
│   └── sqlc/                   # Автогенерированный безопасный код sqlc
├── internal/
│   ├── config/                 # Загрузка env-переменных и настройка Structured Logger
│   ├── domain/                 # Чистые бизнес-модели, валидация и доменные ошибки
│   ├── repository/             # Слой данных (PostgreSQL pgx/v5 + sqlc)
│   │   └── migrations/         # Дубликат миграций для runtime //go:embed
│   ├── service/                # Бизнес-логика и сценарии использования (Use Cases)
│   ├── handler/                # HTTP-адаптеры (Gin, CORS, /health, Prometheus /metrics)
│   └── grpc/                   # gRPC-сервер с Reflection и Health Check
└── main.go                     # Точка входа: graceful shutdown, errgroup, DI
```

---

## Ключевые паттерны и стандарты

1. **База данных и миграции (PostgreSQL + pgx/v5 + sqlc):**
   - Никаких ORM. Запросы пишутся на чистом SQL в `db/queries/`.
   - **Правило двойного размещения (Go embed trap):** Файлы миграций лежат в `db/migrations/` (для генератора `sqlc`) и в `internal/repository/migrations/` (для `//go:embed` в рантайме).
   - В рантайме миграции автоматически применяются и версионируются в таблице `schema_migrations`.

2. **Параллелизм и Graceful Shutdown:**
   - Серверы HTTP (`avtoplaneta/pkg/httpserver`) и gRPC запускаются через `errgroup.WithContext(ctx)`.
   - При получении `SIGINT` или `SIGTERM` контекст отменяется, соединения корректно закрываются без обрыва текущих запросов.

3. **Observability (Метрики и логирование):**
   - `GET /health` — проверка жизнеспособности сервиса.
   - `GET /metrics` — сбор метрик в формате Prometheus (RPS, latency p95/p99, статус-коды).
   - JSON-логирование в production (`logrus.JSONFormatter`).

4. **Безопасность контейнера:**
   - Мультистейдж сборка: компиляция в Alpine Go с кеш-моунтами `--mount=type=cache`, запуск в легковесном `alpine:3.19`.
   - Запуск под непривилегированным пользователем `appuser` (UID без root прав).
   - Встроенный Docker `HEALTHCHECK`.

---

## Как создать новый сервис из этого шаблона

Воспользуйтесь автоматическим генератором:
```bash
cd backend
make new-service NAME=billing PORT=8086 GRPC_PORT=9086
```

Скрипт автоматически:
1. Скопирует шаблон в `backend/billing-service`.
2. Переименует модули, порты и сущности.
3. Добавит новый сервис в `go.work` и `backend/Makefile`.
4. Сгенерирует код `sqlc generate` и выведет блок для `docker-compose.yml`.
