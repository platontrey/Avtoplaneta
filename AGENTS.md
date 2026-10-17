# Архитектурные правила, контекст и протокол сеньор-разработчика Avtoplaneta (AGENTS.md)

Данный документ является **главным источником инженерной истины** и обязательных стандартов проектирования для всех ИИ-ассистентов (агентов) и разработчиков проекта **Avtoplaneta**.

Документ определяет:
1. **Как сеньор принимает архитектурные решения** под наш стек (Go, PostgreSQL, Redis Streams, Elasticsearch, gRPC, React, Flutter).
2. **Когда и как агент ОБЯЗАН спрашивать пользователя** перед реализацией.
3. **Стандартные технические решения (Senior Defaults)** для параллелизма, очередей, миграций БД и полнотекстового поиска.
4. **Критические доменные инварианты**, нарушение которых ломает бизнес-логику склада и заказов.

---

## 1. Архитектурная матрица сеньор-разработчика (Senior Default Matrix)

Сеньор не выбирает инструменты наугад. В стеке Avtoplaneta под каждый класс задач закреплено строго определенное стандартное решение:

```
                  ┌─────────────────────────────────────────────────────────┐
                  │                 ТИП ЗАДАЧИ / ОПЕРАЦИИ                   │
                  └────────────────────────────┬────────────────────────────┘
                                               │
       ┌───────────────────────┬───────────────┴───────────────┬──────────────────────┐
       ▼                       ▼                               ▼                      ▼
┌──────────────┐       ┌───────────────┐               ┌───────────────┐      ┌───────────────┐
│ Внутри одного│       │ Синхронный    │               │ Асинхронное   │      │ Синхронизация │
│   сервиса    │       │ вызов между   │               │ событие между │      │ БД с ES/Redis │
│ (CPU / I/O)  │       │  сервисами    │               │  сервисами    │      │ (No Dual-Write│
└──────┬───────┘       └───────┬───────┘               └───────┬───────┘      └───────┬───────┘
       │                       │                               │                      │
       ▼                       ▼                               ▼                      ▼
┌──────────────┐       ┌───────────────┐               ┌───────────────┐      ┌───────────────┐
│Стандартные Go│       │  gRPC HTTP/2  │               │ Redis Streams │      │ Transactional │
│ инструменты  │       │  (Protobuf)   │               │Consumer Groups│      │    Outbox     │
└──────────────┘       └───────────────┘               └───────────────┘      └───────────────┘
```

### 1.1. Стандартные инструменты Go (Concurrency внутри процесса)

#### Когда использовать:
- Операция **локальна для одного сервиса** и не выходит за пределы его оперативной памяти.
- Не требуется гарантия сохранности задачи при внезапном рестарте контейнера.
- Параллелизация CPU или локального I/O (например, параллельное сжатие 10 загруженных фотографий деталей, валидация и парсинг строк Excel-ведомости дефектовки).
- Фоновый сборщик мусора или периодический локальный опрос (in-memory cleanup по `time.NewTicker`).
- Локальное ограничение частоты запросов (rate limiting) или потокобезопасный кэш справочников (`sync.RWMutex`).

#### Технический стандарт реализации:
1. Для параллельных пачек задач **всегда использовать `errgroup.WithContext`** с семафором для ограничения количества одновременных горутин:
   ```go
   g, ctx := errgroup.WithContext(ctx)
   sem := make(chan struct{}, 4) // ограничение до 4 параллельных воркеров
   for _, photo := range photos {
       photo := photo
       g.Go(func() error {
           select {
           case sem <- struct{}{}:
               defer func() { <-sem }()
           case <-ctx.Done():
               return ctx.Err()
           }
           return processAndSavePhoto(ctx, photo)
       })
   }
   if err := g.Wait(); err != nil {
       return fmt.Errorf("обработка фото завершилась с ошибкой: %w", err)
   }
   ```
2. Для фоновых воркеров внутри процесса **всегда использовать `select` с `<-ctx.Done()`**:
   ```go
   ticker := time.NewTicker(10 * time.Minute)
   defer ticker.Stop()
   for {
       select {
       case <-ctx.Done():
           return ctx.Err()
       case <-ticker.C:
           cleanupTempUploads(ctx)
       }
   }
   ```
3. **Строгий запрет:** Запрещено запускать неконтролируемые горутины `go func() { ... }()` без контекста, обработки паники и механизма отмены при graceful shutdown.

---

### 1.2. Синхронный gRPC (Protobuf) через HTTP/2

