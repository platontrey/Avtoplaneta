# Документация по архитектуре проекта Avtoplaneta

## Обзор проекта

Проект Avtoplaneta представляет собой микросервисную архитектуру на языке Go для управления автозапчастями. Система включает в себя управление запасами, заказами и аутентификацией пользователей.

## Общая архитектура

```
┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   API Gateway   │
│   (React)       │◄──►│   (Gin)         │
│                 │    │   Port: 8080    │
└─────────────────┘    └───────┬─────────┘
                               │
           HTTP/REST           │ gRPC (inter-service)
           (frontend)          │
                     ┌─────────┼─────────────────┐
                     │         │         │       │
             ┌───────▼───┐ ┌───▼───┐ ┌───▼───┐ ┌─▼──────────┐
             │ Auth      │ │Orders │ │Parts  │ │Messaging   │
             │ Service   │ │Service│ │Service│ │Service     │
             │ :8083/9083│ │:8082/ │ │:8081/ │ │:8084/9084  │
             └───────────┘ │ 9082  │ │ 9081  │ └────────────┘
                     │     └───────┘ └───────┘       │
                     │         │         │           │
                     │ gRPC    │ Redis   │ Redis     │ HTTP
                     │◄────────┤ Streams │ Streams   │ (Drom.ru)
                     │         │         │           │
             ┌───────▼─────────▼─────────▼───────────▼──┐
             │              PostgreSQL                  │
             │              Redis                       │
             │              Elasticsearch (Parts only)  │
             └─────────────────────────────────────────┘
```

## Компоненты системы

### 1. Frontend (Пользовательский интерфейс)

**Технологии:** React/Vue.js + TypeScript/JavaScript
**Порты разработки:** 5173, 5174, 3000 (Vite, Create React App, Next.js)
**Развертывание:** Статические файлы, обслуживаемые веб-сервером

#### Основные функции:
- **Пользовательский интерфейс** для управления автозапчастями
- **Аутентификация пользователей** через Google OAuth и локальный вход
- **Управление инвентарем** - просмотр, добавление, редактирование запчастей
- **Управление заказами** - создание заказов, отслеживание статусов
- **Админ-панель** - управление пользователями, просмотр логов
- **Загрузка изображений** для запчастей
- **Поиск и фильтрация** с Elasticsearch интеграцией

#### Архитектура Frontend:
```
Frontend App
├── Components/
│   ├── Auth/           # Вход, регистрация, профиль
│   ├── Inventory/      # Список запчастей, формы добавления
│   ├── Orders/         # Заказы, статусы, создание
│   ├── Admin/          # Управление пользователями
│   └── Shared/         # Общие компоненты (таблицы, формы)
├── Services/
│   ├── api.js          # HTTP клиент для API
│   ├── auth.js         # Сервис аутентификации
│   └── websocket.js    # Real-time обновления (опционально)
├── Stores/             # State management (Redux, Zustand, Context)
├── Utils/              # Вспомогательные функции
└── Pages/              # Маршруты приложения
```

#### Интеграция с Backend:
- **CORS**: Настроен для локального развития (localhost:5173, 5174, 3000)
- **CSRF защита**: Получение и отправка CSRF токенов
- **Сессии**: Cookie-based аутентификация
- **API calls**: RESTful запросы к API Gateway
- **File uploads**: Multipart формы для изображений

#### Ключевые возможности:
- **Responsive дизайн** для мобильных устройств
- **Real-time поиск** с Elasticsearch
- **Drag & drop** для загрузки фото
- **Color-coded статусы** заказов (красный/коричневый/желтый/зеленый)
- **Role-based UI** - разные интерфейсы для admin/manager/operator

## gRPC: Межсервисная коммуникация

### Обзор

Межсервисная коммуникация (gateway ↔ services, service ↔ service) реализована через **gRPC** с Protocol Buffers вместо классического HTTP/JSON-проксирования. Это даёт:

- **HTTP/2 мультиплексирование** — одно TCP-соединение на все вызовы
- **Protobuf-сериализация** — компактнее и быстрее JSON (в 3-10 раз)
- **Строгая контрактная типизация** через `.proto` файлы
- **Streaming** для загрузки файлов (голосовые сообщения, фото)
- **Встроенный health check** протокол (`grpc.health.v1`)

