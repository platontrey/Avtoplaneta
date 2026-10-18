# Архитектурные правила, контекст и протокол сеньор-разработчика Avtoplaneta (.agents/AGENTS.md)

> Данный файл синхронизирован с корневым [AGENTS.md](../AGENTS.md) и является обязательным руководством для всех ИИ-ассистентов проекта.

---

## 1. Архитектурная матрица сеньор-разработчика (Senior Default Matrix)

| Класс задачи | Стандартное решение (Default) | Критерии применения |
| :--- | :--- | :--- |
| **Параллельная работа внутри одного сервиса** (сжатие пачки фото, парсинг ведомостей) | **Стандартные инструменты Go** (`errgroup.WithContext`, семафор на `chan struct{}`, горутины) | Задача локальна для памяти одного процесса, не требует выживания после рестарта. Обязательна передача `ctx` и лимитирование параллелизма. |
| **Синхронный межсервисный вызов** (списание деталей при заказе `DecreasePartQuantity`, проверка прав) | **gRPC HTTP/2 (Protobuf)** | Требуется немедленная сильная согласованность (Strong Consistency). Идентификаторы сущностей строго `int64` (Bigint Standard), конвертеры `ToProto()` / `FromProto()`, безопасные геттеры `.Get...()`, таймаут `context.WithTimeout`, идемпотентный `operation_id`, контракты в `backend/proto/`. |
| **Асинхронные события между сервисами** (`seller_renamed`, рассылка уведомлений клиентам) | **Redis Streams** (`XADD`, Consumer Groups `XREADGROUP`, `XACK`, `XAUTOCLAIM`) | Декаплинг сервисов, доставка At-Least-Once, распределение нагрузки по репликам, восстановление зависших сообщений (PEL). |
| **Синхронизация БД с Elasticsearch или брокером (No Dual-Write)** | **Transactional Outbox Pattern** (`outbox_events` + PostgreSQL `LISTEN/NOTIFY` + Outbox Dispatcher) | Атомарная запись в бизнес-таблицу и таблицу `outbox_events` в одной `pgx.Tx`. Мгновенное пробуждение по `pg_notify`. Выборка через `FOR UPDATE SKIP LOCKED`. |
| **Полнотекстовый поиск, опечатки, морфология** | **Elasticsearch** (русский стеммер, синонимы `synonyms.txt`, `QueryBuilder`) | PostgreSQL вторичен для сложного поиска автозапчастей. ES работает как CQRS Read Model. Фильтр `quantity >= 0` обязателен. |
| **Хранилище состояния и транзакции** | **PostgreSQL** (`Database-per-Service`, `sqlc`, `squirrel`, `pgx/v5`) | Строгие ACID транзакции, пакетные вставки `pgx.Batch`, запрет на ORM. `sqlc` для статических запросов (90%), `squirrel` строго для динамических фильтров каталога с 30+ полями и PATCH-апдейтов. `pgx/v5` — фундамент и транспорт. |
| **Кэш витрины и быстрые ответы** | **Redis + ETag / `If-None-Match`** | Сброс ключей `parts:inventory` и `parts:statistics` при изменении остатков. Моментальный ответ `304 Not Modified`. |

---

## 2. Золотой протокол сеньор-инженера: Агент обязан спрашивать пользователя

Агент никогда не принимает архитектурных решений молча.

### 2.1. Триггеры, когда агент ОБЯЗАН СТОПИТЬСЯ и задать вопрос:
1. **Транспортная развилка:** синхронный gRPC vs Redis Streams vs in-process Go concurrency (`errgroup`).
2. **Изменения схемы БД (DDL):** новые таблицы, внешние ключи, составные индексы, добавление колонок с дефолтами, смена статусов.
3. **Фоновая обработка:** воркер внутри процесса (Go ticker) vs Transactional Outbox + Redis Streams.
4. **Контракты API:** добавление или изменение полей в `.proto`, ломающие изменения в REST DTO для React или Flutter.
5. **Стратегия отказоустойчивости:** поведение сервиса при недоступности внешних систем (Elasticsearch / Redis / gRPC).

### 2.2. Обязательный формат вопроса агента:
> ### 🛑 Архитектурная развилка: [Название задачи]
> **Контекст и суть задачи:** [Краткий анализ]
> **Вариант 1 (Рекомендуемый сеньором):** [Решение, обоснование, плюсы и риски]
> **Вариант 2 (Альтернативный):** [Решение, компромиссы]
> **Вопрос к разработчику:** [Конкретный вопрос на утверждение или выбор]

---

## 3. База данных: Миграции, sqlc, squirrel и pgx/v5