#### Когда использовать:
- Синхронное взаимодействие между микросервисами (`service ↔ service`), когда вызывающему сервису **необходим немедленный результат (Strong Consistency)** для завершения текущей бизнес-операции.
- Примеры в Avtoplaneta:
  - `orders-service` завершает заказ и вызывает `parts-service.DecreasePartQuantity` для списания остатка на складе.
  - `export-service` потоком вычитывает склад из `parts-service` (`ListPartsForExport`) и продавцов из `auth-service` (`GetUsers`).
  - Проверка прав и сессий через `auth-service`.

#### Технический стандарт реализации:
1. Контракт объявляется в `backend/proto/<service>/v1/<service>.proto`.
2. После любых правок запускается компиляция: `cd backend && make proto`.
3. **Единый 64-битный стандарт первичных ключей (Bigint Standard):**
   - Все идентификаторы сущностей (`id`, `user_id`, `part_id`, `order_id`, `seller_id`, `resource_id` и срезы `ids`, `user_ids`) в Protobuf контрактах объявляются строго как `int64`.
   - **Запрет на неоднородность типов:** В PostgreSQL первичные ключи `BIGSERIAL`/`BIGINT` соответствуют типу `int64` в Go. Использование `uint32` или `int32` для идентификаторов строго запрещено, так как это приводит к нагромождению небезопасных приведений типов `uint32(...)`/`int64(...)` и риску переполнения.
   - Числовые счетчики и пагинация (`quantity`, `page`, `limit`, `offset`, `total`) остаются `int32`.
4. **Конвертеры доменных моделей (`ToProto()` / `FromProto()`):**
   - Каждая доменная сущность обязана иметь метод `(m *Model) ToProto() *protov1.Model` (и при необходимости фабрики `ModelFromProto`).
   - Запрещено перекладывать десятки полей вручную внутри gRPC-хэндлеров и создавать громоздкие локальные переменные под каждое поле.
5. **Безопасные геттеры Protobuf (`.Get...()`):**
   - Чтение полей входящих gRPC-запросов обязательно выполнять через безопасные сгенерированные геттеры Protobuf (`req.GetId()`, `req.GetEmail()`, `req.GetItems()`). Это гарантирует отсутствие `nil`-паник при чтении опциональных структур и дефолтных значений.
6. Все вызовы gRPC-клиентов обязаны иметь явный таймаут:
   ```go
   callCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
   defer cancel()
   resp, err := partsClient.DecreasePartQuantity(callCtx, req)
   ```
7. Идемпотентность: любые изменяющие команды обязаны передавать `operation_id` (например, `order-<id>-part-<part_id>`), а получатель обязан фиксировать его в таблице `part_stock_operations`.
8. Ошибки маппятся в стандартные gRPC-статусы (`codes.NotFound`, `codes.AlreadyExists`, `codes.InvalidArgument`, `codes.Internal`).

---

### 1.3. Брокер сообщений: Redis Streams

#### Когда использовать:
- Асинхронные события между сервисами (Event-Driven Architecture), когда отправитель **не должен ждать завершения обработки** получателями (Eventual Consistency).
- Событие должно гарантированно доставиться нескольким сервисам (Fan-out) даже в случае временного падения получателей.
- Примеры в Avtoplaneta:
  - `seller_renamed` — продавец сменил имя/организацию в `auth-service`, события доставляются в `parts-service` и `orders-service` для фонового обновления денормализованных полей.
  - Уведомления клиентам (push / SMS / Telegram) при смене статуса заказа (`events:orders`).

#### Технический стандарт реализации:
1. Создание Consumer Group с автоматическим созданием стрима:
   `err := redisClient.XGroupCreateMkStream(ctx, stream, group, "$").Err()` (игнорировать `BUSYGROUP`).
2. Чтение пачками по 10–50 сообщений:
   ```go
   streams, err := client.XReadGroup(ctx, &redis.XReadGroupArgs{
       Group:    group,
       Consumer: consumerName,
       Streams:  []string{stream, ">"},
       Count:    10,
       Block:    2 * time.Second,
   }).Result()
   ```
3. При успешной обработке сообщения обязателен вызов `XACK`:
   `client.XAck(ctx, stream, group, message.ID)`
4. Воркер обязан иметь фоновый цикл восстановления зависших сообщений (Pending Entries List — PEL) через `XPending` + `XClaim` или `XAutoClaim` для гарантированной обработки At-Least-Once после падений.

---

### 1.4. Гарантированная синхронизация: Transactional Outbox Pattern