### Proto-файлы и генерация

```bash
# Генерация Go-кода из .proto файлов
cd backend
make proto        # Требуется buf и protoc

# Структура:
backend/
├── proto/                # Исходные .proto файлы
│   ├── buf.yaml          # Конфигурация buf
│   ├── buf.gen.yaml      # Правила генерации
│   ├── auth/v1/auth.proto
│   ├── parts/v1/parts.proto
│   ├── orders/v1/orders.proto
│   └── messaging/v1/messaging.proto
└── gen/                  # Сгенерированный код
    ├── auth/v1/          # auth_grpc.pb.go + auth.pb.go
    ├── parts/v1/
    ├── orders/v1/
    └── messaging/v1/
```

### Порты сервисов

| Сервис | HTTP порт | gRPC порт | Назначение |
|--------|-----------|-----------|------------|
| Auth Service | 8083 | 9083 | Валидация сессий, CRUD пользователей, логи активности |
| Parts Service | 8081 | 9081 | Инвентарь запчастей, фото, статистика, экспорт |
| Orders Service | 8082 | 9082 | Заказы, продажи, статусы |
| Messaging Service | 8084 | 9084 | Чаты, сообщения, уведомления, Drom.ru |

### Текущие gRPC-вызовы

| От | Кому | RPC | Заменяет |
|----|------|-----|----------|
| Gateway | Auth Service | `ValidateSession` | HTTP `GET /auth/me` (на каждый запрос) |
| Gateway | Auth Service | `GetUsers` | HTTP `GET /internal/users` |
| Parts Service | Auth Service | `LogActivity` | HTTP `POST /internal/log-activity` |
| Parts Service | Orders Service | `GetMonthlySales` | HTTP `GET /monthly-sales` |

### Gradual migration strategy

Каждый сервис запускает **одновременно HTTP и gRPC серверы**. Это позволяет:
- Мигрировать вызовы поэтапно
- gRPC клиент пробует gRPC, при недоступности — HTTP fallback
- Frontend продолжает использовать HTTP/REST без изменений

### Добавление новых gRPC-методов

1. Добавить RPC в соответствующий `.proto` файл
2. Запустить `make proto` для регенерации кода
3. Реализовать метод в `grpc_server.go` нужного сервиса
4. Добавить gRPC клиент в gateway или другом сервисе

### 3. API Gateway (Корневой сервис)

**Файл:** `main.go`
**Порт:** 8080
**Технологии:** Gin Framework, gRPC, HTTP Proxy с resiliency patterns

#### Основные функции:
- **Проксирование запросов** к микросервисам
- **CORS middleware** для фронтенда
- **Безопасность:** CSP, HSTS, X-Frame-Options, CSRF защита
- **Логирование** всех запросов

#### Маршруты проксирования:

| Префикс | Сервис | Порт | Описание |
|---------|--------|------|----------|
| `/auth/*` | Auth Service | 8083 | Аутентификация и управление пользователями |
| `/admin/*` | Auth Service | 8083 | Админ-панель |
| `/api/inventory/*` | Parts Service | 8081 | Управление запасами |
| `/orders/*` | Orders Service | 8082 | Управление заказами |
| `/api/messaging/*` | Messaging Service | 8084 | Чаты и сообщения |
| `/uploads/*` | Parts Service | 8081 | Статические файлы изображений |
| `/api/users` | Auth Service | gRPC | Список пользователей (через gRPC) |

#### Аутентификация через gRPC

`authMiddleware` в Gateway использует gRPC-вызов `ValidateSession` к Auth Service вместо HTTP-запроса. Это горячий путь — вызывается на **каждый защищённый запрос**. Persistent gRPC-соединение (HTTP/2 мультиплексирование) снижает latency на этом вызове в 2-5 раз по сравнению с HTTP/1.1 (нет TCP handshake на каждый запрос).

В случае недоступности gRPC — автоматический fallback на HTTP `GET /auth/me`.

#### Middleware безопасности:
- **Content Security Policy (CSP)** - строгий в продакшене, разрешает eval в разработке
- **HSTS** - только в продакшене
- **X-Frame-Options: DENY** - защита от clickjacking
- **X-Content-Type-Options: nosniff** - защита от MIME sniffing
- **Referrer-Policy** - strict-origin-when-cross-origin
- **Permissions-Policy** - блокирует камеру, микрофон, геолокацию

