# Архитектурные правила и протокол разработки Avtoplaneta (AGENTS.md)

Главный источник инженерных стандартов проекта **Avtoplaneta** (Go, PostgreSQL, Redis Streams, Elasticsearch, gRPC, Traefik, React, Flutter). Обязателен для всех разработчиков и ИИ-ассистентов.

---

## 1. Архитектурная матрица выбора инструментов (Senior Default Matrix)

| Класс задачи | Стандартное решение (Default) | Ограничения и контракт |
| :--- | :--- | :--- |
| **Параллелизм внутри процесса** | `errgroup.WithContext` + семафор на `chan struct{}` | Локально для памяти процесса; передача `ctx` обязательна; запрещены горутины без контекста и graceful shutdown. |
| **Синхронный вызов `service ↔ service`** | **gRPC HTTP/2 (Protobuf)** | Необходима сильная согласованность (Strong Consistency). Идентификаторы строго `int64` (Bigint Standard), конвертеры `ToProto()` / `FromProto()`, безопасные геттеры `.Get...()`, явный таймаут `context.WithTimeout`, идемпотентный `operation_id`. |
| **Асинхронные события между сервисами** | **Redis Streams** (Consumer Groups) | Слабая связанность (Eventual Consistency). Гарантия At-Least-Once (`XREADGROUP`, `XACK`), фоновое авто-восстановление зависших сообщений из PEL (`XAutoClaim`) через `pkg/redisstream`. |
| **Синхронизация БД с ES/Redis (No Dual-Write)** | **Transactional Outbox Pattern** | Запись сущности и события в `outbox_events` строго в одной `pgx.Tx`. Мгновенное пробуждение через PostgreSQL `LISTEN/NOTIFY` (`backend/pkg/outbox/dispatcher.go`). Выборка воркером: `FOR UPDATE SKIP LOCKED`. |
| **Полнотекстовый поиск, опечатки, морфология** | **Elasticsearch** (CQRS Read Model) | Русский стеммер, синонимы `synonyms.txt`. PostgreSQL — единственный Source of Truth. Обязателен Graceful Fallback на PostgreSQL (`ILIKE`) при отказе ES. |
| **Хранение состояния и транзакции** | **PostgreSQL 16** (`Database-per-Service`) | Пул `*pgxpool.Pool` (`pkg/postgres`), компилируемый SQL (`sqlc` — 90% операций), динамический SQL (`squirrel` — только фильтры каталога 30+ полей и PATCH). Запрещены ORM. |
| **Кэш витрины** | **Redis + ETag / `If-None-Match`** | Сброс ключей `parts:inventory` и `parts:statistics` через Outbox. Моментальный ответ `304 Not Modified` через `pkg/httpcache`. |

---

## 2. Протокол согласования: Агент обязан спрашивать пользователя

Агент **не принимает архитектурных решений молча**. 

### 2.1. Триггеры обязательной остановки (Stop & Ask):
1. **Выбор транспорта/исполнения:** gRPC vs Redis Streams vs фоновый воркер Outbox vs горутины в процессе.
2. **Изменения схемы БД (DDL):** новые таблицы, внешние ключи, составные индексы, добавление колонок с дефолтами, смена типов/статусов.
3. **Моделирование данных и поиск:** добавление полей в поисковый индекс Elasticsearch, изменение анализаторов текста.
4. **Контракты API:** изменение сигнатур Protobuf RPC, удаление полей в `.proto`, ломающие изменения REST эндпоинтов.
5. **Стратегия отказоустойчивости:** поведение при отказе внешнего сервиса (падать с ошибкой, возвращать деградированные данные, fallback в БД).

### 2.2. Обязательный формат вопроса:
> ### 🛑 Архитектурная развилка: [Название задачи]
> **Контекст:** [Краткая суть развилки]  
> **Вариант 1 (Рекомендуемый):** [Суть решения, транзакционность, надежность]  
> **Вариант 2 (Альтернативный):** [Компромиссы, почему уступает первому]  
> **Вопрос к разработчику:** [Конкретный вопрос на подтверждение]