#### Когда использовать:
- **ЛЮБАЯ операция вставки, обновления или удаления бизнес-сущности в PostgreSQL, которая должна отразиться в Elasticsearch или броере Redis.**
- **КАТЕГОРИЧЕСКИЙ ЗАПРЕТ НА ДВОЙНУЮ ЗАПИСЬ (Dual-Write):** HTTP-обработчик никогда не должен писать напрямую сначала в PostgreSQL, а затем отдельным вызовом в Elasticsearch или Redis. Если сеть упадет между ними, база данных и поисковый индекс разойдутся навсегда.

#### Технический стандарт реализации (как это устроено в `parts-service`):
1. **Единая транзакция:** В одной транзакции `pgx.Tx` выполняется сохранение бизнес-сущности и запись события в таблицу `outbox_events`:
   ```go
   tx, err := r.pool.Begin(ctx)
   if err != nil { return err }
   defer tx.Rollback(ctx)

   qtx := r.queries.WithTx(tx)
   created, err := qtx.CreatePart(ctx, ...)
   if err != nil { return err }

   payload, _ := json.Marshal(map[string]interface{}{"id": created.ID, "name": created.Name})
   _, err = qtx.CreateOutboxEvent(ctx, sqlc.CreateOutboxEventParams{
       AggregateType: "part",
       AggregateID:   created.ID,
       EventType:     "part_created",
       Payload:       payload,
   })
   if err != nil { return err }

   return tx.Commit(ctx)
   ```
2. **Мгновенное пробуждение (Zero-latency):** Триггер в PostgreSQL при коммите выполняет `PERFORM pg_notify('outbox_events_channel', NEW.id::text)`.
3. **Диспетчер аутбокса (`internal/outbox/dispatcher.go`):**
   - Держит постоянное соединение через `LISTEN outbox_events_channel`.
   - При получении сигнала (или по запасному тикеру раз в 3 секунды) забирает пачку событий:
     `SELECT * FROM outbox_events WHERE status = 'pending' ORDER BY id ASC LIMIT 50 FOR UPDATE SKIP LOCKED;`
   - Индексирует деталь в Elasticsearch (`esClient.IndexPart`), сбрасывает кэш в Redis (`parts:inventory`, `parts:statistics`) и при необходимости публикует событие в Redis Streams.
   - Помечает событие как `processed` (`MarkOutboxEventProcessed`) либо инкрементирует `retry_count` при ошибках.

---

## 2. Золотой протокол сеньор-инженера: Агент обязан спрашивать пользователя

Агент никогда не принимает архитектурных решений молча.

### 2.1. Точный список ситуаций (триггеров), когда агент ОБЯЗАН СТОПИТЬСЯ и задать вопрос:
1. **Выбор механизма транспорта или исполнения:**
   - Делать синхронный вызов gRPC или ставить задачу в Redis Streams?
   - Выполнять обработку внутри одного сервиса через горутины или выносить в фоновый воркер через Outbox?
2. **Изменения в структуре базы данных (DDL):**
   - Создание новых таблиц, внешних ключей или составных индексов.
   - Выбор типа хранения: отдельные реляционные колонки против `JSONB`.
   - Добавление полей с дефолтными значениями или изменение существующих статусов заказов/запчастей.
3. **Моделирование данных и поиск в Elasticsearch:**
   - Нужно ли индексировать новое поле в Elasticsearch?
   - Требуется ли добавление синонимов в `synonyms.txt` или специфический анализатор текста?
4. **Контракты API и обратная совместимость:**
   - Изменение сигнатур Protobuf RPC, удаление или реорганизация полей в `.proto`.
   - Модификация REST-эндпоинтов, затрагивающая мобильное приложение (Flutter) или фронтенд (React).
5. **Стратегия отказоустойчивости:**
   - Поведение при отказе внешнего сервиса (Elasticsearch / Redis / gRPC): падать с ошибкой, отвечать частичными данными из кэша, или выполнять fallback-запрос в PostgreSQL?

### 2.2. Обязательный шаблон вопроса агента к пользователю:
Перед тем как писать код, агент обязан вывести структурированный блок:

> ### 🛑 Архитектурная развилка: [Название задачи / фичи]
> **Контекст и суть задачи:** [Краткий анализ того, что требуется сделать]
> 
> **Вариант 1 (Рекомендуемый сеньором):**  
> *Решение:* [Техническая суть решения под наш стек]  
> *Обоснование:* [Почему это оптимально: транзакционность, производительность, надежность]  
> *Плюсы и минусы:* [+ Плюсы / - Минусы или накладные расходы]
> 
> **Вариант 2 (Альтернативный):**  
> *Решение:* [Альтернативный вариант]  
> *Компромиссы:* [В чем уступает рекомендуемому]
> 
> **Вопрос к разработчику:** [Конкретный вопрос, требующий вашего утверждения или выбора]

---