### 4. Auth Service (Сервис аутентификации)

**Директория:** `auth-service/`
**Порт:** 8083
**Технологии:** Gin, sqlc, pgx/v5, PostgreSQL, Gorilla Sessions, Goth (Google OAuth)

#### Основные функции:
- **Аутентификация пользователей** (локальная + Google OAuth)
- **Управление сессиями** с безопасными cookie
- **Ролевая система** (admin, manager, operator)
- **CSRF защита** с токенами
- **Ограничение скорости** (rate limiting)
- **Управление пользователями** (CRUD операции)

#### Модели данных:

```go
type User struct {
    ID       int64  `json:"id"`
    Email    string `json:"email"`
    Name     string `json:"name"`
    Provider string `json:"provider"` // "google" или "local"
    Role     string `json:"role"`     // "admin", "manager", "operator"
    Password string `json:"-"`        // Хэшируется bcrypt
}
```

#### Ролевая система:
- **admin**: Полный доступ ко всем функциям, управление пользователями, серверные операции
- **manager**: Управление заказами (создание, просмотр, обновление статусов, удаление), просмотр инвентаря
- **operator**: CRUD операции с запчастями (создание, чтение, обновление, удаление), управление дефектными ведомостями, загрузка фото, статистика

#### Безопасность:
- **Сессии**: Cookie с HttpOnly, Secure, SameSite=Lax
- **Пароли**: bcrypt с DefaultCost
- **CSRF**: Токены с 24-часовым сроком действия
- **Rate limiting**: 10 попыток входа в минуту на IP
- **Логирование**: Все действия безопасности

### 5. Orders Service (Сервис заказов)

**Директория:** `orders-service/`
**Порт:** 8082
**Технологии:** Gin, sqlc, Squirrel, pgx/v5, PostgreSQL

#### Основные функции:
- **Создание заказов** с привязкой к частям
- **Управление статусами заказов** (цветовая кодировка)
- **Просмотр заказов** с пагинацией
- **Удаление заказов**
- **Интеграция с частями** (автоматическое помечение для удаления)

#### Модели данных:

```go
type Order struct {
    ID                 int64       `json:"id"`
    CustomerID         int64       `json:"customer_id"`
    SellerID           int64       `json:"seller_id"` // ID продавца
    Seller             string      `json:"seller"`    // Имя продавца
    Part               string      `json:"part"`      // Название детали
    PartID             int64       `json:"part_id"`   // ID детали
    Location           string      `json:"location"`  // Склад
    BuyerNumber        string      `json:"buyer_number"`
    Status             string      `json:"status"` // "red", "brown", "yellow", "green"
    StatusText         string      `json:"status_text"`
    CreatedAt          time.Time   `json:"created_at"`
    Items              []OrderItem `json:"items"`
}

type OrderItem struct {
    ID       int64   `json:"id"`
    OrderID  int64   `json:"order_id"`
    PartID   int64   `json:"part_id"`
    Quantity int     `json:"quantity"`
    Price    float64 `json:"price"`
}
```

#### Статусы заказов:
- **🔴 Red**: "Need to order transport company"
- **🟤 Brown**: "Waiting for response"
- **🟡 Yellow**: "Need to deliver"
- **🟢 Green**: "Transported"

#### Особенности:
- **Аутентификация**: Через middleware, проверка сессий
- **Форматирование дат**: CreatedAtFormatted, TimeAgo (относительное время)
- **Интеграция**: Получает имя пользователя напрямую из auth БД
- **Списание запасов**: При полном списании запасов детали по заказу (`quantity - amount <= 0`), `orders-service` присваивает детальке статус `quantity = -1`, убирая её из UI и XML-выгрузок.

### 6. Parts Service (Сервис запчастей)

**Директория:** `parts-service/`
**Порт:** 8081
**Технологии:** Gin, sqlc, Squirrel, pgx/v5, PostgreSQL, Elasticsearch