1. **КРИТИЧЕСКОЕ ПРАВИЛО ДВОЙНОГО РАЗМЕЩЕНИЯ (Go embed trap):**
   - В `parts-service`, `orders-service`, `auth-service` файл `database.go` находится в `internal/repository/`.
   - Директива `//go:embed` в Go **не может подниматься вверх (`../`)**.
   - Поэтому ЛЮБАЯ новая миграция создается **ОДНОВРЕМЕННО** в двух папках:
     1. `backend/<service>/db/migrations/0000XX_<name>.up.sql` (и `.down.sql`)
     2. `backend/<service>/internal/repository/migrations/0000XX_<name>.up.sql` (и `.down.sql`)
2. **Безопасность DDL:** `CREATE TABLE IF NOT EXISTS`, `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`, обязательный `.down.sql`, версионирование через таблицу `schema_migrations`.
3. **Senior Data Access Matrix (sqlc vs squirrel vs pgx/v5):**
   - **`pgx/v5` (`*pgxpool.Pool`):** Низкоуровневый драйвер, транспорт и пул соединений. Реализует бинарный wire protocol PostgreSQL, управление сокетами, транзакции `pgx.Tx`, `pgx.Batch`, `LISTEN/NOTIFY`. Это фундамент, без которого запрос в БД вообще не уйдет.
   - **`sqlc` (90% всех операций — Default):** Статический компилируемый SQL. Запросы в `db/queries/*.sql` с явными аргументами `sqlc.arg` / `sqlc.narg`. Автогенерация Go-структур, валидация синтаксиса при компиляции, нулевой рантайм-оверхед. После любых правок запускать `sqlc generate`.
   - **`squirrel` (10% операций — Динамический SQL):** Query Builder в памяти процесса. Применяется **строго для динамических фильтров** витрины с 30+ параметрами (`GetPartsWithFilters`), динамических PATCH-апдейтов (`UpdateOrder`) и фильтров аудита, где статический SQL привел бы к 40 условиям `($1::text IS NULL OR col = $1)` и сломал план выполнения Postgres (Seq Scan). `squirrel` не заменяет `pgx` — сгенерированная строка передается на исполнение в `r.pool.Query(...)` драйвера `pgx`.
4. **Архитектура слоя SQL и стандарты запросов (Domain-Driven SQL & CTE):**
   - **Запрет на монолитный `queries.sql`:** Файлы запросов декомпозируются по субдоменам в `db/queries/`: `<entity>_crud.sql`, `<entity>_stock.sql`, `<entity>_batches.sql`, `<entity>_stats.sql`, `outbox.sql`.
   - **Стандарт CTE (`WITH ... AS`):** Запрещены нечитаемые вложенные подзапросы `FROM (SELECT ...) sub` и inline-SELECT в JOIN. Любая предварительная подготовка данных, фильтрация или агрегация оформляется через декларативные CTE с понятными именами.
5. **pgx/v5 практики:** Пул `*pgxpool.Pool`. Запрещены одиночные `INSERT` в цикле — только `pgx.Batch`. Для очередей — `FOR UPDATE SKIP LOCKED`.
6. **Автоматическое индексирование сеньора:**
   - **Индекс на каждый Foreign Key:** `CREATE INDEX IF NOT EXISTS idx_<table>_<fk> ON <table>(<fk>)` (Postgres не делает это автоматически; без индекса DELETE родителя блокирует дочернюю таблицу).
   - **Частичные индексы (`WHERE ...`):** Для очередей (`WHERE status = 'pending'`), Soft-Delete (`WHERE deleted_at IS NULL`), ETag (`WHERE deleted_at IS NULL` по `updated_at DESC`) и дефектных партий (`WHERE quantity = 0 AND to_delete_at IS NULL AND deleted_at IS NULL`).
   - **Составные индексы:** Правило левого префикса: сначала равенство (`=`), затем диапазоны (`>`, `<`) и сортировка (`ORDER BY`).
   - **Покрывающие (`INCLUDE`):** Для Index-Only Scans без чтения Heap с диска.
   - **Функциональные и GIN:** `LOWER(TRIM(email))`, `TRIM(phone)`, `BTRIM(supplier_code)` (предотвращает Seq Scan при вызове строковых функций), `USING GIN (photos)` для массивов/JSONB.
   - **Constraints:** `CHECK (quantity >= -1)`, `CHECK (price >= 0)`, осознанный `ON DELETE RESTRICT/CASCADE`.
   - **Пагинация:** Keyset pagination `WHERE (created_at, id) < ...` вместо антипаттерна `OFFSET`.
