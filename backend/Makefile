export PATH := $(PATH):$(HOME)/go/bin:$(shell go env GOPATH 2>/dev/null)/bin

.DEFAULT_GOAL := help
.PHONY: help proto proto-lint proto-format proto-breaking proto-clean proto-deps proto-rebuild \
        build-all build-services build-binaries deps start-dev test test-integration \
        vehicle-catalog setup-hooks new-service swagger-ui godoc

help: ## Показать интерактивную справку по всем командам бэкенда
	@echo "Доступные команды backend/Makefile:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Примеры:"
	@echo "  make proto                  - Скомпилировать protobuf и обновить OpenAPI"
	@echo "  make swagger-ui             - Запустить локальный Swagger UI на http://localhost:8088"
	@echo "  make test                   - Запустить юнит-тесты всех микросервисов"
	@echo "  make build-services         - Проверить компиляцию всех микросервисов"
	@echo "  make new-service NAME=...   - Создать новый микросервис из шаблона"

# Генерация Go-кода из proto-файлов и синхронизация OpenAPI спецификации
proto: ## Сгенерировать Go-код и OpenAPI спецификацию из proto-файлов через Buf CLI
	@echo "Generating protobuf code..."
	cd proto && buf generate
	cd proto && buf generate --template buf.gen.swagger.yaml
	@mkdir -p ../frontend/src/data
	cp gen/swagger/api.swagger.json ../frontend/src/data/swagger.json
	@echo "Protobuf and OpenAPI specification generated and synced to frontend."

# Настройка git hooks для автоматической генерации при коммитах
setup-hooks: ## Настроить git hooks для валидации и генерации перед коммитами
	git config core.hooksPath .githooks
	chmod +x ../.githooks/* 2>/dev/null || true
	@echo "Git hooks configured successfully (.githooks)."

# Обновление справочника марок и моделей с drom.ru
vehicle-catalog: ## Синхронизировать справочник марок и моделей с drom.ru
	cd tools/vehicle-catalog-sync && go mod tidy && go run . -out ../../parts-service/vehicles/vehicles.json
	cd parts-service && go test -run Vehicle ./...
	@echo "Готово. Проверьте дифф parts-service/vehicles/vehicles.json перед коммитом."

# Линтинг proto-файлов
proto-lint: ## Запустить строгий линтинг Protobuf (Buf CLI, категория STANDARD)
	cd proto && buf lint

# Автоформатирование proto-файлов
proto-format: ## Отформатировать proto-файлы по Google API Style Guide (buf format -w)
	cd proto && buf format -w

# Проверка обратной совместимости API относительно ветки main
proto-breaking: ## Проверить обратную совместимость proto-контрактов относительно ветки main
	cd proto && buf breaking --against '.git#branch=main'

# Очистка сгенерированного кода (сохраняя go.mod и go.sum)
proto-clean: ## Удалить весь ранее сгенерированный Protobuf/gRPC код
	rm -rf gen/auth gen/messaging gen/orders gen/parts gen/swagger gen/google

# Установка зависимостей для protobuf и Buf тулчейна
proto-deps: ## Установить генераторы Protobuf, gRPC-Gateway, OpenAPI и Buf CLI
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install github.com/bufbuild/buf/cmd/buf@latest

# Полная перегенерация
proto-rebuild: proto-clean proto ## Полная перегенерация Protobuf с предварительной очисткой

# Сборка всех сервисов (проверка компиляции)
build-all: build-services ## Алиас для build-services

build-services: ## Проверить компиляцию всех микросервисов и общих пакетов
	cd auth-service && go build -o /dev/null ./...
	cd parts-service && go build -o /dev/null ./...
	cd orders-service && go build -o /dev/null ./...
	cd messaging-service && go build -o /dev/null ./...
	cd export-service && go build -o /dev/null ./...
	cd template-service && go build -o /dev/null ./...
	cd pkg && go test -run=^$$ ./...
	@echo "All services build OK"

# Сборка бинарников для запуска в контейнерах (без компиляции внутри Docker)
build-binaries: ## Скомпилировать статические бинарники всех сервисов в папку bin/
	@echo "Building auth-service static binary..."
	cd auth-service && mkdir -p bin && CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/auth-service .
	@echo "Building parts-service static binary..."
	cd parts-service && mkdir -p bin && CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/parts-service .
	@echo "Building orders-service static binary..."
	cd orders-service && mkdir -p bin && CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/orders-service .
	@echo "Building messaging-service static binary..."
	cd messaging-service && mkdir -p bin && CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/messaging-service .
	@echo "Building export-service static binary..."
	cd export-service && mkdir -p bin && CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/export-service .
	@echo "Building template-service static binary..."
	cd template-service && mkdir -p bin && CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/template-service .
	@echo "All static binaries built successfully."

# Установка всех Go-зависимостей
deps: ## Загрузить все Go-зависимости (go mod download во всех сервисах и pkg)
	cd auth-service && go mod download
	cd parts-service && go mod download
	cd orders-service && go mod download
	cd messaging-service && go mod download
	cd export-service && go mod download
	cd template-service && go mod download
	cd gen && go mod download
	cd pkg && go mod download
	@echo "All dependencies downloaded"

# Запуск всех сервисов в фоне (для разработки)
start-dev: ## Запустить локально все микросервисы в фоновом режиме
	cd auth-service && go run . &
	cd parts-service && go run . &
	cd orders-service && go run . &
	cd messaging-service && go run . &
	cd export-service && go run . &
	@echo "All services started"

# Запуск всех юнит-тестов во всех сервисах
test: ## Запустить юнит-тесты всех микросервисов и библиотеки pkg
	@echo "Running unit tests for all services..."
	cd auth-service && go test ./...
	cd parts-service && go test ./...
	cd orders-service && go test ./...
	cd messaging-service && go test ./...
	cd export-service && go test ./...
	cd template-service && go test ./...
	cd pkg && go test ./...
	@echo "All unit tests passed!"

# Генерация нового микросервиса из эталонного шаблона
new-service: ## Создать новый микросервис: make new-service NAME=billing PORT=8086 GRPC_PORT=9086
	@bash scripts/new-service.sh "$(NAME)" "$(PORT)" "$(GRPC_PORT)"

# Запуск интеграционных тестов (требуется запущенная БД)
test-integration: ## Запустить интеграционные тесты (-tags=integration)
	@echo "Running integration tests..."
	cd auth-service && go test -tags=integration ./...
	@echo "Integration tests passed!"

# Запуск локального интерактивного Swagger UI
swagger-ui: ## Запустить Swagger UI в Docker на http://localhost:8088 для просмотра api.swagger.json
	@echo "Starting Swagger UI on http://localhost:8088 (spec: gen/swagger/api.swagger.json)..."
	docker run --rm -it -p 8088:8080 -e SWAGGER_JSON=/spec/api.swagger.json -v $(CURDIR)/gen/swagger/api.swagger.json:/spec/api.swagger.json:ro swaggerapi/swagger-ui

# Запуск сервера документации GoDoc
godoc: ## Запустить локальный сервер документации кода Go на http://localhost:6060
	@echo "Starting GoDoc documentation server on http://localhost:6060..."
	@which pkgsite >/dev/null 2>&1 && pkgsite -http=:6060 || godoc -http=:6060
