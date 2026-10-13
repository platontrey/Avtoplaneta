# Архитектурные правила и контекст проекта Avtoplaneta (AGENTS.md)

Данный документ является **главным источником архитектурной истины** и обязательных инженерных правил для всех ИИ-ассистентов (агентов) и разработчиков, работающих с репозиторием **Avtoplaneta**.

Все агенты **обязаны** строго следовать описанным стандартам, паттернам проектирования и доменным инвариантам во всех трёх ключевых частях системы: **Backend**, **Frontend** и **Mobile (Flutter)**.

---

## 1. Коммуникация и Git / CI/CD Workflow

- **Язык общения:** **ИСКЛЮЧИТЕЛЬНО РУССКИЙ ЯЗЫК** (Русский язык). Все ответы, описания шагов, комментарии к коду, отчеты и вопросы формулируются на русском языке.
- **Локальные коммиты:** При любых существенных изменениях, рефакторинге или фиксе багов обязательно выполнять локальный коммит (`git add` и `git commit`) с осмысленным описанием в формате Conventional Commits или четким описанием на русском/английском.
- **СТРОГИЙ ЗАПРЕТ НА `git push`:** **НИКОГДА не выполнять `git push` от имени пользователя.** Терминал требует аутентификации, а пользователь вручную пушит ветку для контроля деплоя и CI/CD пайплайна на продакшен-сервере.
- **Docker Compose:** Всегда использовать синтаксис `docker compose` (с пробелом). Устаревшая команда с дефисом `docker-compose` запрещена.

---

## 2. Глобальная топология системы

```
┌────────────────────────┐      ┌────────────────────────┐
│  Web Frontend (React)  │      │ Mobile App (Flutter)   │
│  React 19 / Vite / TS  │      │ Riverpod / Dio / GoRtr │
└───────────┬────────────┘      └───────────┬────────────┘
            │                               │
            └───────────────┬───────────────┘
                            ▼
     ┌─────────────────────────────────────────────────────────┐
     │ Traefik Ingress (SSL/TLS, Routing, ForwardAuth Verify)  │
     └───────────┬──────────────┬─────────────┬────────────┬───┘
                 │ /auth/*      │ /orders/*   │ /api/v1/*  │ /api/messaging/*
                 │              │             │ /uploads/* │
          ┌──────▼───┐   ┌──────▼───┐  ┌──────▼───┐ ┌──────▼───┐
          │ Auth     │   │ Orders   │  │ Parts    │ │Messaging │
          │ Service  │   │ Service  │  │ Service  │ │ Service  │
          │:8083/9083│   │:8082/9082│  │:8081/9081│ │:8084/9084│
          └──────▲───┘   └──────┬───┘  └──────▲───┘ └──────▲───┘
                 │              │             │            │
                 │              │ gRPC        │ gRPC       │
                 │              └────────────►│            │
                 │                    gRPC    │            │
                 ├────────────────────────────┴────────────┘
                 │ (ValidateSession, GetUsers, Statistics)
                 │
                 │ Redis Streams (seller_renamed, orders)
                 ▼
          ┌─────────────────────────────────────────────────────┐
          │ PostgreSQL (Database-per-Service / Fallback)        │
          │ Redis (Кэш сессий, CSRF, WebSocket Hub, Streams)   │
          │ Elasticsearch (Полнотекстовый поиск запчастей)      │
          └─────────────────────────────────────────────────────┘
```

---

## 3. Бэкенд (Go Microservices)

Бэкенд организован в виде микросервисов в каталоге `backend/` с использованием рабочего пространства Go (`go.work`).

### 3.1. Каталог микросервисов и зоны ответственности
1. **Auth Service (`backend/auth-service/`)** — HTTP `:8083`, gRPC `:9083`:
   - Локальная аутентификация (bcrypt) и Google OAuth (`gothic`).
   - Управление сессиями (HttpOnly cookies, Redis session store).
   - Traefik ForwardAuth верификация: `/auth/verify` и `/auth/verify-admin`.
   - Справочник пользователей (`GetUsers` по gRPC) и аудит-лог активности (`LogActivity`).
   - Публикация событий в Redis Streams (`seller_renamed`).
2. **Parts Service (`backend/parts-service/`)** — HTTP `:8081`, gRPC `:9081`:
   - Складской учет запчастей, карточки деталей, галереи изображений (`/uploads/*`).
   - Полнотекстовый поиск с Elasticsearch (морфология, синонимы `synonyms.txt`, fuzzy search).
   - Единый каталог дефектовок и шаблонов (`catalog.json`, `vehicles.json`).
   - Пакетная вставка ведомостей (`pgx.Batch` + Elasticsearch `BulkIndexParts`).
   - Идемпотентное списание/возврат остатков (`DecreasePartQuantity`, `IncreasePartQuantity`).
   - ETag отпечаток склада (`InventoryVersion`).
