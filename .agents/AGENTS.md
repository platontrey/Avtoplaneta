# Архитектурные правила, контекст и протокол сеньор-разработчика Avtoplaneta (.agents/AGENTS.md)

> Данный файл синхронизирован с корневым [AGENTS.md](../AGENTS.md) и является обязательным руководством для всех ИИ-ассистентов проекта.

---

## 1. Архитектурная матрица сеньор-разработчика (Senior Default Matrix)

| Класс задачи | Стандартное решение (Default) | Критерии применения |
| :--- | :--- | :--- |
| **Параллельная работа внутри одного сервиса** (сжатие пачки фото, парсинг ведомостей) | **Стандартные инструменты Go** (`errgroup.WithContext`, семафор на `chan struct{}`, горутины) | Задача локальна для памяти одного процесса, не требует выживания после рестарта. Обязательна передача `ctx` и лимитирование параллелизма. |
| **Синхронный межсервисный вызов** (списание деталей при заказе `DecreasePartQuantity`, проверка прав) | **gRPC HTTP/2 (Protobuf)** | Требуется немедленная сильная согласованность (Strong Consistency). Таймаут `context.WithTimeout`, идемпотентный `operation_id`, контракты в `backend/proto/`. |
| **Асинхронные события между сервисами** (`seller_renamed`, рассылка уведомлений клиентам) | **Redis Streams** (`XADD`, Consumer Groups `XREADGROUP`, `XACK`, `XAUTOCLAIM`) | Декаплинг сервисов, доставка At-Least-Once, распределение нагрузки по репликам, восстановление зависших сообщений (PEL). |
| **Синхронизация БД с Elasticsearch или брокером (No Dual-Write)** | **Transactional Outbox Pattern** (`outbox_events` + PostgreSQL `LISTEN/NOTIFY` + Outbox Dispatcher) | Атомарная запись в бизнес-таблицу и таблицу `outbox_events` в одной `pgx.Tx`. Мгновенное пробуждение по `pg_notify`. Выборка через `FOR UPDATE SKIP LOCKED`. |
| **Полнотекстовый поиск, опечатки, морфология** | **Elasticsearch** (русский стеммер, синонимы `synonyms.txt`, `QueryBuilder`) | PostgreSQL вторичен для сложного поиска автозапчастей. ES работает как CQRS Read Model. Фильтр `quantity >= 0` обязателен. |
| **Хранилище состояния и транзакции** | **PostgreSQL** (`Database-per-Service`, `sqlc`, `pgx/v5`) | Строгие ACID транзакции, пакетные вставки `pgx.Batch`, запрет на ORM. |
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

## 3. База данных: Миграции, sqlc и pgx/v5

1. **КРИТИЧЕСКОЕ ПРАВИЛО ДВОЙНОГО РАЗМЕЩЕНИЯ (Go embed trap):**
   - В `parts-service`, `orders-service`, `auth-service` файл `database.go` находится в `internal/repository/`.
   - Директива `//go:embed` в Go **не может подниматься вверх (`../`)**.
   - Поэтому ЛЮБАЯ новая миграция создается **ОДНОВРЕМЕННО** в двух папках:
     1. `backend/<service>/db/migrations/0000XX_<name>.up.sql` (и `.down.sql`)
     2. `backend/<service>/internal/repository/migrations/0000XX_<name>.up.sql` (и `.down.sql`)
2. **Безопасность DDL:** `CREATE TABLE IF NOT EXISTS`, `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`, обязательный `.down.sql`, версионирование через таблицу `schema_migrations`.
3. **sqlc:** Запросы в `db/queries/*.sql`. Использовать явные аргументы `sqlc.arg(name)` / `sqlc.narg(name)`. После любых правок запускать `sqlc generate`.
4. **pgx/v5:** Пул `*pgxpool.Pool`. Запрещены одиночные `INSERT` в цикле — только `pgx.Batch`. Для очередей — `FOR UPDATE SKIP LOCKED`.
5. **Автоматическое индексирование сеньора:**
   - **Индекс на каждый Foreign Key:** `CREATE INDEX IF NOT EXISTS idx_<table>_<fk> ON <table>(<fk>)` (Postgres не делает это автоматически; без индекса DELETE родителя блокирует дочернюю таблицу).
   - **Частичные индексы (`WHERE ...`):** Для очередей (`WHERE status = 'pending'`), Soft-Delete (`WHERE deleted_at IS NULL`) и ETag (`WHERE deleted_at IS NULL` по `updated_at DESC`).
   - **Составные индексы:** Правило левого префикса: сначала равенство (`=`), затем диапазоны (`>`, `<`) и сортировка (`ORDER BY`).
   - **Покрывающие (`INCLUDE`):** Для Index-Only Scans без чтения Heap с диска.
   - **Функциональные и GIN:** `LOWER(TRIM(email))`, `TRIM(phone)`, `USING GIN (photos)` для массивов/JSONB.
   - **Constraints:** `CHECK (quantity >= -1)`, `CHECK (price >= 0)`, осознанный `ON DELETE RESTRICT/CASCADE`.
   - **Пагинация:** Keyset pagination `WHERE (created_at, id) < ...` вместо антипаттерна `OFFSET`.
6. **Шаблон и генератор микросервисов (`make new-service`):**
   - Эталонный сервис находится в `backend/template-service`. Содержит готовый `Dockerfile` с кэш-моунтами, `sqlc.yaml`, миграции `schema_migrations`, `/health`, `/metrics` (Prometheus) и Graceful Shutdown.
   - Для создания нового микросервиса запускать: `cd backend && make new-service NAME=<name> PORT=<port> GRPC_PORT=<grpc_port>`. Скрипт сам пропишет модуль в `go.work`, `Makefile` и проверит сборку.

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
