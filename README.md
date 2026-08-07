# Avtoplaneta - Система управления автозапчастями

## О проекте

Avtoplaneta - это комплексная система управления инвентарем автозапчастей, разработанная с использованием современных технологий. Система включает веб-приложение, мобильное приложение и микросервисную backend архитектуру для эффективного управления запасами, заказами и пользователями.

## Архитектура проекта

Проект состоит из следующих основных компонентов:

### 🖥️ Frontend (Веб-приложение)
- **Технологии**: React 18, TypeScript, Vite, Tailwind CSS, Nginx (с Brotli & Gzip сжатием)
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
- **API Gateway** (порт 8080) - основной шлюз, gRPC-клиент ко всем микросервисам с OTEL instrumentation
- **Auth Service** (HTTP :8083, gRPC :9083) - аутентификация, управление пользователями, gRPC ValidateSession / GetUsers / ActivityLogs
- **Parts Service** (HTTP :8081, gRPC :9081) - управление инвентарем, gRPC API, шаблоны каталога (`catalog.json`), Elasticsearch
- **Orders Service** (HTTP :8082, gRPC :9082) - управление заказами, списание остатков (`quantity = -1`), аналитика продаж через gRPC
- **Messaging Service** (HTTP :8084, gRPC :9084) - чаты, уведомления, интеграция с Drom.ru

#### Коммуникация и Наблюдаемость (Observability):
- **REST/HTTP** — внешний веб-интерфейс к Gateway и статической раздаче через Nginx
- **gRPC + Protobuf** — высокоскоростная межсервисная коммуникация и внутренние эндпоинты
- **Redis Streams** — для асинхронной доставки событий (дефектные отчёты, списания)
- **OpenTelemetry & Grafana Tempo** — распределённый сквозной трейсинг запросов от Gateway до баз данных
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

# Генерация gRPC-кода из proto-файлов (требуется protoc + buf)
make proto

# Установка Go-зависимостей
go mod download
cd auth-service && go mod download && cd ..
cd orders-service && go mod download && cd ..
cd parts-service && go mod download && cd ..
cd messaging-service && go mod download && cd ..

# Настройте переменные окружения в .env файле
cp .env.example .env
# Отредактируйте .env файл с вашими настройками

# Запуск API Gateway
go run main.go

# В отдельных терминалах запустить сервисы:
cd auth-service && go run .
cd orders-service && go run .
cd parts-service && go run .
cd messaging-service && go run .
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
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
SESSION_SECRET=secure_random_key
fOPENROUTER_API_KEY=sk-or-v1-your-openrouter-key
```

**Frontend (.env):**
```env
VITE_API_BASE_URL=http://localhost:8080
VITE_GOOGLE_CLIENT_ID=your_google_client_id
```

## API Документация

### Основные endpoints
- `GET /api/inventory` - Получение списка запчастей
- `POST /api/addpart` - Добавление новой запчасти
- `GET /orders` - Получение списка заказов
- `POST /auth/login` - Аутентификация пользователя

Подробная документация API доступна в [backend/ARCHITECTURE.md](backend/ARCHITECTURE.md)

## Безопасность

- **CSRF защита** на всех state-changing операциях
- **Ролевая авторизация** с middleware проверками
- **Безопасные сессии** с HttpOnly cookies
- **Парольный хэшинг** с bcrypt
- **Rate limiting** для предотвращения атак

## Разработка

### Структура проекта
```
avtoplaneta/
├── backend/                  # Микросервисы Go
│   ├── main.go              # API Gateway
│   ├── gateway.go           # gRPC-клиенты, маршрутизация
│   ├── proto/               # Proto-определения (.proto)
│   │   ├── auth/v1/         # Auth Service API
│   │   ├── parts/v1/        # Parts Service API
│   │   ├── orders/v1/       # Orders Service API
│   │   └── messaging/v1/    # Messaging Service API
│   ├── gen/                 # Сгенерированный Go-код из proto
│   ├── auth-service/        # Сервис аутентификации
│   ├── orders-service/      # Сервис заказов
│   ├── parts-service/       # Сервис запчастей
│   ├── messaging-service/   # Сервис сообщений/чатов
│   ├── Makefile             # Proto-генерация и утилиты
│   └── ARCHITECTURE.md      # Документация архитектуры
├── frontend/                # React приложение
│   ├── src/
│   ├── package.json
│   └── README.md
├── AvtoplanetaApp/          # Android приложение
│   ├── app/
│   └── README.md
├── scripts/                 # Скрипты генерации данных
├── docker-compose.yml       # Docker конфигурация
└── README.md               # Этот файл
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