#### Основные функции:
- **Управление запасами** (CRUD операции с частями)
- **Полнотекстовый поиск и фильтрация** через Elasticsearch (`quantity >= 0`)
- **Единый каталог дефектовок и форм** (`catalog.json` как единый источник правды для ведомостей и динамических категорий)
- **Загрузка фотографий** с параллельной обработкой
- **Статистика** по категориям и стоимости
- **Автоматическое удаление** просроченных частей (14 дней)

#### Состояния количества (`quantity`):
- `quantity > 0`: Стандартная деталь в наличии на складе.
- `quantity == 0`: Дефектные ведомости / спецпозиции. Всегда отображаются в поиске и XML-выгрузках (`quantity >= 0`).
- `quantity == -1`: Полностью проданные и списанные позиции через `orders-service`. Скрыты от покупателей и выгрузок.

#### Модели данных:

```go
type Part struct {
    ID          int64       `json:"id"`
    Name        string      `json:"name"`
    Quantity    int         `json:"quantity"`
    Description string      `json:"description,omitempty"`
    Category    string      `json:"category,omitempty"`
    Price       float64     `json:"price,omitempty"`
    Salesman    string      `json:"salesman,omitempty"`
    Location    string      `json:"location,omitempty"`
    Status      bool        `json:"status,omitempty"`
    Brand       string      `json:"brand,omitempty"`
    Model       string      `json:"model,omitempty"`
    Photos      []string    `json:"photos,omitempty"`
    Photo       string      `json:"photo,omitempty"`
    SellerID    int64       `json:"seller_id,omitempty"`
    ToDeleteAt  *time.Time  `json:"to_delete_at,omitempty"`
}
```

#### Elasticsearch интеграция:
- **Индексация**: Все детали с `quantity >= 0` автоматически индексируются
- **Поиск**: Полнотекстовый поиск по названию и описанию
- **Фильтры**: По категории, бренду, модели, локации, продавцу, статусу
- **Релевантность**: Boost для названия, fuzzy search для опечаток

#### Функции управления частями:
- **addPart**: Создание новой части с валидацией
- **updatePart**: Обновление полей с индивидуальной обработкой типов
- **deletePart**: Удаление с очисткой файлов
- **markPartForDeletion**: Отметка для удаления через 14 дней
- **uploadPartPhoto**: Загрузка изображений (макс 5MB, только изображения)

#### Статистика:
- Общее количество частей (включая нулевые дефектовки)
- Общая стоимость (price * quantity)
- Распределение по категориям

## Хранилища данных и Сообщения

### 1. PostgreSQL (Основная реляционная БД)
**Все сервисы используют одну базовую БД** с быстрыми SQL-запросами через `sqlc` и драйвер `pgx/v5`:
- **users / sessions / activity_logs** — аккаунты, роли и логи аудита (`auth-service`).
- **orders / order_items** — заказы и связанные позиции товаров (`orders-service`).
- **parts / spec_bindings** — каталог инвентаря, артикулы и спецификации запчастей (`parts-service`).

### 2. Redis & Redis Streams (Кэш и Асинхронные события)
- **Redis Streams**:
  - Асинхронная гарантированная отправка и обработка дефектных ведомостей (события генерации до 1 400+ частей за дефектовку через Consumer Groups).
    Разворачивание ведомости по каталогу и публикация события живут в одном месте — `DefectReportWorkflow` (`parts-service/defect_report_workflow.go`).
    HTTP- и gRPC-эндпоинты являются только транспортными адаптерами и не создают запчасти напрямую: и `POST /api/defect-reports`, и `PartsService.CreateDefectReport`
    ставят ведомость в очередь, а записью в БД занимается consumer. Превью (`/api/defect-reports/preview`, `PartsService.PreviewDefectReport`) строит тот же набор, ничего не публикуя.
  - Фоновая шина событий между `orders-service` и `parts-service` для мгновенной рассылки сообщений о списании и обновлении количеств товаров.
- **Redis Caching**:
  - Кэширование активных сессий пользователей и токенов CSRF.
  - Кэш чатов, сообщений и временных диалогов (`messaging-service`).

### 2.1. Условные запросы и кэширование HTTP

`pkg/httpcache` — общий для всех сервисов разбор `If-None-Match` и выставление `ETag`.
Валидатор всегда вычисляется **до** сборки ответа, чтобы повторный запрос не стоил ничего:

- **Справочники** (`/api/part-catalog`, `/api/vehicle-catalog`) — валидатор это их собственная версия.
- **Агрегаты и выборки по складу** (`/api/statistics`, `/api/admin/supplier-codes`, прайс-лист) —
  валидатор `InventoryVersion`: пара «последнее изменение + число живых строк», один индексный
  запрос по `idx_parts_updated_at_alive`. Пара нужна потому, что жёсткое удаление не двигает
  `updated_at`, а мягкое не двигает максимум.
- **Фото** (`/uploads`) — кэшируются навсегда: имя файла содержит время загрузки, поэтому замена
  фотографии меняет адрес. Исключение — `pricelist.xml`, он живёт по постоянному адресу.
- **Персональные ответы** (`/auth/me`) помечены `private`, CSRF-токен — `no-store`.

Списки запчастей (`/api/inventory`) сознательно оставлены без ETag: валидатор по отфильтрованной
выборке стоил бы столько же, сколько сам запрос. Там работают пагинация и клиентский кэш.

### 3. Elasticsearch (Высокопроизводительный полнотекстовый поиск)
- **Индекс:** `parts`
- **Функции:**
  - Полнотекстовый поиск по 100k+ запчастям (по названию, артикулам, OEM-кодам, VIN-кодам и описаниям).
  - Динамическая фильтрация по категориям, брендам, моделям, состоянию, расположению складов и продавцам.
  - Поддержка `fuzzy search` для опечаток и весовых коэффициентов (boost для точных совпадений названий).
  - Индексируются только позиции с `quantity >= 0` (включая товары с дефектовок `quantity == 0` и исключая проданные `quantity == -1`).

## Безопасность

### Уровни безопасности:
1. **API Gateway**: CORS, CSP, HSTS, заголовки безопасности
2. **Auth Service**: Сессии, CSRF, rate limiting, bcrypt пароли
3. **Application Level**: Ролевая система, middleware аутентификации

### Аутентификация:
- **Сессии**: Cookie-based с безопасными настройками
- **OAuth**: Google OAuth интеграция
- **Локальная**: Email + пароль с хэшированием

### Авторизация:
- **Роли**: admin > manager > operator
- **Middleware**: Проверка ролей для защищенных маршрутов
- **CSRF**: Защита от cross-site request forgery

## Запуск и развертывание

### Требования:
- Go 1.21+
- PostgreSQL
- Elasticsearch (опционально, fallback на PostgreSQL)
- Redis (опционально для продакшен CSRF)

### Переменные окружения:
```bash
# База данных
DATABASE_URL=postgres://user:pass@localhost:5432/autoplanet

# OAuth Google
GOOGLE_CLIENT_ID=your_client_id
GOOGLE_CLIENT_SECRET=your_client_secret
GOOGLE_CALLBACK_URL=http://localhost:8082/auth/google/callback

# Сессии
SESSION_SECRET=secure_random_key_32_chars_min

# Окружение
NODE_ENV=production  # или development
ALLOWED_ORIGIN=https://yourdomain.com
```

### Запуск сервисов:
```bash
# API Gateway
go run main.go

# Auth Service
cd auth-service && go run .

# Orders Service
cd orders-service && go run .

# Parts Service
cd parts-service && go run .
```

## API Endpoints

### Auth Service
- `POST /auth/login` - Вход пользователя
- `GET /auth/google` - OAuth Google
- `GET /auth/me` - Текущий пользователь
- `GET /auth/csrf-token` - Получить CSRF токен
- `POST /auth/logout` - Выход

### Admin (Auth Service)
- `GET /admin/users` - Список пользователей
- `POST /admin/users` - Создать пользователя
- `DELETE /admin/users/:id` - Удалить пользователя
- `GET /admin/status` - Статус сервера
- `GET /admin/logs` - Логи сервера

### Orders Service
- `GET /orders` - Получить заказы
- `POST /orders` - Создать заказ
- `PUT /admin/orders/:id/status` - Обновить статус
- `DELETE /admin/orders/:id` - Удалить заказ

### Parts Service
- `GET /api/inventory` - Получить инвентарь
- `POST /api/addpart` - Добавить часть
- `PUT /api/updatepart/:id` - Обновить часть
- `DELETE /api/deletepart/:id` - Удалить часть
- `POST /api/uploadpartphoto/:id` - Загрузить фото
- `POST /api/markpartfordeletion/:id` - Отметить для удаления
- `GET /api/statistics` - Статистика