## 3. Стандарты работы с Базой Данных (PostgreSQL, sqlc, pgx/v5)

Каждый сервис изолирован по модели **Database-per-Service** (`AUTH_DATABASE_URL`, `PARTS_DATABASE_URL`, `ORDERS_DATABASE_URL`, `MESSAGING_DATABASE_URL` с fallback на `DATABASE_URL`). Использование чужих БД напрямую строго запрещено.

### 3.1. Как правильно писать и организовывать миграции

#### 1. Структура и наименование файлов:
- Миграции именуются строго последовательно: `0000XX_name.up.sql` и `0000XX_name.down.sql`.
- Номер версии состоит из 6 цифр (например, `000010_create_outbox_events_table.up.sql`).

#### 2. КРИТИЧЕСКОЕ ПРАВИЛО ДВОЙНОГО РАЗМЕЩЕНИЯ (Go embed trap):
- В сервисах `parts-service`, `orders-service`, `auth-service` файл `database.go` находится в папке `internal/repository/`.
- Директива Go `//go:embed` **не имеет права подниматься в родительские каталоги (`../`)**.
- Поэтому генератор `sqlc.yaml` читает схему из `db/migrations/`, а бинарник в рантайме вшивает миграции из `internal/repository/migrations/`.
- **ЖЕЛЕЗНОЕ ПРАВИЛО:** Любая новая миграция ОБЯЗАНА создаваться ОДНОВРЕМЕННО в двух директориях:
  1. `backend/<service>/db/migrations/0000XX_<name>.up.sql` (и `.down.sql`)
  2. `backend/<service>/internal/repository/migrations/0000XX_<name>.up.sql` (и `.down.sql`)
  *(В `messaging-service` `database.go` лежит в корне сервиса и читает напрямую из `db/migrations/`).*

#### 3. Безопасность DDL и идемпотентность:
```sql
-- Таблицы:
CREATE TABLE IF NOT EXISTS customer_notes (
    id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT NOT NULL,
    note TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Колонки:
ALTER TABLE parts ADD COLUMN IF NOT EXISTS supplier_code VARCHAR(100) DEFAULT '';

-- Индексы (всегда именовать idx_<table>_<column>):
CREATE INDEX IF NOT EXISTS idx_customer_notes_customer_id ON customer_notes(customer_id);

-- Down-миграция ОБЯЗАНА полностью и чисто откатывать up-миграцию:
DROP TABLE IF EXISTS customer_notes;
```

#### 4. Версионирование в рантайме (`schema_migrations`):
- При старте сервиса встроенный раннер в `database.go` проверяет таблицу `schema_migrations`.
- Если миграция уже выполнена, она пропускается. Новые `.up.sql` файлы выполняются в алфавитном порядке версий.

### 3.2. Стандарты работы с `sqlc`
1. SQL-запросы пишутся строго в `backend/<service>/db/queries/<entity>.sql`.
2. Обязательные аннотации типизации sqlc:
   - `-- name: GetPartByID :one` — ожидает ровно 1 строку (возвращает `(Model, error)`).
   - `-- name: ListParts :many` — возвращает срез `([]Model, error)`.
   - `-- name: UpdatePart :exec` — не возвращает данные `(error)`.
   - `-- name: DeletePart :execrows` — возвращает количество затронутых строк `(int64, error)`.
3. Использование именованных аргументов:
   Всегда использовать `sqlc.arg(name)` или `sqlc.narg(name)` (nullable), избегая слепых позиционных `$1, $2, $3`:
   ```sql
   -- name: UpdateStock :exec
   UPDATE parts
   SET quantity = sqlc.arg(quantity)::integer, updated_at = NOW()
   WHERE id = sqlc.arg(id);
   ```
4. **ОБЯЗАТЕЛЬНАЯ КОМАНДА:** После любых правок в `.sql` файлах выполнять генерацию:
   ```bash
   cd backend/<service> && sqlc generate
   ```

### 3.3. Стандарты работы с `pgx/v5` в Go-коде
1. **Пакетные операции (`pgx.Batch`):**
   Массовые вставки (дефектовки, импорт каталогов, парсинг) **ЗАПРЕЩЕНО** выполнять в цикле одиночными `INSERT`. Использовать `pgx.Batch`:
   ```go
   batch := &pgx.Batch{}
   for _, item := range items {
       batch.Queue(insertQuery, item.Name, item.Price, item.Quantity)
   }
   br := pool.SendBatch(ctx, batch)
   defer br.Close()
   for range items {
       if _, err := br.Exec(); err != nil {
           return fmt.Errorf("ошибка пакетной вставки: %w", err)
       }
   }
   ```
