# Avtoplaneta - Система управления автозапчастями

## О проекте

Avtoplaneta - это комплексная система управления инвентарем автозапчастей, разработанная с использованием современных технологий. Система включает веб-приложение, мобильное приложение и микросервисную backend архитектуру для эффективного управления запасами, заказами и пользователями.

## 🏛️ Архитектура проекта и сетевая модель

Система спроектирована по стандартам промышленной микросервисной архитектуры с жестким разделением на **два сетевых контура**:

```
                       ВНЕШНИЙ КОНТУР (Интернет / Клиенты)
   React (Web SPA)     Flutter (Mobile)      Drom.ru (Робот)      Prometheus (Scraper)
        │                     │                     │                      │
        │ HTTPS/JSON          │ HTTPS/JSON          │ GET /pricelist.xml   │ GET /metrics
        ▼                     ▼                     ▼                      ▼
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                      Traefik Edge API Gateway (Порты :80, :443)                        │
│  • TLS Termination (Let's Encrypt)              • Dynamic Routing (Host/PathPrefix)   │
│  • ForwardAuth (/auth/verify в auth-service)     • Rate Limiting (100 req/min)         │
│  • Response Compression (Brotli & Gzip)         • Security Headers (HSTS, XSS, Frame) │
└─────────────────────────────────────────┬──────────────────────────────────────────────┘
                                          │ HTTP / JSON / Multipart
                                          ▼
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                   ФАСАДНЫЕ HTTP-СЕРВЕРЫ МИКРОСЕРВИСОВ (Порты 8081–8085)                 │
│  auth-service (:8083)  parts-service (:8081)  orders-service (:8082) export-service(:8085) │
└───────────────────────────────┬─────────────────────────┬──────────────────────────────┘
                                │                         │
════════════════════════════════╪═════════════════════════╪══════════════════════════════
                                │                         │
                                ▼                         ▼
┌────────────────────────────────────────────────────────────────────────────────────────┐
│               ВНУТРЕННИЙ МЕЖСЕРВИСНЫЙ КОНТУР (gRPC HTTP/2, Порты 9081–9084)            │
│  • 100% прямое межсервисное общение без посредников     • Строгие контракты Protobuf   │
│  • Мультиплексирование потоков в одном TCP-соединении   • gRPC Streaming данных        │
└────────────────────────────────────────────────────────────────────────────────────────┘
```

### Основные компоненты:

#### 🖥️ Frontend (Веб-приложение)
- **Стек**: React 19, TypeScript, Vite, Tailwind CSS v4, Nginx (Brotli & Gzip)
- **Расположение**: `frontend/` (документация: [frontend/README.md](frontend/README.md))

#### 📱 Mobile App (iOS / Android)
- **Стек**: Flutter, Dart, Android SDK, iOS SDK
- **Расположение**: `AvtoplanetaApp/` (документация: [AvtoplanetaApp/README.md](AvtoplanetaApp/README.md))

#### 🔧 Backend (Микросервисы на Go)
- **Служебный шлюз Traefik** (HTTP :80, HTTPS :443) — единый Edge API Gateway, TLS-терминация, централизованная аутентификация через ForwardAuth (`/auth/verify`), rate limiting.
- **Auth Service** (HTTP :8083, gRPC :9083) — аутентификация, ForwardAuth верификация сессий (<1 мс), роли (Admin, Manager, Operator), gRPC `ValidateSession` / `GetUsers`.
- **Parts Service** (HTTP :8081, gRPC :9081) — складской инвентарь, загрузка фото, шаблоны каталога (`catalog.json`), CQRS с Elasticsearch, Transactional Outbox.
- **Orders Service** (HTTP :8082, gRPC :9082) — оформление заказов, клиентская база с нормализацией телефонов, атомарное списание остатков (`quantity = -1`) через gRPC.
- **Messaging Service** (HTTP :8084, gRPC :9084) — чаты, WebSocket/polling сообщения, уведомления.
- **Export Service** (HTTP :8085) — генерация и отдача XML-каталогов для классифайдов (Drom.ru) с инкрементальным слепком на базе `bbolt` и фоновой очередью Redis Streams.

---

## 🧩 Ключевые архитектурные паттерны (Senior Architecture Patterns)

В проекте строго соблюдаются архитектурные паттерны корпоративного уровня:

1. **Edge API Gateway & ForwardAuth (Traefik):**
   - Ни один микросервис не выставлен в интернет напрямую (порты закрыты внутри `avtoplaneta_network`).
   - Все защищенные операции (`/api/addpart`, `/api/updatepart`, `/admin/*`) перехватываются Traefik и валидируются в `auth-service` через подзапрос `/auth/verify`. При успехе Traefik прозрачно инжектирует в заголовки сервиса доверенные `X-User-ID`, `X-User-Role`, `X-User-Email`.

2. **Двухконтурная модель коммуникации (Dual Engine: HTTP + gRPC):**
   - **Фасадный контур (HTTP/REST):** Микросервисы поднимают легковесный HTTP-сервер (`pkg/httpserver` на Gin) для внешних клиентов — React, Flutter, мобильных приложений, робота Drom.ru (`GET /uploads/pricelist.xml`) и загрузки фото (`multipart/form-data`).
   - **Межсервисный контур (gRPC HTTP/2):** Общение между бэкендами (`service ↔ service`) на 100% выполняется по gRPC с бинарным Protobuf без накладных расходов на JSON, с мультиплексированием и строгими контрактами (`backend/proto/`). Межсервисный REST запрещен.
   - **Bigint Standard & Clean Mappers:** Первичные ключи в Protobuf строго выровнены со схемой PostgreSQL (`BIGSERIAL`/`BIGINT`) как `int64`. Доменные модели реализуют методы `ToProto()` / `FromProto()`, а доступ к полям запросов осуществляется через безопасные геттеры `.Get...()`, что исключает раздувание кода и нагромождение кастов `uint32(...)` / `int64(...)`.

3. **Изоляция данных: Database-per-Service:**
   - Каждый микросервис изолированно владеет своей схемой БД (`AUTH_DATABASE_URL`, `PARTS_DATABASE_URL`, `ORDERS_DATABASE_URL`, `MESSAGING_DATABASE_URL`).
   - Прямой доступ к чужим базам данных исключен — интеграция происходит только через публичные gRPC контракты.

4. **Гарантированная синхронизация: Transactional Outbox Pattern (No Dual-Write):**
   - Категорический запрет на двойную запись (Dual-Write) в PostgreSQL и Elasticsearch из HTTP-хэндлеров.
   - Изменение сущности и запись события в таблицу `outbox_events` происходят в рамках **одной транзакции PostgreSQL** (`pgx.Tx`).
   - Триггер БД моментально будит воркер через `PERFORM pg_notify('outbox_events_channel', ...)`, который асинхронно обновляет поисковый индекс Elasticsearch и кэш Redis.

5. **CQRS (Command Query Responsibility Segregation):**
   - **Write Model:** PostgreSQL (pgx/v5, sqlc) — единственный Source of Truth с поддержкой строгих ACID-транзакций.
   - **Read Model:** Elasticsearch — денормализованная поисковая проекция со стеммингом русского языка, опечатками и синонимами (`synonyms.txt`). При отказе Elasticsearch сервис автоматически деградирует на fallback-поиск через PostgreSQL (`ILIKE`).
   - **Кэширование витрины:** Redis с поддержкой ETag / `If-None-Match` для мгновенных ответов `304 Not Modified`.

6. **Асинхронные события: Event-Driven Architecture (Redis Streams):**
   - Слабосвязанная доставка событий между сервисами (`seller_renamed`, события смены статусов заказов) с гарантией At-Least-Once через Consumer Groups (`XREADGROUP`, `XACK`, `XAUTOCLAIM`).

7. **Инкрементальный слепок каталога (Incremental Catalog Snapshot):**
   - В `export-service` реализован встроенный ACID-слепок на `bbolt` (`/app/uploads/catalog.db`).
   - При изменении детали через Redis Streams точечно обновляется ровно один XML-кусочек.
   - Полный XML-каталог на 50 000 позиций собирается за **0.4 мс** линейным сбросом готовых байт на диск без единого запроса в PostgreSQL и без нагрузки на Go GC.

8. **Идемпотентность списаний (Idempotent Consumer):**
   - Списание складских остатков по заказу фиксируется с уникальным `operation_id` в таблице `part_stock_operations`, предотвращая повторные списания при сетевых сбоях.