---

## 3. Стандарты базы данных (PostgreSQL, sqlc, pgx/v5)

Изоляция: **Database-per-Service** (`AUTH_DATABASE_URL`, `PARTS_DATABASE_URL`, `ORDERS_DATABASE_URL`, `MESSAGING_DATABASE_URL` с fallback на `DATABASE_URL`). Прямой доступ к чужим БД запрещен.

### 3.1. Миграции и правило двойного размещения (Go embed trap)
- Именование: `0000XX_name.up.sql` и `0000XX_name.down.sql` (6 цифр, идемпотентность `IF NOT EXISTS`).
- **Правило размещения:** Директива `//go:embed` в Go не поднимается в родительские каталоги (`../`). Новая миграция создается **ОДНОВРЕМЕННО** в двух директориях:
  1. `backend/<service>/db/migrations/` (для `sqlc.yaml`).
  2. `backend/<service>/internal/repository/migrations/` (для рантайм-раннера `schema_migrations`).

### 3.2. Архитектура SQL слоя и sqlc
- Запрещен монолитный `queries.sql`. Файлы декомпозируются по субдоменам в `db/queries/`: `<entity>_crud.sql`, `<entity>_stock.sql`, `<entity>_batches.sql`, `<entity>_stats.sql`, `outbox.sql`.
- **Запрещены вложенные подзапросы** `FROM (SELECT ...)`. Использовать Common Table Expressions (**CTE**, `WITH ... AS`).
- Использовать именованные аргументы: `sqlc.arg(name)` или `sqlc.narg(name)`. Позиционные `$1, $2` запрещены.
- После правок в `.sql`: выполнить `cd backend/<service> && sqlc generate`.

### 3.3. Разделение sqlc vs squirrel vs pgx/v5
- **`pgx/v5` (`*pgxpool.Pool`):** Фундамент. Сетевой wire protocol, пул соединений, транзакции `pgx.Tx`, пакетные вставки `pgx.Batch`, `LISTEN/NOTIFY`.
- **`sqlc` (90% всех операций — Default):** Статический типизированный SQL. Нулевой рантайм-оверхед, автогенерация Go-структур.
- **`squirrel` (10% операций — Динамический SQL):** Применяется **строго для динамических фильтров** каталога с 30+ параметрами (`GetPartsWithFilters`) и PATCH-апдейтов. Защищает оптимизатор PostgreSQL от деградации в Seq Scan при 40 условиях `($1 IS NULL OR col = $1)`.

### 3.4. Индексирование, Constraints и производительность
1. **Каждый Foreign Key ОБЯЗАН иметь B-Tree индекс:** PostgreSQL не индексирует FK автоматически; отсутствие индекса блокирует дочернюю таблицу при DELETE/UPDATE родителя.
2. **Частичные индексы (`WHERE ...`):**
   - Для очередей Outbox: `CREATE INDEX idx_outbox_pending ON outbox_events(status, created_at) WHERE status = 'pending';`
   - Для мягкого удаления: `WHERE deleted_at IS NULL;`
   - Для ETag: `ON parts(updated_at DESC) WHERE deleted_at IS NULL;`
   - Для нулевых остатков: `WHERE quantity = 0 AND to_delete_at IS NULL AND deleted_at IS NULL;`
3. **Составные индексы:** Правило левого префикса: сначала равенство (`=`), затем диапазоны (`>`, `<`) и сортировка (`ORDER BY`).
4. **Покрывающие индексы (`INCLUDE`):** Для выполнения Index-Only Scan без чтения Heap с диска.
5. **Функциональные индексы:** `LOWER(TRIM(email))`, `TRIM(phone)`, `BTRIM(supplier_code)` исключают Seq Scan при сравнении нормализованных строк.
6. **GIN-индексы:** Для массивов (`photos TEXT[]`) и JSONB.
7. **Constraints:** Бизнес-инварианты фиксируются в БД: `CHECK (quantity >= -1)`, `CHECK (price >= 0)`.
8. **Пагинация:** Keyset pagination `WHERE (created_at, id) < (arg, arg) ORDER BY created_at DESC, id DESC LIMIT n;` вместо антипаттерна `OFFSET`.
9. **Пакетная запись:** Массовые вставки выполнять строго через `pgx.Batch`, одиночные `INSERT` в цикле запрещены.