7. **Шаблон и генератор микросервисов (`make new-service`):**
   - Эталонный сервис находится в `backend/template-service`. Содержит готовый `Dockerfile` с кэш-моунтами, `sqlc.yaml`, миграции `schema_migrations`, `/health`, `/metrics` (Prometheus) и Graceful Shutdown.
   - Для создания нового микросервиса запускать: `cd backend && make new-service NAME=<name> PORT=<port> GRPC_PORT=<grpc_port>`. Скрипт сам пропишет модуль в `go.work`, `Makefile` и проверит сборку.
7. **Стандарты gRPC, Protobuf и Buf (Bigint Standard, Clean Mappers & Protovalidate):**
   - **Buf тулчейн и категория `STANDARD`:** Управление контрактами выполняется строго через Buf CLI (`buf.yaml`, зависимости BSR `googleapis`, `protovalidate`). Строго соблюдается категория `STANDARD`: каждый RPC обязан иметь строго парные `<Method>Request` и `<Method>Response`. Отключение правил линтинга запрещено. Проверка: `make proto-lint`, форматирование: `make proto-format`. Для совместимости с REST/Frontend используется аннотация `response_body: "<поле>"` в `google.api.http`.
   - **Декларативная валидация Protovalidate (CEL):** Все проверки (непустые поля, форматы email/телефон, диапазоны `id > 0`, `quantity >= -1`, `price >= 0`) описываются в `.proto` через правила `buf.validate`. Ручная валидация в хэндлерах запрещена — всё валидируется автоматически через `pkg/grpcserver.ValidationUnaryInterceptor()`.
   - **Bigint Standard (единый 64-битный ID):** Все идентификаторы сущностей (`id`, `user_id`, `part_id`, `order_id`, `seller_id`, `resource_id` и срезы `ids`, `user_ids`) в `.proto` объявляются строго как `int64`. Категорически запрещено использовать `uint32`/`int32` для PK/FK — в PostgreSQL это `BIGSERIAL`/`BIGINT` (`int64` в Go). Неоднородность типов порождает опасные и громоздкие касты `uint32(...)` / `int64(...)`.
   - **Методы конвертеров `ToProto()` / `FromProto()`:** Каждая доменная модель (`internal/domain`) обязана иметь метод `(m *Model) ToProto() *protov1.Model`. gRPC-хэндлеры не должны заниматься ручной поэлементной перекладкой десятков полей и объявлением временных переменных.
   - **Безопасные геттеры Protobuf:** Все чтения полей входящих запросов gRPC (`req`) производятся строго через сгенерированные геттеры (`req.GetId()`, `req.GetEmail()`, `req.GetItems()`), исключающие панику при `nil`-указателях.

---

## 4. Elasticsearch и CQRS

1. **PostgreSQL — единственный Source of Truth.** Elasticsearch — читающая проекция (CQRS read-model).
2. **Синхронизация:** Строго через Transactional Outbox (никакого прямого Dual-Write из HTTP-хэндлеров).
3. **Отказоустойчивость:** Если Elasticsearch недоступен, сервис логирует `Warn` и переключается на fallback-поиск через PostgreSQL (`ILIKE`), не аварийно завершая работу.
4. **Фильтрация:** В любых поисковых запросах обязателен фильтр `quantity >= 0`.

---

## 5. Критические доменные инварианты

- **Семантика поля `quantity`:**
  - `quantity > 0`: в наличии на складе.
  - `quantity == 0`: дефектные ведомости / спецпозиции. **ОБЯЗАНЫ ОСТАВАТЬСЯ ВИДИМЫМИ** в поиске, UI и XML-выгрузках.
  - `quantity == -1`: полностью списанные/проданные детали через `orders-service`. Маркер сокрытия из выдачи.
  - **Правило SQL/ES:** ВСЕГДА использовать `WHERE quantity >= 0`. Запрещено писать `> 0`.
- **Идемпотентность списаний:** Учет в таблице `part_stock_operations`.
- **CRM и дедупликация:** Нормализация `parseBuyerInfo(raw)` к `+7XXXXXXXXXX`, уникальный индекс `idx_customers_unique_phone`, `409 Conflict` на дубликаты.

---

## 6. Коммуникация, Git и Workflow

- **Язык общения:** **ИСКЛЮЧИТЕЛЬНО РУССКИЙ ЯЗЫК**.
- **Локальные коммиты:** Обязательны (`git add`, `git commit` с понятным сообщением).
- **СТРОГИЙ ЗАПРЕТ НА `git push`:** **НИКОГДА не пушить от имени пользователя.**
- **Docker Compose:** Всегда `docker compose` (с пробелом).