3. **Orders Service (`backend/orders-service/`)** — HTTP `:8082`, gRPC `:9082`:
   - Жизненный цикл заказов (статусы: `red` -> `brown` -> `yellow` -> `green`).
   - Автоматическое списание запчастей при закрытии заказа через gRPC вызов `DecreasePartQuantity` в `parts-service`.
   - Встроенная CRM клиентов (`customers`), расчет LTV и персональных скидок.
   - Нормализация телефонов (`+7XXXXXXXXXX`) и дедупликация клиентов.
4. **Messaging Service (`backend/messaging-service/`)** — HTTP `:8084`, gRPC `:9084`:
   - Внутренние чаты пользователей, WebSocket hub (`/api/messaging/ws`).
   - Потоковая загрузка голосовых сообщений через gRPC streaming.
   - Интеграция с Drom.ru (парсинг сообщений, отправка ответов).
5. **Export Service (`backend/export-service/`)** — HTTP `:8085`:
   - Генерация XML-фидов (`/uploads/pricelist.xml`) для Drom.ru и Avito.
   - **НАМЕРЕННО НЕ ИМЕЕТ СВОЕЙ БД**: считывает склад потоком через gRPC из `parts-service` (`ListPartsForExport`) и реквизиты продавцов (ИНН) из `auth-service` (`GetUsers`).
   - Интеллектуальный сборщик: перед пересборкой сверяет `InventoryVersion` отпечаток с `pricelist.meta.json`. Если склад не менялся — генерация пропускается.

### 3.2. Общие библиотеки (`backend/pkg/`)
- `pkg/authcontext` — стандартизированное извлечение контекста пользователя (`X-User-ID`, `X-User-Role`, `X-User-Name`, `X-User-Email`), проброшенного Traefik ForwardAuth.
- `pkg/httpcache` — вычисление `ETag` и быстрая обработка условных запросов `If-None-Match`.
- `pkg/httpserver` — запуск HTTP сервера с поддержкой Graceful Shutdown.
- `pkg/redisclient` — унифицированный клиент Redis и шины событий Redis Streams.
- `pkg/tracing` — инициализация OpenTelemetry трейсинга с экспортом в Grafana Tempo.
- `pkg/userdirectory` — потокобезопасный кэш справочника продавцов.

### 3.3. Правила межсервисной коммуникации и API
- **100% межсервисных вызовов — строго через gRPC (Protobuf):**
  - Никаких внутренних прямых HTTP-вызовов между сервисами.
  - Прямой доступ к БД соседнего сервиса строго запрещен.
  - Контракты описаны в `backend/proto/`. При любых правках контрактов запускать `make proto` (из каталога `backend/`).
- **Разделение маршрутов `/api/v1/*` и `/api/*`:**
  - `/api/v1/*` — строго типизированные эндпоинты, сгенерированные через **gRPC-Gateway** (принимают/возвращают JSON представления protobuf-сообщений).
  - `/api/*` — нативные REST эндпоинты Gin для задач, где Protobuf нецелесообразен:
    - Multipart-загрузка фото (`/api/uploadpartphoto/:id`, `/api/v1/uploadpartphoto/:id`).
    - Загрузка и парсинг Excel ведомостей (`/api/defect-reports`).
    - Статические каталоги (`/api/part-catalog`, `/api/vehicle-catalog`).
    - WebSocket подключение чата (`/api/messaging/ws`).

### 3.4. Базы данных и sqlc
- **Database-per-Service:** Каждый сервис работает со своей строкой подключения (`AUTH_DATABASE_URL`, `PARTS_DATABASE_URL`, `ORDERS_DATABASE_URL`, `MESSAGING_DATABASE_URL`) с fallback на `DATABASE_URL`.
- **Строгий запрет на тяжелые ORM:** Никакого GORM/Ent. Использовать связку **`sqlc` + `pgx/v5`**.
- SQL-запросы хранятся в `db/queries/*.sql`. При изменении запросов обязательно запускать `sqlc generate`.
- Таблица версионирования миграций: `schema_migrations`.
- **Идемпотентность складских операций:** Таблица `part_stock_operations` фиксирует `(operation_id, part_id, operation_type, amount)`. Повторный вызов с тем же `operation_id` игнорируется.

---

## 4. Критические доменные инварианты (Inventory & Orders)

### 4.1. Семантика поля `quantity` (КРИТИЧЕСКИ ВАЖНО)
В системе определены 3 фундаментальных состояния остатка деталей:
1. `quantity > 0` — обычная запчасть в наличии на складе.
2. `quantity == 0` — дефектная ведомость (акт дефектовки) или спецдеталь. **ОНИ ОБЯЗАНЫ ОСТАВАТЬСЯ ВИДИМЫМИ** в поиске, UI и XML-выгрузках.
3. `quantity == -1` — запчасть полностью продана через `orders-service`. Служит маркером сокрытия из UI и XML-выгрузок, сохраняя историчность и целостность заказов.

