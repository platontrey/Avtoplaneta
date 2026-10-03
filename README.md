# Avtoplaneta - Система управления автозапчастями

## О проекте

Avtoplaneta - это комплексная система управления инвентарем автозапчастей, разработанная с использованием современных технологий. Система включает веб-приложение, мобильное приложение и микросервисную backend архитектуру для эффективного управления запасами, заказами и пользователями.

## Архитектура проекта

Проект состоит из следующих основных компонентов:

### 🖥️ Frontend (Веб-приложение)
- **Технологии**: React 19, TypeScript, Vite, Tailwind CSS v4, Nginx (с Brotli & Gzip сжатием)
- **Расположение**: `frontend/`
- **Документация**: [frontend/README.md](frontend/README.md)

### 📱 Mobile App (iOS / Android)
- **Технологии**: Flutter, Dart, Android SDK, iOS SDK
- **Расположение**: `AvtoplanetaApp/`
- **Документация**: [AvtoplanetaApp/README.md](AvtoplanetaApp/README.md)

### 🔧 Backend (Микросервисы)
- **Технологии**: Go, Gin Framework, gRPC, Protobuf, PostgreSQL (sqlc, pgx/v5), Elasticsearch, Redis Streams, OpenTelemetry, Grafana Tempo
- **Расположение**: `backend/`
- **Документация**: [backend/ARCHITECTURE.md](backend/ARCHITECTURE.md)

#### Сервисы:
- **Traefik Ingress** (порты 80, 443) — современный reverse proxy, SSL termination, прямая маршрутизация на сервисы и проверка авторизации через ForwardAuth (`/auth/verify`)
- **Auth Service** (HTTP :8083, gRPC :9083) — аутентификация, ForwardAuth-проверка, управление пользователями, gRPC ValidateSession / GetUsers / ActivityLogs
- **Parts Service** (HTTP :8081, gRPC :9081) — управление инвентарем, gRPC API и встроенный gRPC-Gateway (`/api/v1/*`), шаблоны каталога (`catalog.json`), Elasticsearch
- **Orders Service** (HTTP :8082, gRPC :9082) — управление заказами, списание остатков (`quantity = -1`), аналитика продаж через gRPC
- **Messaging Service** (HTTP :8084, gRPC :9084) — чаты, уведомления, интеграция с Drom.ru
- **Export Service** (HTTP :8085) — генерация прайс-листов XML для Drom.ru через потоковый gRPC без прямого доступа к БД

#### Коммуникация и Наблюдаемость (Observability):
- **REST/HTTP & gRPC-Gateway** — внешний веб-интерфейс через Traefik Ingress
- **gRPC + Protobuf** — 100% межсервисной коммуникации между микросервисами (без внутренних HTTP-вызовов)
- **Redis Streams** — для асинхронной доставки доменных событий (переименование продавцов, события заказов)
- **OpenTelemetry & Grafana Tempo** — распределённый сквозной трейсинг запросов
- **Prometheus & Grafana** — сбор метрик производительности, количества ошибок и системного мониторинга

## Инвентарь и статус количества (Quantity States)

В системе используется строгое распределение статусов количества запчастей:
- `quantity > 0`: Обычные запчасти в наличии на складе.
- `quantity == 0`: Дефектные ведомости или специальные запчасти. Они **должны оставаться видимыми** в интерфейсе и при генерации XML-выгрузок (`quantity >= 0`).
- `quantity == -1`: Запчасти, полностью проданные через `orders-service`. Маркер `-1` отфильтровывает запчасть из выдачи UI и XML, сохраняя историчность базы данных.

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