### 3.5. Шаблон и генерация новых сервисов (`make new-service`)
- Эталонный сервис находится в `backend/template-service` (многоэтапный `Dockerfile` с кэш-моунтами, `sqlc.yaml`, раннер миграций, graceful shutdown, health/metrics).
- Команда создания: `cd backend && make new-service NAME=<name> PORT=<port> GRPC_PORT=<grpc_port>`.

---

## 4. Elasticsearch и CQRS

1. **PostgreSQL — единственный Source of Truth.** Elasticsearch — денормализованная поисковая проекция (CQRS Read Model).
2. **Синхронизация:** Исключительно через Transactional Outbox. Прямой Dual-Write из HTTP/gRPC хэндлеров категорически запрещен.
3. **Отказоустойчивость (Graceful Fallback):** При недоступности Elasticsearch сервис логирует `Warn` и автоматически переключается на fallback-поиск через PostgreSQL (`ILIKE`), не завершая работу аварийно.
4. **Поисковые фильтры:** В любых поисковых запросах деталей обязателен фильтр `quantity >= 0`.

---

## 5. Архитектура сервисов (Pragmatic Clean Architecture & DDD в Go)

### 5.1. Слоистая структура (`internal/`)
Однонаправленный поток зависимостей:
`Transport (handler, grpc) ──► Application (service) ──► Domain (domain) ◄── Infrastructure (repository, outbox)`
- **`internal/domain` (Ядро бизнеса):** Чистые бизнес-сущности, методы инвариантов, ошибки. Не импортирует другие внутренние пакеты.
- **`internal/repository` (Инфраструктура БД):** Работа с базой через `sqlc` + `squirrel` + `pgx/v5`.
- **`internal/service` (Прикладной слой):** Оркестрация сценариев, транзакции, вызовы gRPC-клиентов, запись Outbox-событий.
- **`internal/handler` и `internal/grpc` (Первичные адаптеры):** Прием внешних запросов, парсинг DTO, валидация.

### 5.2. Богатая доменная модель (Rich Domain Model) vs DTO
- Бизнес-правила и валидации инкапсулируются в методах самих сущностей в `domain` (чистые юнит-тесты без моков БД).
- Запрещено засорять `domain` форматированными строками представления (`CreatedAtFormatted`, `TimeAgo`) или запросами фронтенда (`CreateOrderRequest`). DTO принадлежат слою `handler`.
- Запрещен «маппинг ради маппинга»: в Go модели `sqlc` маппятся напрямую в `domain`, без промежуточных пустых слоев.

### 5.3. Реестр общей библиотеки `backend/pkg/*` (Обязательное переиспользование)
Категорически запрещено дублировать в `internal/` то, что реализовано в общей библиотеке `avtoplaneta/pkg/`:
- **`pkg/logger`**: Структурированный `log/slog` с поддержкой JSON/текста и привязкой к OTel `trace_id`.
- **`pkg/tracing`**: Инициализация OpenTelemetry TracerProvider и экспорт спанов в Grafana Tempo.
- **`pkg/postgres`**: Пул соединений PostgreSQL (`*pgxpool.Pool`) с автоматическим трейсингом `otelpgx`.
- **`pkg/redisclient`**: Клиент Redis (`*redis.Client`) со встроенным трейсингом `otelredis`.
- **`pkg/metrics`**: Единый сбор Prometheus метрик (`metrics.GinMiddleware(service)` и метрики Outbox).
- **`pkg/outbox`**: Диспетчер Transactional Outbox (PostgreSQL LISTEN/NOTIFY, поллинг, ретраи).
- **`pkg/grpcserver` / `pkg/grpcclient`**: Серверные и клиентские фабрики gRPC с интерцепторами трейсинга, метрик и `protovalidate`.
- **`pkg/httpserver`**: Базовый Gin сервер с graceful shutdown через `context` и `errgroup`.
- **`pkg/health`**: Liveness (`/health/live`) и Readiness (`/health/ready`) проверки с опросом БД и Redis.
- **`pkg/ginmw`**: Middleware CORS, аутентификации Traefik ForwardAuth (`RequireAuth`), ролей (`RequireRole`).
- **`pkg/constants`**: Общесистемные константы (статусы, роли, стримы, каналы).
- **`pkg/redisstream`**: `ConsumerGroupWorker` с надежным PEL-восстановлением (`XAutoClaim`) и DLQ.
- **`pkg/authcontext`**: Извлечение данных пользователя из заголовков ForwardAuth (`X-User-ID`, `X-User-Role`).
- **`pkg/httpcache`**: Хелперы валидации ETag и ответа `304 Not Modified`.
- **`pkg/userdirectory`**: Кэширующий gRPC-клиент каталога пользователей.

