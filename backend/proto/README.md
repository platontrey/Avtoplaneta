# Protobuf API Контракты Avtoplaneta (Buf CLI & Protovalidate)

Данный каталог содержит канонические Protobuf контракты всех микросервисов платформы **Avtoplaneta**.

---

## 1. Архитектурные стандарты

1. **Buf CLI:** Все контракты управляются исключительно через утилиту `buf`. Никакого прямого вызова `protoc` или локального вендоринга `google/api`. Зависимости подтягиваются из Buf Schema Registry (BSR):
   - `buf.build/googleapis/googleapis`
   - `buf.build/bufbuild/protovalidate`
2. **Строгая категория линтинга `STANDARD`:**
   - Включена категория линтинга `STANDARD` (Google API Style Guide).
   - Запрещено отключать правила:
     - `RPC_REQUEST_RESPONSE_UNIQUE` — каждый RPC метод обязан иметь уникальные Request и Response типы.
     - `RPC_REQUEST_STANDARD_NAME` — имя запроса обязано быть `<Method>Request`.
     - `RPC_RESPONSE_STANDARD_NAME` — имя ответа обязано быть `<Method>Response`.
3. **Декларативная валидация Protovalidate (CEL):**
   - Все входящие поля валидируются декларативно на уровне `.proto` схемы через аннотации `(buf.validate.field)`.
   - В рантайме gRPC-серверы используют `pkg/grpcserver.ValidationUnaryInterceptor()`, отклоняющий невалидные запросы с `codes.InvalidArgument`.
4. **Bigint Standard (единый 64-битный ID):**
   - Все идентификаторы (`id`, `user_id`, `part_id`, `order_id`, `seller_id`, `resource_id`) объявляются как `int64`.
5. **Совместимость с REST (gRPC-Gateway):**
   - Для сохранения существующей структуры ответов во фронтенде используется аннотация `response_body: "<поле>"` в опциях `google.api.http`.

---

## 2. Основные команды

Все команды выполняются из каталога `backend/`:

```bash
# Генерация Go-кода и OpenAPI (Swagger) спецификации
make proto

# Проверка линтинга по категории STANDARD (должно быть 0 ошибок и 0 предупреждений)
make proto-lint

# Автоматическое форматирование всех .proto файлов
make proto-format

# Проверка обратной совместимости API относительно ветки main
make proto-breaking

# Очистка сгенерированного кода
make proto-clean
```

---

## 3. Структура каталогов

```
backend/proto/
├── buf.yaml              # Конфигурация модуля Buf, зависимости BSR и линтер
├── buf.lock              # Зафиксированные версии BSR зависимостей
├── buf.gen.yaml          # Плагины кодогенерации Go (go, go-grpc, grpc-gateway)
├── buf.gen.swagger.yaml  # Плагин генерации OpenAPI v2 спецификации
├── auth/v1/auth.proto
├── parts/v1/parts.proto
├── orders/v1/orders.proto
└── messaging/v1/messaging.proto
```