## Мониторинг и логирование

### Логирование:
- **API Gateway**: Все HTTP запросы с IP и методом
- **Auth Service**: Все действия безопасности (входы, выходы, создание пользователей)
- **Services**: Ошибки и важные операции

### Мониторинг:
- **Статус сервера**: `/admin/status`
- **Статистика**: `/api/statistics`
- **Логи**: `/admin/logs`

## Производительность

### Оптимизации:
- **Elasticsearch**: Для быстрого поиска в больших объемах данных
- **Индексы БД**: Создаются во время SQL-миграций (`db/migrations/`) и выполняются через embedded миграции при запуске сервисов.
- **Пагинация**: Во всех списочных запросах
- **Кэширование**: CSRF токены (в памяти, Redis в продакшене)

### Масштабируемость:
- **Микросервисы**: Независимое масштабирование каждого сервиса
- **База данных**: Общая БД с разделением по таблицам
- **Elasticsearch**: Отдельный кластер для поиска

## Разработка и тестирование

### 7. Messaging Service (Сервис сообщений)

**Директория:** `messaging-service/`
**Порты:** HTTP 8084, gRPC 9084
**Технологии:** Gin, sqlc, Squirrel, pgx/v5, PostgreSQL, Redis, goquery (парсинг Drom.ru)

#### Основные функции:
- **Чаты/диалоги** между пользователями системы
- **Текстовые и голосовые сообщения** (gRPC streaming для voice upload)
- **Реакции** (emoji) на сообщения
- **Уведомления** о новых сообщениях
- **Поиск** по истории сообщений
- **Интеграция с Drom.ru** — парсинг и отправка сообщений через Drom API
- **Кэширование** сообщений в Redis

#### gRPC методы (21 RPC):
- Conversations: CRUD + RemoveParticipant
- Messages: Get/Send/Delete/MarkRead + Streaming Voice Upload
- Reactions: Add/Remove
- Notifications: Get/MarkRead/MarkAllRead
- User Status: Get/Update
- Search: Full-text search
- Drom: GetDialogs/GetMessages/SendMessage

### Структура проекта (gRPC):
```
backend/
├── main.go                 # API Gateway
├── gateway.go              # gRPC-клиенты, маршрутизация, middleware
├── proto/                  # Proto-определения (.proto)
│   ├── buf.yaml            # buf конфигурация
│   ├── buf.gen.yaml        # правила генерации
│   ├── auth/v1/auth.proto
│   ├── parts/v1/parts.proto
│   ├── orders/v1/orders.proto
│   └── messaging/v1/messaging.proto
├── gen/                    # Сгенерированный Go-код из proto
│   ├── auth/v1/
│   ├── parts/v1/
│   ├── orders/v1/
│   └── messaging/v1/
├── auth-service/
│   ├── main.go            # HTTP + gRPC сервер
│   ├── grpc_server.go     # gRPC-реализация AuthService
│   ├── models.go
│   ├── handlers.go
│   └── jwt.go
├── orders-service/
│   ├── main.go            # HTTP + gRPC сервер
│   ├── grpc_server.go     # gRPC-реализация OrdersService
│   ├── models.go
│   └── handlers.go
├── parts-service/
│   ├── main.go            # HTTP + gRPC сервер
│   ├── grpc_server.go     # gRPC-реализация PartsService
│   ├── grpc_client.go     # gRPC клиенты к auth и orders
│   ├── models.go
│   └── handlers.go
└── messaging-service/
    ├── main.go            # HTTP + gRPC сервер
    ├── grpc_server.go     # gRPC-реализация MessagingService
    ├── models.go
    └── routes.go
```

### Лучшие практики:
- **Валидация входных данных** во всех endpoints
- **Обработка ошибок** с понятными сообщениями
- **Логирование** всех важных операций
- **Безопасность** на всех уровнях
- **Тестирование** API endpoints

## Заключение

Архитектура проекта Avtoplaneta демонстрирует современный подход к разработке микросервисных приложений с акцентом на безопасность, производительность и масштабируемость. Каждый сервис имеет четко определенные обязанности, что облегчает поддержку и развитие системы.