### 5.4. Стандарты наблюдаемости (Observability: slog & OpenTelemetry)
- **Логирование:** Использовать только `log/slog` через `pkg/logger`. **СТРОГИЙ ЗАПРЕТ** на `fmt.Println`, `log.Println` и сторонние логгеры (logrus, zap) в продакшене.
- **Контекст:** Обязательно передавать `ctx` в логгер: `log.InfoContext(ctx, "сообщение", slog.Int64("id", id))`. Это связывает логи с трейсами по `trace_id` в Loki. Форматирование ошибок — через `logger.Err(err)`.
- **Трейсинг:** Все вызовы gRPC, внешние запросы к PostgreSQL и Redis трассируются спанами автоматически через интерцепторы и обертки `pkg/postgres`, `pkg/redisclient`.

---

## 6. Критические доменные инварианты (Inventory & Orders)

### 6.1. Семантика поля `quantity` (ЖЕЛЕЗНЫЙ ЗАКОН СКЛАДА)
1. `quantity > 0` — деталь в наличии на складе.
2. `quantity == 0` — дефектная ведомость / спецпозиция. **ОБЯЗАНЫ ОСТАВАТЬСЯ ВИДИМЫМИ** в каталоге, поиске, UI и XML-выгрузках.
3. `quantity == -1` — деталь списана/продана по заказу. Маркер сокрытия из выдачи.
> **ПРАВИЛО ЗАПРОСОВ SQL И ELASTICSEARCH:**  
> При выборке деталей для каталога, поиска и выгрузок **ВСЕГДА** использовать условие:  
> `WHERE quantity >= 0` (включает нули и отсекает проданные `-1`).  
> ❌ **ЗАПРЕЩЕНО:** `WHERE quantity > 0`.

### 6.2. Идемпотентность списаний
Списание остатков (`DecreasePartQuantity`) фиксируется в таблице `part_stock_operations` с уникальным ключом `operation_id` (защита от повторных списаний при сетевых сбоях). При исчерпании остаток переходит в `-1`.

### 6.3. Клиенты и дедупликация (CRM)
Телефон покупателя нормализуется через `parseBuyerInfo(raw)` к формату `+7XXXXXXXXXX`. На уровне БД действует `UNIQUE INDEX idx_customers_unique_phone`. При дубликатах API возвращает `409 Conflict`.

---

## 7. Межсервисные контракты и маршрутизация Traefik

- **100% межсервисных вызовов — строго через gRPC (Protobuf):** Внутренний межсервисный HTTP и прямой доступ к чужим БД запрещены. Контракты в `backend/proto/`. После правок запускать `make proto`.
- **Стандарты Protobuf:**
  - Buf CLI, категория `STANDARD` Google API Style Guide (парные `<Method>Request` и `<Method>Response`).
  - Валидация входных данных строго декларативная через правила `buf.validate` (CEL), без ручных проверок в хэндлерах.
  - **Bigint Standard:** Все идентификаторы сущностей (`id`, `user_id`, `part_id`, `order_id` и срезы `ids`) объявляются strictly `int64` (соответствует `BIGSERIAL` в PostgreSQL). Запрещены `int32`/`uint32` для PK/FK.
  - Методы конвертеров: доменные модели обязаны иметь метод `(m *Model) ToProto() *protov1.Model`.
  - Безопасные геттеры: чтение полей входящих gRPC-запросов выполняется строго через `req.Get...()`.