9. **Чистая слоистая архитектура (Clean Layered Architecture):**
   - Микросервисы структурированы по доменным слоям:
     - `internal/domain` — чистые сущности бизнеса.
     - `internal/repository` — слой персистентности (sqlc, pgx/v5, bbolt).
     - `internal/service` — сценарии и бизнес-логика.
     - `internal/events` — очереди и стримы событий.
     - `internal/grpc` — gRPC-клиенты и серверы.
     - `internal/handler` — HTTP-хэндлеры и роуты.
   - Общий код переиспользуется через `backend/pkg` (`httpserver`, `redisclient`, `authcontext`, `httpcache`, `tracing`).

10. **Сквозная наблюдаемость (Distributed Tracing & Metrics):**
    - Распределенный трейсинг по стандарту W3C Trace Context через **OpenTelemetry** и **Grafana Tempo**.
    - Сбор системных и бизнес-метрик через **Prometheus** и дашборды **Grafana**.

---

## 📦 Инвентарь и статус количества (Quantity States)

В системе действует строгий доменный инвариант семантики поля `quantity`:
- `quantity > 0`: Запчасти в наличии на складе.
- `quantity == 0`: Дефектные ведомости или спецпозиции. Они **ОБЯЗАНЫ ОСТАВАТЬСЯ ВИДИМЫМИ** в поиске, UI и XML-выгрузках для классифайдов (фильтр: `WHERE quantity >= 0`).
- `quantity == -1`: Полностью списанные и проданные через `orders-service` детали. Служит маркером сокрытия из активной выдачи с сохранением историчности.

## Производительность и оптимизации

### 🚀 Оптимизации производительности

Проект оптимизирован для высокой производительности веб-приложения:

- **INP (Interaction to Next Paint)**: <200ms для всех взаимодействий
- **Параллельная загрузка ресурсов**: Фото загружаются параллельно для сокращения времени блокировки
- **Оптимизированные анимации**: Минимизированы тяжелые CSS-анимации для улучшения рендеринга
- **Debounced поиск**: Предотвращает избыточные API-запросы при вводе текста

### Последние оптимизации:
- **Параллельная загрузка фото**: Изменена последовательная загрузка на параллельную в AddPart компоненте
- **Убраны тяжелые анимации**: Удалены whileHover/whileTap эффекты в PartBlock для снижения INP
- **Исправлены логи производительности**: Добавлены таймеры для измерения времени выполнения обработчиков

## Основные возможности

### 👥 Управление пользователями
- Ролевая система: Администратор, Менеджер, Оператор
- Аутентификация через локальную систему или Google OAuth
- **Права доступа:**
  - **Администратор**: Полный доступ ко всем функциям, управление пользователями
  - **Менеджер**: Управление заказами, просмотр инвентаря
  - **Оператор**: CRUD операции с запчастями (создание, чтение, обновление, удаление), управление дефектными ведомостями
- Детальная документация ролей: [frontend/src/components/Readme.tsx](frontend/src/components/Readme.tsx)

### 🔧 Управление запчастями
- CRUD операции с автозапчастями
- Загрузка и управление фотографиями
- Полнотекстовый поиск с Elasticsearch
- Категоризация и фильтрация
- Каталог шаблонов дефектовок и форм (`catalog.json`)

### 📋 Управление заказами
- Создание и отслеживание заказов
- Цветовая кодировка статусов (красный/коричневый/желтый/зеленый)
- Автоматическое списание запчастей (установка статуса `quantity = -1`)

### 📊 Аналитика и отчеты
- Статистика по категориям запчастей
- Мониторинг системы (Prometheus, Grafana, Tempo)
- Логирование действий пользователей

## Быстрый старт

### Требования
- Docker и Docker Compose (рекомендуется)
- Или: Go 1.21+, Node.js 18+, Flutter 3.x+, PostgreSQL, Elasticsearch

### Настройка API ключа для ИИ

