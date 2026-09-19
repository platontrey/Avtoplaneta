# Документация по архитектуре проекта Avtoplaneta

## Обзор проекта

Проект Avtoplaneta представляет собой микросервисную архитектуру на языке Go для управления автозапчастями. Система включает в себя управление запасами, заказами и аутентификацией пользователей.

## Общая архитектура

```
┌─────────────────┐       ┌─────────────────────────────────────────────────────────┐
│   Frontend      │◄─────►│ Traefik Ingress (SSL/TLS, Routing, ForwardAuth Verify)  │
│   (React)       │       └───────────┬──────────────┬─────────────┬────────────┬───┘
└─────────────────┘                   │              │             │            │
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
                              │ Redis (Кэш, Сессии, Redis Streams)                 │
                              │ Elasticsearch (Полнотекстовый поиск деталей)        │
                              └─────────────────────────────────────────────────────┘
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
| Parts Service | 8081 | 9081 | Инвентарь запчастей, фото, статистика |
| Orders Service | 8082 | 9082 | Заказы, продажи, статусы |
| Messaging Service | 8084 | 9084 | Чаты, сообщения, уведомления, Drom.ru |
| Export Service | 8085 | — | Прайс-лист для Drom.ru |

### Текущие gRPC-вызовы (100% межсервисная коммуникация)

| От | Кому | RPC | Назначение |
|----|------|-----|------------|
| Orders Service | Parts Service | `ChangePartQuantity` | Идемпотентное списание/возврат остатков склада с фиксацией в `part_stock_operations` |
| Orders Service | Parts Service | `GetPart` | Получение актуальной информации о детали при оформлении заказа |
| Orders Service | Auth Service | `ValidateSession` | Валидация сессии при прямых операциях с заказами |
| Parts Service | Auth Service | `LogActivity` | Запись аудита действий операторов в лог активности |
| Parts Service | Orders Service | `GetMonthlySales` | Аналитика продаж за месяц для панели управления |
| Parts Service | Auth Service | `GetUsers` | Получение списка пользователей (продавцов) для карточек деталей |
| Messaging Service | Auth Service | `GetUsers` | Получение списка пользователей для адресной книги чата |
| Export Service | Parts Service | `ListPartsForExport`, `InventoryVersion` | Потоковая выгрузка склада для формирования прайс-листа Drom.ru без прямого доступа к БД склада |
| Export Service | Auth Service | `GetUsers` | ИНН и реквизиты продавцов для прайс-листа |

Все межсервисные вызовы работают строго по **gRPC** с автоматическим HTTP-fallback при сбоях.

### 3. Маршрутизация и Безопасность (Traefik Ingress & ForwardAuth)

В проекте **ликвидирован рукописный API Gateway**. Его функции перенесены на промышленный edge-прокси **Traefik**:

1. **Traefik Ingress (TLS / Routing):**
   - Прямое проксирование клиентского трафика в микросервисы без накладных расходов и промежуточных сериализаций.
   - Сжатие ответов (Brotli / Gzip) через middleware `compress-res`.
   - Автоматический выпуск и продление SSL-сертификатов Let's Encrypt.
2. **Traefik ForwardAuth (`auth-service /auth/verify`):**
   - Защищённые роуты перед передачей в бэкенд вызывают внутренний эндпоинт `auth-service:8083/auth/verify` (или `/auth/verify-admin`).
   - Если сессия валидна, `auth-service` возвращает `200 OK` и заголовки `X-User-ID`, `X-User-Role`, `X-User-Name`, `X-User-Email`, которые Traefik автоматически пробрасывает в целевой микросервис.
   - Если сессия недействительна, Traefik мгновенно блокирует запрос с `401 Unauthorized` / `403 Forbidden`.

#### Маршруты проксирования:

| Префикс | Сервис | Тип / Middleware | Описание |
|---------|--------|------------------|----------|
| `/auth/*` | Auth Service (:8083) | Публичный / Сессионный | Вход, регистрация, OAuth, сессии |
| `/admin/*` | Auth Service (:8083) | `auth-forward-admin` | Управление пользователями, журнал аудита |
| `/api/v1/inventory`, `/api/v1/statistics` | Parts Service (:8081) | Публичный (кэш/ETag) | gRPC-Gateway инвентаря и статистики склада |
| `/api/v1/parts/*`, `/api/v1/admin/*` | Parts Service (:8081) | `auth-forward` | gRPC-Gateway CRUD деталей, пакетное удаление/обновление |
| `/api/inventory`, `/api/part-catalog`, `/api/vehicle-catalog` | Parts Service (:8081) | Публичный | Нативные REST каталоги и справочники |
| `/api/addpart`, `/api/uploadpartphoto/*`, `/api/defect-reports` | Parts Service (:8081) | `auth-forward` | Загрузка фото, дефектовки, создание деталей |
| `/orders/*` | Orders Service (:8082) | `auth-forward` | Создание и управление заказами |
| `/api/messaging/*` | Messaging Service (:8084) | `auth-forward` | Внутренние чаты, сообщения, интеграция Drom |
| `/uploads/*` | Parts Service (:8081) | Статика (кэширование) | Фотографии деталей |
| `/uploads/pricelist.xml` | Export Service (:8085) | Публичный | Готовый XML-файл прайса для Drom |
| `/api/export/*` | Export Service (:8085) | `auth-forward-admin` | Принудительная сборка и синхронизация XML |

#### Почему сосуществуют `/api/` и `/api/v1/`:
- **`/api/v1/*` (Protobuf gRPC-Gateway):** Строго типизированные методы спецификации `proto/parts/v1/parts.proto`. Сгенерированы через `protoc-gen-grpc-gateway` и принимают/возвращают Protobuf JSON структуры (`partsApi.ts`, `Statistics.tsx`, `BulkDeleteDialog.tsx`).
- **`/api/*` (Нативный REST Gin):** Специализированные HTTP-обработчики, которые нецелесообразно заворачивать в Protobuf:
  - Мультипарт-загрузка бинарных файлов и изображений деталей (`/api/uploadpartphoto/:id`).
  - Загрузка и парсинг Excel-файлов дефектовочных ведомостей (`/api/defect-reports`, `/api/defect-reports/preview`).
  - Статические каталоги авто и деталей (`/api/vehicle-catalog`, `/api/part-catalog`).
  - REST-эндпоинты чатов (`/api/messaging/*`).

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

Выгрузка прайс-листа сюда не входит: ею занимается `export-service`. Склад не знает ни про Drom, ни про ИНН продавцов.

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

### 8. Export Service (Сервис выгрузки)

**Директория:** `export-service/`
**Порт:** 8085
**Технологии:** Gin, gRPC-клиенты

Собирает прайс-лист для площадки Drom.ru: `/uploads/pricelist.xml`.

**Своей базы данных у сервиса нет намеренно.** Он ничем не владеет — запчасти
читает потоком из `parts-service` (`ListPartsForExport`), реквизиты продавцов из
`auth-service` (`GetUsers`). Данные остаются у тех, кто ими владеет, а здесь
только формат выгрузки.

Раньше всё это жило внутри `parts-service`, и тот ради одного поля в XML ходил в
`auth-service` за ИНН — за персональными данными сотрудников, которые складу не
нужны ни для чего. Теперь ИНН знает только тот, кто его печатает.

#### Когда пересобирается файл:
Прайс-лист — это полный проход по складу. Чтобы не делать его на каждый запрос,
рядом с файлом лежит `pricelist.meta.json` с отпечатком склада. Перед сборкой
сервис спрашивает у `parts-service` текущий отпечаток (`InventoryVersion` — одна
агрегирующая строка) и, если он совпал, не трогает ни склад, ни файл.

Планировщик раз в час задаёт тот же вопрос. Отдельного хранилища для расписания
не нужно: состояние — это сам файл и его отпечаток.

#### Маршруты:
- `GET /uploads/pricelist.xml` — сам файл, анонимно (его забирает Drom)
- `GET /api/export/xml` — пересобрать, если склад менялся (admin)
- `POST /api/export/drom` — собрать и отправить на API площадки (admin)

## Хранилища данных и Сообщения

### 1. PostgreSQL (Изоляция Database-per-Service и Идемпотентность)
Микросервисы спроектированы под архитектурный паттерн **Database-per-Service**:
- **Изолированные БД**: Каждый сервис может работать с отдельной БД (`AUTH_DATABASE_URL`, `PARTS_DATABASE_URL`, `ORDERS_DATABASE_URL`, `MESSAGING_DATABASE_URL`). Для локальной разработки и обратной совместимости предусмотрен единый fallback на `DATABASE_URL`.
- **Строгая типизация SQL (`sqlc`)**: Все статичные SQL-запросы вынесены в файлы `.sql` и компилируются в типобезопасный Go-код с драйвером `pgx/v5`.
- **Идемпотентность остатков склада (`part_stock_operations`)**: Операции изменения остатков защищены на уровне БД. Каждая операция сопровождается записью `(operation_id, part_id, operation_type)` с уникальным ключом, что исключает повторное списание при сетевых ретраях.

### 2. Redis & Redis Streams (Кэш и Асинхронные события)
- **Redis Streams (Шина доменных событий)**:
  - **Синхронизация продавцов (`seller_renamed`)**: При изменении данных пользователя `auth-service` публикует событие в стрим. Сервисы `parts-service` и `orders-service` получают событие и атомарно обновляют реквизиты в карточках и заказах без периодических тяжёлых фоновых опросов.
- **Пакетная обработка дефектовок (без оверхеда очередей)**:
  - Генерация позиций из дефектовочных ведомостей (до 1 400+ запчастей за раз) переведена на синхронный `pgx.Batch` и пакетную индексацию Elasticsearch `BulkIndexParts`. Это снизило время обработки с 10–20 секунд до 30–50 миллисекунд и устранило паразитные очереди.
- **Redis Caching**:
  - Кэширование активных сессий пользователей и токенов CSRF (`auth-service`).
  - Кэш чатов, сообщений и временных диалогов (`messaging-service`).

### 2.1. Условные запросы и кэширование HTTP

`pkg/httpcache` — общий для всех сервисов разбор `If-None-Match` и выставление `ETag`.
Валидатор всегда вычисляется **до** сборки ответа, чтобы повторный запрос не стоил ничего:

- **Справочники** (`/api/part-catalog`, `/api/vehicle-catalog`) — валидатор это их собственная версия.
- **Агрегаты и выборки по складу** (`/api/v1/statistics`, `/api/admin/supplier-codes`, прайс-лист) —
  валидатор `InventoryVersion`: пара «последнее изменение + число живых строк», один индексный
  запрос по `idx_parts_updated_at_alive`.
- **Фото** (`/uploads`) — кэшируются навсегда: имя файла содержит время загрузки, поэтому замена
  фотографии меняет адрес. Исключение — `pricelist.xml`, он живёт по постоянному адресу.
- **Персональные ответы** (`/auth/me`) помечены `private`, CSRF-токен — `no-store`.

Списки запчастей (`/api/v1/inventory`) сознательно оставлены без ETag: валидатор по отфильтрованной
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
1. **Traefik Ingress + ForwardAuth**: TLS/SSL termination, HSTS, централизованная проверка сессий на `/auth/verify`, проброс проверенных заголовков `X-User-*`
2. **Auth Service**: Сессии, CSRF, rate limiting, bcrypt пароли, Google OAuth
3. **Application Level**: Ролевая система (admin, manager, operator), middleware ролей внутри микросервисов

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