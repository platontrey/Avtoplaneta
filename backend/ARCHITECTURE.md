# Документация по архитектуре проекта Avtoplaneta

## Обзор проекта

Проект Avtoplaneta представляет собой микросервисную архитектуру на языке Go для управления автозапчастями. Система включает в себя управление запасами, заказами и аутентификацией пользователей.

## Общая архитектура

```
┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   API Gateway   │
│   (React/Vue)   │◄──►│   (Gin)         │
│                 │    │   Port: 8080    │
└─────────────────┘    └─────────────────┘
                              │
                    ┌─────────┼─────────┐
                    │         │         │
            ┌───────▼───┐ ┌───▼───┐ ┌───▼───┐
            │ Auth      │ │Orders │ │Parts  │
            │ Service   │ │Service│ │Service│
            │ Port:8083 │ │:8082  │ │:8081  │
            └───────────┘ └───────┘ └───────┘
                    │         │         │
            ┌───────▼───┐ ┌───▼───┐ ┌───▼───┐
            │PostgreSQL │ │Postgre│ │Postgre│
            │Database   │ │SQL DB │ │SQL DB │
            │(Users)    │ │(Orders│ │(Parts)│
            └───────────┘ │Items) │ └───────┘
                          └───────┘     │
                                        │
                                ┌───────▼─────┐
                                │Elasticsearch│
                                │Port: 9200   │
                                │(Search)     │
                                └─────────────┘
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

### 3. API Gateway (Корневой сервис)

**Файл:** `main.go`
**Порт:** 8080
**Технологии:** Gin Framework, HTTP Proxy

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
| `/uploads/*` | Parts Service | 8081 | Статические файлы изображений |

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
**Технологии:** Gin, GORM, PostgreSQL, Gorilla Sessions, Goth (Google OAuth)

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
    ID       uint   `gorm:"primaryKey"`
    Email    string `gorm:"unique;not null"`
    Name     string
    Provider string // "google" или "local"
    Role     string // "admin", "manager", "operator"
    Password string `gorm:"not null;default:''"` // Хэшируется bcrypt
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
**Технологии:** Gorilla Mux, GORM, PostgreSQL, Gorilla Sessions

#### Основные функции:
- **Создание заказов** с привязкой к частям
- **Управление статусами заказов** (цветовая кодировка)
- **Просмотр заказов** с пагинацией
- **Удаление заказов**
- **Интеграция с частями** (автоматическое помечение для удаления)

#### Модели данных:

```go
type Order struct {
    ID                 uint        `gorm:"primaryKey"`
    CustomerID         int
    SellerID           uint        // ID продавца
    Seller             string      // Имя продавца
    Part               string      // Описание части
    OrderNumber        string      // Номер заказа
    BuyerNumber        string      // Номер покупателя
    Status             string      // "red", "brown", "yellow", "green"
    StatusText         string      // Текстовое описание статуса
    CreatedAt          time.Time
    Items              []OrderItem `gorm:"foreignKey:OrderID"`
}

type OrderItem struct {
    ID       uint `gorm:"primaryKey"`
    OrderID  uint
    PartID   uint
    Quantity int
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

### 6. Parts Service (Сервис запчастей)

**Директория:** `parts-service/`
**Порт:** 8081
**Технологии:** Gin, GORM, PostgreSQL, Elasticsearch

#### Основные функции:
- **Управление запасами** (CRUD операции с частями)
- **Поиск и фильтрация** через Elasticsearch
- **Загрузка фотографий** с валидацией
- **Статистика** по категориям и стоимости
- **Автоматическое удаление** просроченных частей (14 дней)

#### Модели данных:

```go
type Part struct {
    ID          uint       `gorm:"primaryKey"`
    Name        string     `gorm:"not null"`
    Quantity    int        `gorm:"not null"`
    Description string
    Category    string
    Price       float64
    Salesman    string
    Location    string
    Status      bool       // true/false
    Brand       string
    Model       string
    Photo       string     // Путь к файлу
    ToDeleteAt  *time.Time // Дата автоматического удаления
}
```

#### Elasticsearch интеграция:
- **Индексация**: Все части автоматически индексируются
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
- Общее количество частей
- Общая стоимость (price * quantity)
- Распределение по категориям

## База данных

### PostgreSQL
**Все сервисы используют одну базу данных** с разделением по схемам/таблицам:

- **users** - таблица пользователей (auth-service)
- **orders** - таблица заказов (orders-service)
- **order_items** - позиции заказов (orders-service)
- **parts** - таблица запчастей (parts-service)

### Elasticsearch
- **Индекс:** `parts`
- **Маппинг:** Оптимизирован для поиска (text + keyword поля)
- **Функции:** Полнотекстовый поиск, фильтры, сортировка

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
- **Индексы БД**: GORM авто-миграции создают необходимые индексы
- **Пагинация**: Во всех списочных запросах
- **Кэширование**: CSRF токены (в памяти, Redis в продакшене)

### Масштабируемость:
- **Микросервисы**: Независимое масштабирование каждого сервиса
- **База данных**: Общая БД с разделением по таблицам
- **Elasticsearch**: Отдельный кластер для поиска

## Разработка и тестирование

### Структура проекта:
```
backend/
├── main.go                 # API Gateway
├── auth-service/
│   ├── main.go            # Auth сервис
│   ├── models.go          # Модели пользователей
│   └── auth.go            # Логика аутентификации
├── orders-service/
│   ├── main.go            # Orders сервис
│   ├── models.go          # Модели заказов
│   └── handlers.go        # HTTP обработчики
└── parts-service/
    ├── main.go            # Parts сервис
    ├── models.go          # Модели частей
    ├── handlers.go        # HTTP обработчики
    └── elasticsearch.go   # Elasticsearch интеграция
```

### Лучшие практики:
- **Валидация входных данных** во всех endpoints
- **Обработка ошибок** с понятными сообщениями
- **Логирование** всех важных операций
- **Безопасность** на всех уровнях
- **Тестирование** API endpoints

## Заключение

Архитектура проекта Avtoplaneta демонстрирует современный подход к разработке микросервисных приложений с акцентом на безопасность, производительность и масштабируемость. Каждый сервис имеет четко определенные обязанности, что облегчает поддержку и развитие системы.