- **Маршрутизация Traefik:**
  - `/api/v1/*` — типизированный gRPC-Gateway.
  - `/api/*` — фасадный REST Gin для multipart-загрузок фото, файлов Excel, XML-фидов и WebSocket (`/api/messaging/ws`).
  - Проверка прав: централизованно через Traefik ForwardAuth (`auth-service /auth/verify`), заголовки извлекаются через `pkg/authcontext`.

---

## 8. Стандарты клиентских приложений

Подробные стандарты разработки фронтенда и мобильного приложения вынесены в специализированные файлы:
- **Web SPA (React 19 / TypeScript / Vite / TanStack Query v5):** [frontend/AGENTS.md](frontend/AGENTS.md) (INP < 200ms, параллельная загрузка изображений, инвалидация кэшей, debounced поиск).
- **Mobile (Flutter / Riverpod / Dio / GoRouter):** [AvtoplanetaApp/AGENTS.md](AvtoplanetaApp/AGENTS.md) (реестр эндпоинтов `api_endpoints.dart`, оффлайн кэш Hive, отказоустойчивость Dio).

---

## 9. Пошаговый чеклист реализации фичи (Senior Engineering Playbook)

Перед сдачей задачи инженер/агент выполняет шаги:
```
[ ] 1. АРХИТЕКТУРНЫЙ АНАЛИЗ:
       - Сформулированы развилки (gRPC vs Streams; DDL; ES)? Задан ли вопрос пользователю?
[ ] 2. БАЗА ДАННЫХ И МИГРАЦИИ:
       - Миграции созданы одновременно в db/migrations/ И internal/repository/migrations/?
       - Запросы в db/queries/*.sql используют CTE, именованные sqlc.arg и индексы на все FK?
       - Выполнен sqlc generate?
[ ] 3. МЕЖСЕРВИСНЫЕ КОНТРАКТЫ (при изменении API):
       - Обновлены proto-файлы? Выполнен make proto?
       - Соблюден ли Bigint Standard (int64) и CEL валидация?
[ ] 4. ВНУТРЕННЯЯ АРХИТЕКТУРА И РЕЗИЛЬЕНТНОСТЬ:
       - Переиспользованы пакеты из backend/pkg/* (logger, postgres, outbox, metrics)?
       - Логирование через slog.*Context(ctx) с пробросом OTel trace_id?
       - Соблюден доменный инвариант quantity >= 0?
       - Использован Transactional Outbox при синхронизации с ES/Redis?
       - Обработана отмена контекста и Graceful Shutdown?
[ ] 5. СБОРКА И ТЕСТИРОВАНИЕ:
       - make build-services проходит без ошибок?
       - make test проходит успешно?
[ ] 6. ЛОКАЛЬНЫЙ КОММИТ:
       - git add и git commit с осмысленным Conventional Commit сообщением.
       - СТРОГИЙ ЗАПРЕТ НА git push!
```

---

## 10. Коммуникация, Git и Workflow

- **Язык общения:** **ИСКЛЮЧИТЕЛЬНО РУССКИЙ ЯЗЫК**. Все ответы, комментарии к коду и отчеты формулируются на русском языке.
- **Локальные коммиты:** Обязательна фиксация изменений через `git add` и `git commit` в формате Conventional Commits (`feat(...)`, `fix(...)`, `docs(...)`).
- **СТРОГИЙ ЗАПРЕТ НА `git push`:** **НИКОГДА не выполнять `git push` от имени пользователя.** Пуш веток выполняет исключительно человек.
- **Docker Compose:** Всегда использовать синтаксис `docker compose` (с пробелом). Команда `docker-compose` запрещена.\n