2. **Конкурентная обработка очередей:**
   При выборке задач или событий из БД несколькими параллельными воркерами обязательно использовать `FOR UPDATE SKIP LOCKED`.

### 3.4. SQL-инструменты сеньора: Индексы, Constraints и производительность PostgreSQL

Сеньор **никогда не создает таблицы и не пишет запросы без автоматического проектирования индексов и ограничений**. Написание SQL без индексирования — признак джуниора.

#### 1. Золотые правила индексирования (Index Automation Rules):

1. **Каждый Foreign Key (FK) ОБЯЗАН иметь B-Tree индекс:**
   - **Почему:** PostgreSQL **НЕ создает** индексы для внешних ключей автоматически.
   - **Катастрофа без индекса:** При удалении или изменении строки в родительской таблице PostgreSQL блокирует дочернюю таблицу (`SHARE ROW EXCLUSIVE`) и сканирует её целиком (`Seq Scan`). При `JOIN` таблицы соединяются через медленный `Hash Join` или `Nested Loop` со сканированием всех страниц.
   - **Автомат сеньора:** Объявил колонку `customer_id` или `order_id` $\rightarrow$ немедленно создал индекс:
     ```sql
     CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);
     CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
     CREATE INDEX IF NOT EXISTS idx_order_items_part_id ON order_items(part_id);
     ```

2. **Частичные индексы (Partial Indexes `WHERE ...`) — стандарт для очередей и Soft-Delete:**
   - **Почему:** Индексировать миллионы неактивных или архивных строк — пустая трата RAM (Buffer Pool) и замедление записи (`write amplification`).
   - **Применение в Avtoplaneta:**
     - **Очереди Outbox:** в таблице `outbox_events` 99% записей уже обработаны (`status = 'processed'`). Полноразмерный индекс раздулся бы на гигабайты. Сеньор индексирует только то, что выбирает воркер:
       ```sql
       CREATE INDEX IF NOT EXISTS idx_outbox_events_pending 
           ON outbox_events(status, created_at) 
           WHERE status = 'pending';
       ```
     - **Мягкое удаление (Soft Delete):**
       ```sql
       CREATE INDEX IF NOT EXISTS idx_parts_seller_id_alive 
           ON parts(seller_id) 
           WHERE deleted_at IS NULL;
       ```
     - **ETag / InventoryVersion (вычисление MAX(updated_at)):**
       ```sql
       CREATE INDEX IF NOT EXISTS idx_parts_updated_at_alive 
           ON parts(updated_at DESC) 
           WHERE deleted_at IS NULL;
       ```

3. **Составные индексы (Composite Indexes) и Правило левого префикса (Leftmost Prefix Rule):**
   - **Формула порядка колонок:** Сначала равенства (`=`, высокая селективность), затем диапазоны (`>`, `<`, `BETWEEN`) и сортировка (`ORDER BY`).
   - **Пример:** Запрос фильтрует по продавцу и сортирует по дате создания:
     `SELECT * FROM parts WHERE seller_id = $1 AND quantity >= 0 ORDER BY created_at DESC;`
     $\rightarrow$ правильный составной индекс:
     ```sql
     CREATE INDEX IF NOT EXISTS idx_parts_seller_created 
         ON parts (seller_id, created_at DESC) 
         WHERE quantity >= 0;
     ```
   - Запрос `WHERE seller_id = $1` может использовать этот индекс. Запрос `WHERE created_at > ...` без `seller_id` — **НЕ может** (правило левого префикса).

4. **Покрывающие индексы (`INCLUDE`) для Index-Only Scan:**
   - Если частый запрос берет всего пару колонок в дополнение к условию:
     `SELECT id, quantity, price FROM parts WHERE category = $1;`
     Индекс с `INCLUDE` позволяет Postgres не обращаться к таблице на диске (Heap), а отдать всё прямо из памяти индекса:
     ```sql
     CREATE INDEX IF NOT EXISTS idx_parts_category_covering 
         ON parts (category) 
         INCLUDE (quantity, price) 
         WHERE deleted_at IS NULL;
     ```

5. **Функциональные / Выраженческие индексы (Expression Indexes):**
   - Для уникальности или поиска без учета регистра:
     ```sql
     -- Уникальность email без учета регистра:
     CREATE UNIQUE INDEX IF NOT EXISTS idx_users_lower_email ON users (LOWER(TRIM(email)));

     -- Уникальный номер клиента с очисткой пробелов и фильтрацией пустых:
     CREATE UNIQUE INDEX IF NOT EXISTS idx_customers_unique_phone 
         ON customers (TRIM(phone)) 
         WHERE phone IS NOT NULL AND TRIM(phone) != '';
     ```