> ⚠️ **ПРАВИЛО ЗАПРОСОВ:** При выборке запчастей для отображения, полнотекстового поиска и генерации выгрузок ВСЕГДА использовать условие:
> ```sql
> WHERE quantity >= 0 -- ПРАВИЛЬНО! Включает нулевые дефектовки и отсекает проданные (-1)
> ```
> ❌ **НИКОГДА НЕ ПИСАТЬ:** `quantity > 0` (это скроет дефектные ведомости!).

### 4.2. Списание остатков в `orders-service`
```sql
-- DecreasePartQuantity:
UPDATE parts
SET quantity = CASE WHEN quantity - sqlc.arg(amount)::integer <= 0 THEN -1 ELSE quantity - sqlc.arg(amount)::integer END
WHERE id = sqlc.arg(id);

-- IncreasePartQuantity (при отмене заказа/возврате):
UPDATE parts
SET quantity = CASE WHEN quantity = -1 THEN sqlc.arg(amount)::integer ELSE quantity + sqlc.arg(amount)::integer END
WHERE id = sqlc.arg(id);
```

### 4.3. CRM, клиенты и дедупликация
- Поле `buyer_number` нормализуется с помощью алгоритма `parseBuyerInfo(raw)`:
  - Отсекаются служебные строки ("Самовывоз", "Продажа на месте", "Без контакта").
  - Номер телефона приводится к каноническому формату `+7XXXXXXXXXX`.
  - Из строки отсекаются лишние пометки (например, "Получатель: ...").
- На уровне БД действует ограничение `UNIQUE INDEX idx_customers_unique_phone`.
- При попытке создания существующего номера API возвращает `409 Conflict`.

---

## 5. Фронтенд (React 19 / Vite / TypeScript)

Фронтенд расположен в каталоге `frontend/`.

### 5.1. Архитектура и стек
- **Стек:** React 19, TypeScript, Vite, TanStack React Query v5, React Router v7, Tailwind CSS v4, Radix UI.
- **Модульная структура:**
  - `src/features/` — доменные модули: `auth`, `parts`, `orders`, `customers`, `defectReport`, `messaging`, `admin`, `catalog`, `vehicles`.
  - `src/components/ui/` — дизайн-система и примитивы shadcn/radix.
  - `src/lib/` — клиенты API (`api.ts`, `queryClient.ts`), безопасность (`csrf.ts`), телеметрия и кэш.
  - `src/docs/` — встроенный интерактивный Swagger/OpenAPI проводник `/readme`.

### 5.2. Требования к производительности
- **INP (Interaction to Next Paint):** Целевое значение **< 200ms**.
- **Параллельная загрузка фото:** Изображения в карточках и формах дефектовок должны обрабатываться и загружаться параллельно (`Promise.all`), а не блокирующей цепочкой.
- **Отказ от тяжелых анимаций:** Запрещены тяжелые `whileHover`/`whileTap` эффекты на карточках в длинных списках запчастей.
- **Debounced поиск:** Все поля ввода поиска обязаны использовать debounce во избежание флуда API и ре-рендеров.

---

## 6. Мобильное приложение (AvtoplanetaApp / Flutter)

Мобильное приложение расположено в каталоге `AvtoplanetaApp/`.

### 6.1. Архитектура и стек
- **Фреймворк:** Flutter 3.x / Dart (sdk `^3.11.4`).
- **State Management:** `flutter_riverpod` (v2.6+) + `riverpod_annotation`.
- **Навигация:** `go_router` (v15.1+), файл `lib/app/router.dart`.
- **Сетевой клиент:** `dio` (v5.8+) с логированием `pretty_dio_logger`.
- **Хранилище:** `flutter_secure_storage` (токены сессий) + `hive_flutter` (локальный кэш).
- **In-App APK Updates:** Автообновление через `package_info_plus`, `path_provider`, `open_filex`.
- **Сканер:** `mobile_scanner` для сканирования штрихкодов и QR-кодов запчастей.

### 6.2. Организация кода
- `lib/core/api/api_endpoints.dart` — **единый источник маршрутов API**. При добавлении эндпоинтов на бэкенде они обязательно синхронизируются здесь.
- `lib/core/models/` — модели данных: `Part`, `Order`, `Customer`, `User`, `Statistics`.
- `lib/features/` — фичи: `inventory`, `orders`, `auth`, `messaging`, `statistics`, `admin`, `updater`.

---

## 7. Справочник команд для агента

```bash
# === BACKEND ===
cd backend

# 1. Генерация Protobuf и gRPC-Gateway кода:
make proto

# 2. Генерация Go-кода SQL запросов (при изменении db/queries/*.sql):
sqlc generate

# 3. Запуск сервисов для локальной разработки:
make start-dev

# === FRONTEND ===
cd frontend

# Установка зависимостей и запуск dev-сервера:
npm install
npm run dev

# Проверка сборки и типов TypeScript:
npm run build

# === DOCKER / ИНФРАСТРУКТУРА ===
# Запуск всей инфраструктуры (Traefik, БД, Redis, ES, сервисы):
docker compose up -d

# Логи конкретного сервиса:
docker compose logs -f parts-service
```