**Получите API ключ OpenRouter:**
1. Перейдите на [OpenRouter.ai](https://openrouter.ai/keys)
2. Зарегистрируйтесь и получите бесплатный API ключ
3. Добавьте ключ в `backend/.env`:

```env
OPENROUTER_API_KEY=sk-or-v1-xxxxxxxxxxxxxxxxxx
```

### Запуск с Docker (рекомендуется)
```bash
# Склонировать репозиторий
git clone <repository-url>
cd avtoplaneta

# Настройте переменные окружения для Traefik (DOMAIN, ACME_EMAIL, DASHBOARD_USER, DASHBOARD_PASSWORD и т.д.)
# Подробности: TRAEFIK-README.md

# Запустить все сервисы (включая Traefik)
docker compose up -d

# Или для разработки
docker compose -f docker-compose.dev.yml up
```

### Ручная установка

#### Backend
```bash
cd backend

# Генерация gRPC-кода и OpenAPI спецификации из proto-файлов (требуется buf)
make proto

# Установка всех Go-зависимостей микросервисов и pkg
make deps

# Настройте переменные окружения в .env файле
cp .env.example .env
# Отредактируйте .env файл с вашими настройками

# Запуск микросервисов для локальной разработки:
make start-dev
# Либо запуск отдельных сервисов:
# cd auth-service && go run .
# cd parts-service && go run .
# cd orders-service && go run .
# cd messaging-service && go run .
# cd export-service && go run .
```

#### Frontend
```bash
cd frontend
npm install
npm run dev
```

#### Mobile App
```bash
cd AvtoplanetaApp
# Открыть в Android Studio и запустить
```

## Конфигурация

### Переменные окружения
Создайте `.env` файлы в соответствующих директориях:

**Backend (.env):**
```env
DATABASE_URL=postgres://user:pass@localhost:5432/avtoplaneta
AUTH_DATABASE_URL=postgres://user:pass@localhost:5432/avtoplaneta_auth
PARTS_DATABASE_URL=postgres://user:pass@localhost:5432/avtoplaneta_parts
ORDERS_DATABASE_URL=postgres://user:pass@localhost:5432/avtoplaneta_orders
MESSAGING_DATABASE_URL=postgres://user:pass@localhost:5432/avtoplaneta_messaging
REDIS_URL=redis://localhost:6379
ELASTICSEARCH_URL=http://localhost:9200
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
SESSION_SECRET=secure_random_key_min_32_chars
OPENROUTER_API_KEY=sk-or-v1-your-openrouter-key
```

**Frontend (.env):**
```env
VITE_API_BASE_URL=http://localhost:8080
VITE_GOOGLE_CLIENT_ID=your_google_client_id
```

## API Документация

Вся документация по API доступна в следующих форматах:
1. **Интерактивный Центр документации и API в веб-интерфейсе:** доступен по маршруту `/readme` прямо в приложении с живым поиском по всем эндпоинтам, фильтрацией по сервисам, генератором cURL команд и схемами запросов.
2. **OpenAPI / Swagger v2 спецификация:** сгенерированный файл `backend/gen/swagger/api.swagger.json` (также доступен статически по URL `/api.swagger.json`).
3. **Protobuf контракты:** `backend/proto/*` — первоисточник типов и RPC контрактов.

### Ключевые группы маршрутов:
- `GET /api/v1/inventory` — Полнотекстовый поиск и каталог деталей (`quantity >= 0`)
- `POST /api/v1/parts` — Добавление детали
- `POST /api/v1/defect-reports` — Создание дефектных ведомостей (`quantity = 0`)
- `GET/POST /orders` — Управление заказами
- `PUT /orders/:id/complete` — Завершение заказа и автоматическое списание деталей (`quantity = -1`)
- `GET/POST /orders/customers` — Клиентская база с нормализацией телефонов `+7...`
- `POST /auth/login` — Аутентификация сотрудников
- `GET /auth/verify` — Проверка сессии через Traefik ForwardAuth (<1 мс)

Подробное описание архитектуры и сетевых потоков доступно в [backend/ARCHITECTURE.md](backend/ARCHITECTURE.md).

## Безопасность

- **Traefik ForwardAuth** на шлюзе: каждый защищенный запрос валидируется в `auth-service` до попадания в бизнес-сервисы
- **CSRF защита** на всех state-changing операциях
- **Ролевая авторизация (RBAC)**: Admin, Manager, Operator
- **Безопасные сессии** с HttpOnly и SameSite cookies
- **Парольный хэшинг** с bcrypt
- **Идемпотентность складских списаний** через таблицу `part_stock_operations`

## Разработка

### Структура проекта
```
avtoplaneta/
├── backend/                  # Микросервисы на Go
│   ├── pkg/                  # Общие библиотеки (httpserver, redisclient, authcontext, httpcache)
│   ├── proto/                # Proto-определения (.proto)
│   │   ├── auth/v1/          # Auth Service API
│   │   ├── parts/v1/         # Parts Service API
│   │   ├── orders/v1/        # Orders Service API
│   │   └── messaging/v1/     # Messaging Service API
│   ├── gen/                  # Сгенерированный Go-код и OpenAPI Swagger
│   │   └── swagger/          # api.swagger.json
│   ├── auth-service/         # Сервис аутентификации, сессий и пользователей
│   ├── parts-service/        # Сервис каталога запчастей, склада и дефектовок
│   ├── orders-service/       # Сервис заказов, клиентов и списания остатков
│   ├── messaging-service/    # Сервис сообщений и чатов
│   ├── export-service/       # Сервис фоновой генерации XML каталогов (Drom, Avito)
│   ├── Makefile              # Сборка, тесты, генерация proto и Swagger
│   └── ARCHITECTURE.md       # Подробная архитектура системы
├── frontend/                 # React 19 веб-приложение
│   ├── src/                  # Исходный код (включая Центр документации и API)
│   ├── public/               # Статические ресурсы (включая api.swagger.json)
│   └── package.json
├── AvtoplanetaApp/           # Мобильное приложение (Flutter)
├── docker-compose.yml        # Docker развертывание с Traefik Ingress
└── README.md                 # Корневая документация
```

### Скрипты
- `scripts/generate_parts.py` - Генерация тестовых данных запчастей

## Тестирование

### Backend
```bash
cd backend
go test ./...
```

### Frontend
```bash
cd frontend
npm test
```

## ИИ-помощник (OpenRouter)

Система включает голосового ИИ-помощника на базе OpenRouter для удобного управления.

### Настройка OpenRouter API

1. **Получите API ключ:**
   - Перейдите на [OpenRouter](https://openrouter.ai/keys)
   - Зарегистрируйтесь и получите бесплатный API ключ

2. **Настройте переменную окружения:**
```bash
export OPENROUTER_API_KEY=sk-or-v1-ваш_api_ключ_здесь
```

### Возможности ИИ-помощника

- **Голосовое управление** - говорите команды голосом
- **Умное понимание** - ИИ понимает естественный язык
- **Выполнение действий** - автоматическая навигация и выполнение команд
- **Контекстная помощь** - учитывает текущую страницу и историю
- **Fallback система** - работает даже без OpenRouter (rule-based логика)
- **Множество моделей** - доступ к Claude, GPT, Gemini и другим через единый API

### Примеры команд

- "Найди тормозные колодки BMW"
- "Добавь новую запчасть в инвентарь"
- "Покажи статистику продаж"
- "Перейди к управлению заказами"
- "Открой админ панель"

### Тестирование интеграции

```bash
# Установите переменную окружения
set OPENROUTER_API_KEY=sk-or-v1-ваш_api_ключ

# Протестируйте API
cd scripts && go run test-openrouter.go
```

Если OpenRouter API недоступен, система автоматически переключается на rule-based логику.

### Модель по умолчанию

- **xAI Grok 4.1 Fast Free** - быстрая бесплатная модель от xAI (создатели Grok)
- **Стоимость**: Бесплатно (free tier)
- **Бесплатный кредит**: Неограниченно для бесплатной версии

## Развертывание

### Production
```bash
# Сборка frontend
cd frontend && npm run build

# Сборка backend
cd backend && go build -o bin/avtoplaneta main.go

# Использовать docker-compose.prod.yml для продакшена
docker-compose -f docker-compose.prod.yml up -d
```

## Команда проекта

- **Разработка**: Команда Avtoplaneta
- **Технологии**: Go, React, Android
- **Контакты**: [указать контакты]

## Лицензия

Copyright (c) 2025 Avtoplaneta. Все права защищены.

## Вклад в проект

1. Форкните репозиторий
2. Создайте feature branch (`git checkout -b feature/AmazingFeature`)
3. Зафиксируйте изменения (`git commit -m 'Add some AmazingFeature'`)
4. Отправьте в branch (`git push origin feature/AmazingFeature`)
5. Откройте Pull Request

---

Для дополнительной информации смотрите документацию отдельных компонентов:
- [Архитектура Backend](backend/ARCHITECTURE.md)
- [Frontend](frontend/README.md)
- [Mobile App](AvtoplanetaApp/README.md)

<p align="center">
<img src="https://fonts.gstatic.com/s/e/notoemoji/latest/1f6f8/512.gif" alt="🛸" width="132" height="132">