6. **GIN-индексы для JSONB, массивов и полнотекстового поиска:**
   - Для массивов строк (например, фото деталей `photos TEXT[]` или JSONB участников чата):
     ```sql
     CREATE INDEX IF NOT EXISTS idx_parts_photos ON parts USING GIN (photos);
     CREATE INDEX IF NOT EXISTS idx_conversations_participants ON conversations USING GIN (participants);
     ```

#### 2. Ограничения целостности (Constraints — защита данных на уровне СУБД):
Сеньор не надеется только на валидацию в Go-коде. Бизнес-инварианты фиксируются в схеме БД:
1. **CHECK Constraints:**
   ```sql
   -- Защита доменного инварианта склада (quantity >= -1):
   ALTER TABLE parts ADD CONSTRAINT chk_parts_quantity CHECK (quantity >= -1);
   -- Защита от отрицательной цены:
   ALTER TABLE parts ADD CONSTRAINT chk_parts_price CHECK (price >= 0);
   ```
2. **Осознанный выбор `ON DELETE` для внешних ключей:**
   - `ON DELETE RESTRICT` (по умолчанию) — запретить удаление клиента или детали, если по ним есть заказы.
   - `ON DELETE CASCADE` — использовать **только** для жестко связанных сущностей (позиции заказа `order_items` удаляются при удалении заказа `orders`).
   - `ON DELETE SET NULL` — для опциональных связей при архивации.

#### 3. Антипаттерны и золотые правила SQL-запросов:
1. **Антипаттерн `OFFSET` $\rightarrow$ Keyset Pagination (Cursor Pagination):**
   - **В чем проблема:** `OFFSET 100000 LIMIT 50` заставляет PostgreSQL просканировать 100 050 строк диска/памяти и выбросить первые 100 000. На глубоких страницах API начнет зависать.
   - **Решение сеньора:** Пагинация по курсору с составным индексом:
     ```sql
     -- name: ListPartsCursor :many
     SELECT * FROM parts 
     WHERE (created_at, id) < (sqlc.arg(created_at), sqlc.arg(id))
       AND quantity >= 0
     ORDER BY created_at DESC, id DESC 
     LIMIT sqlc.arg(page_size);
     ```
     Поддерживается составным индексом `CREATE INDEX IF NOT EXISTS idx_parts_cursor ON parts (created_at DESC, id DESC) WHERE quantity >= 0;`. Запрос выполняется за доли миллисекунды независимо от глубины страницы.
2. **Идемпотентный UPSERT (`ON CONFLICT`):**
   ```sql
   INSERT INTO part_stock_operations (operation_id, part_id, operation_type, amount)
   VALUES ($1, $2, $3, $4)
   ON CONFLICT (operation_id) DO NOTHING;
   ```
3. **Профилирование через `EXPLAIN (ANALYZE, BUFFERS)`:**
   - Перед отправкой тяжелого запроса в прод сеньор тестирует план:
     - ❌ **Красный флаг:** `Seq Scan` на таблицах более 10 000 строк, большой `Buffers: shared read` (медленный диск).
     -  **Идеал:** `Index Scan` или `Index Only Scan` со 100% `Buffers: shared hit` (быстрое чтение из RAM).

### 3.5. Стандарт создания новых микросервисов (Template & Scaffolding)

Чтобы избежать ошибок с зависимостями в Docker, забытых директив `replace` и ручного дублирования бойлерплейта, в проекте внедрен эталонный шаблон и автоматический генератор:

1. **Эталонный сервис (`backend/template-service`):**
   - Компилируемый микросервис с чистой архитектурой (`domain`, `repository`, `service`, `handler`, `grpc`).
   - Настроенный многоэтапный `Dockerfile` с кеш-моунтами `--mount=type=cache`, пользователем `appuser`, встроенным `HEALTHCHECK`.
   - Полная интеграция с `sqlc`, PostgreSQL (`pgx/v5`), встроенным раннером миграций `schema_migrations`.
   - Подключение `pkg/httpserver`, `pkg/authcontext`, `/health`, `/metrics` (Prometheus) и Graceful Shutdown через `errgroup`.
   - Регулярно тестируется в `make test` и `make build-services` монорепозитория, что гарантирует 100% актуальность шаблона со временем.

