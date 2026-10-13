# Архитектурные правила и контекст проекта Avtoplaneta (.agents/AGENTS.md)

> Данный файл синхронизирован с корневым [AGENTS.md](../AGENTS.md) и является руководством для ИИ-ассистентов.

## 1. Коммуникация и Git / CI/CD Workflow
- **Язык общения:** **ИСКЛЮЧИТЕЛЬНО РУССКИЙ ЯЗЫК** (Русский язык). Все ответы, объяснения и комментарии формулируются на русском.
- **Локальные коммиты:** При любых существенных изменениях фиксировать их локально (`git add` и `git commit`) с понятными описаниями.
- **СТРОГИЙ ЗАПРЕТ НА `git push`:** **НИКОГДА не выполнять `git push` от имени пользователя.** Пользователь вручную контролирует отправку и запуск CI/CD пайплайна на сервере.
- **Docker Compose:** Всегда использовать синтаксис `docker compose` (с пробелом). Команда `docker-compose` запрещена.

## 2. Архитектура и сетевая топология
- **Бэкенд:** Микросервисная архитектура на Go (1.21+): `auth-service`, `parts-service`, `orders-service`, `messaging-service`, `export-service`.
- **Маршрутизация:** Все внешние запросы идут через **Traefik Ingress** (порты 80, 443). Монолитный API Gateway полностью удален.
- **Безопасность:** Проверка сессий через **Traefik ForwardAuth** (`auth-service /auth/verify` и `/auth/verify-admin`). Проверенные заголовки `X-User-ID`, `X-User-Role`, `X-User-Name`, `X-User-Email` пробрасываются в сервисы и читаются через `pkg/authcontext`.
- **Межсервисная коммуникация:** 100% межсервисных вызовов — строго **gRPC (Protobuf)**. Внутренние HTTP-вызовы и доступ к чужим БД запрещены. Контракты в `backend/proto/` компилируются через `make proto`.
- **База данных:** PostgreSQL по модели **Database-per-Service** через **`sqlc` + `pgx/v5`**. Тяжелые ORM запрещены. При изменении `db/queries/*.sql` запускать `sqlc generate`.
- **Шина событий и кэш:** Redis Streams (события `seller_renamed`, заказы), кэширование сессий и CSRF.
- **Поиск:** Elasticsearch с морфологией русского языка, синонимами `synonyms.txt` и fuzzy-поиском.
- **Export Service:** Сервис намеренно не имеет БД. Читает склад потоком через gRPC из `parts-service` и реквизиты из `auth-service`. Проверяет ETag-отпечаток склада (`InventoryVersion`) перед сборкой XML.

## 3. Критические доменные инварианты (Инвентарь и заказы)
- **Статусы количества (`quantity`):**
  - `quantity > 0`: деталь в наличии на складе.
  - `quantity == 0`: дефектные ведомости / спецпозиции. **ОБЯЗАНЫ ОСТАВАТЬСЯ ВИДИМЫМИ** в поиске, UI и XML-выгрузках.
  - `quantity == -1`: полностью списанные/проданные детали через `orders-service`. Маркер сокрытия из выдачи при сохранении истории.
  - **Правило запросов:** ВСЕГДА использовать `WHERE quantity >= 0`. Никогда не использовать `> 0`.
- **Списание заказов:** При исчерпании остатка (`quantity - amount <= 0`) выставлять `quantity = -1`. Изменения фиксируются идемпотентно в `part_stock_operations`.
- **CRM и дедупликация:** `parseBuyerInfo(raw)` нормализует телефоны к `+7XXXXXXXXXX`. Уникальный индекс `idx_customers_unique_phone` на уровне БД.

## 4. Frontend & Mobile
- **Frontend:** React 19, TypeScript, Vite, TanStack Query v5, Tailwind CSS v4. INP < 200ms, параллельная загрузка фото (`Promise.all`), debounced поиск, отсутствие тяжелых анимаций в списках.
- **Mobile (AvtoplanetaApp):** Flutter 3.x, Riverpod 2.6, GoRouter 15.1, Dio, FlutterSecureStorage, Hive, In-App APK обновления, сканер штрихкодов. Единый каталог маршрутов в `lib/core/api/api_endpoints.dart`.