2. **Команда генерации (`make new-service`):**
   ```bash
   cd backend && make new-service NAME=<name> PORT=<http_port> GRPC_PORT=<grpc_port>
   # Пример:
   make new-service NAME=billing PORT=8086 GRPC_PORT=9086
   ```
   Скрипт автоматически:
   - Клонирует эталонный шаблон в `backend/<name>-service`.
   - Заменяет имена пакетов, модулей, портов и сущностей.
   - Регистрирует сервис в корневом `go.work` и `backend/Makefile`.
   - Запускает `go mod tidy`, `sqlc generate` и валидирует юнит-тесты `go test ./...`.
   - Генерирует готовый блок конфигурации для вставки в `docker-compose.yml`.

---

## 4. Elasticsearch: Полнотекстовый поиск, CQRS и отказоустойчивость

### 4.1. CQRS и источник правды
- **PostgreSQL — единственный Source of Truth (первичный источник истины).**
- **Elasticsearch — читающая поисковая проекция (CQRS Read Model)** для быстрого полнотекстового поиска с морфологией, автодополнением и опечатками.
- Синхронизация БД $\rightarrow$ Elasticsearch выполняется **ИСКЛЮЧИТЕЛЬНО через Transactional Outbox**.

### 4.2. Морфология и построение запросов
- Поиск автозапчастей использует русский стеммер (`russian_stemmer`), фильтры lowercase и справочник синонимов `backend/parts-service/catalog/synonyms.txt`.
- Поисковые запросы строятся через `internal/search/query_builder.go`:
  - Multi-match запросы с бустом названия запчасти.
  - Fuzzy-поиск для автоматического исправления опечаток («каролла» $\rightarrow$ «corolla»).
  - Фасетная фильтрация (бренд, модель, категория, сторона: лев/прав, перед/зад).

### 4.3. Отказоустойчивость (Graceful Fallback)
- Если Elasticsearch недоступен при старте или упал в рантайме, `parts-service` **НЕ ДОЛЖЕН падать в фатал**.
- Сервис логирует `Warn("Elasticsearch is unavailable, running with database fallback")`.
- Поиск автоматически переключается на PostgreSQL fallback: SQL-запрос с `ILIKE` по колонкам `name`, `description`, `brand`, `model`.

---

## 5. Критические доменные инварианты (Inventory & Orders)

### 5.1. Семантика поля `quantity` (ЖЕЛЕЗНЫЙ ЗАКОН СКЛАДА)
В системе определены 3 фундаментальных состояния остатка деталей:
1. `quantity > 0` — деталь в наличии на складе.
2. `quantity == 0` — дефектная ведомость (акт дефектовки) или спецдеталь. **ОНИ ОБЯЗАНЫ ОСТАВАТЬСЯ ВИДИМЫМИ** в поиске, UI и XML-выгрузках.
3. `quantity == -1` — деталь полностью продана через `orders-service`. Служит маркером сокрытия из выдачи, сохраняя целостность завершенных заказов.

> ⚠️ **ПРАВИЛО ЗАПРОСОВ SQL И ELASTICSEARCH:**
> При выборке деталей для отображения в каталоге, поиске и генерации выгрузок ВСЕГДА использовать условие:
> ```sql
> WHERE quantity >= 0 -- ПРАВИЛЬНО! Включает нулевые дефектовки и отсекает проданные (-1)
> ```
> ❌ **НИКОГДА НЕ ПИСАТЬ:** `WHERE quantity > 0` (это скроет дефектные ведомости!).

### 5.2. Списание остатков и учет операций
- Списание при выполнении заказа (`green`) выполняется через gRPC `DecreasePartQuantity`.
- Таблица `part_stock_operations` фиксирует операцию с уникальным ключом `operation_id` (идемпотентность).
- При исчерпании остатка количество переходит в `-1`:
  ```sql
  UPDATE parts
  SET quantity = CASE WHEN quantity - sqlc.arg(amount)::integer <= 0 THEN -1 ELSE quantity - sqlc.arg(amount)::integer END
  WHERE id = sqlc.arg(id);
  ```

### 5.3. CRM, клиенты и дедупликация
- Номер телефона покупателя нормализуется через `parseBuyerInfo(raw)` к формату `+7XXXXXXXXXX`.
- Ограничение на уровне БД: `UNIQUE INDEX idx_customers_unique_phone`.
- При попытке создания дубликата API возвращает `409 Conflict`.

---

## 6. Межсервисные контракты и маршрутизация Traefik

- **100% межсервисных вызовов — строго через gRPC (Protobuf):**
  - Никаких внутренних HTTP-вызовов между сервисами.
  - Прямой доступ в БД соседнего сервиса запрещен.
  - Контракты описаны в `backend/proto/`. После правок запускать `make proto` (из каталога `backend/`).
  - Команда `make proto` автоматически компилирует код и копирует `swagger.json` во фронтенд (`frontend/src/data/swagger.json`).
- **Разделение маршрутов:**
  - `/api/v1/*` — строго типизированные gRPC-Gateway эндпоинты (JSON представление protobuf).
  - `/api/*` — нативные REST эндпоинты Gin для multipart-загрузок фото, файлов Excel, статических каталогов и WebSocket (`/api/messaging/ws`).
- **Безопасность:** Проверка сессий через Traefik ForwardAuth (`auth-service /auth/verify`). Заголовки `X-User-ID`, `X-User-Role`, `X-User-Name`, `X-User-Email` извлекаются через `pkg/authcontext`.

---

## 7. Frontend (React 19) и Mobile (Flutter)

### 7.1. Frontend (React 19 / TypeScript / Vite / TanStack Query v5)
- **INP (Interaction to Next Paint):** Целевое значение **< 200ms**.
- **Параллельная загрузка фото:** Изображения в карточках обрабатываются параллельно (`Promise.all`), а не по одному.
- **Инвалидация кэша:** После мутаций инвалидировать соответствующие query-ключи (`queryClient.invalidateQueries({ queryKey: [...] })`).
- **Debounced поиск:** Все поисковые поля обязаны использовать debounce во избежание спама запросами.

### 7.2. Mobile App (Flutter / Riverpod / Dio / GoRouter)
- **Единый источник маршрутов:** `AvtoplanetaApp/lib/core/api/api_endpoints.dart`. При добавлении или изменении эндпоинтов на бэкенде обязательно синхронизировать их здесь.
- **Отказоустойчивость:** Обработка оффлайн-режима через Hive кэш и перехват ошибок Dio (`DioException`).

---

## 8. Пошаговый чеклист реализации фичи (Senior Engineering Playbook)

Перед тем как считать задачу выполненной, сеньор проходит по шагам:

```
[ ] 1. АРХИТЕКТУРНЫЙ АНАЛИЗ И ВОПРОС ПОЛЬЗОВАТЕЛЮ:
       - Сформулированы ли развилки (gRPC vs Streams vs Go; DDL; ES)?
       - Задан ли структурированный вопрос пользователю? Получено ли подтверждение?
[ ] 2. БАЗА ДАННЫХ И МИГРАЦИИ:
       - Созданы 0000XX_name.up.sql и 0000XX_name.down.sql?
       - Продублированы ли миграции в db/migrations/ И internal/repository/migrations/?
       - Написаны ли запросы в db/queries/*.sql с именованными аргументами sqlc.arg?
       - Выполнен ли sqlc generate в каталоге сервиса?
[ ] 3. МЕЖСЕРВИСНЫЕ КОНТРАКТЫ (если меняется API):
       - Обновлены proto-файлы в backend/proto/?
       - Выполнен make proto из каталога backend/?
       - Синхронизирован ли swagger.json с фронтендом?
[ ] 4. БИЗНЕС-ЛОГИКА И РЕЗИЛЬЕНТНОСТЬ:
       - Соблюден ли доменный инвариант quantity >= 0?
       - Ошибки обернуты через fmt.Errorf("...: %w", err)?
       - Контекст ctx context.Context проброшен первым аргументом во все вызовы?
       - Обработана ли отмена контекста и Graceful Shutdown?
       - Использован ли Transactional Outbox при синхронизации с ES или Redis?
[ ] 5. КЛИЕНТСКИЕ ПРИЛОЖЕНИЯ:
       - Проверена совместимость с Frontend и Flutter Mobile (api_endpoints.dart)?
[ ] 6. ВЕРИФИКАЦИЯ И СБОРКА:
       - make build-all проходит без ошибок?
       - make test проходит успешно?
[ ] 7. ЛОКАЛЬНЫЙ КОММИТ:
       - git add и git commit с осмысленным сообщением.
       - СТРОГИЙ ЗАПРЕТ НА git push!
```

---

## 9. Коммуникация и Git / CI/CD Workflow

- **Язык общения:** **ИСКЛЮЧИТЕЛЬНО РУССКИЙ ЯЗЫК**. Все ответы, комментарии к коду, вопросы к пользователю и отчеты формулируются на русском языке.
- **Локальные коммиты:** Обязательно фиксировать изменения через `git add` и `git commit` с четким описанием в формате Conventional Commits (например, `feat(parts): add outbox dispatcher for search index`).
- **СТРОГИЙ ЗАПРЕТ НА `git push`:** **НИКОГДА не выполнять `git push` от имени пользователя.** Пользователь вручную пушит ветку для контроля деплоя.
- **Docker Compose:** Всегда использовать синтаксис `docker compose` (с пробелом). Команда `docker-compose` запрещена.
