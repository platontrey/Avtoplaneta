.PHONY: proto proto-lint proto-clean proto-deps proto-rebuild build-all build-gateway build-services start-dev test test-integration vehicle-catalog

# Генерация Go-кода из proto-файлов
proto:
	@echo "Generating protobuf code..."
	cd proto && buf generate
	cd proto && buf generate --template buf.gen.swagger.yaml
	@echo "Done."

# Обновление справочника марок и моделей с drom.ru.
# Ходит в сеть, поэтому запускается руками (или отдельной job в CI), а не сервисом:
# результат коммитится и вшивается в бинарь через go:embed.
vehicle-catalog:
	cd tools/vehicle-catalog-sync && go mod tidy && go run . -out ../../parts-service/vehicles/vehicles.json
	cd parts-service && go test -run Vehicle ./...
	@echo "Готово. Проверьте дифф parts-service/vehicles/vehicles.json перед коммитом."

# Линтинг proto-файлов
proto-lint:
	cd proto && buf lint

# Очистка сгенерированного кода
proto-clean:
	rm -rf gen/

# Установка зависимостей для protobuf
proto-deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/bufbuild/buf/cmd/buf@latest

# Полная перегенерация
proto-rebuild: proto-clean proto

# Сборка всех сервисов (проверка компиляции)
build-all: build-gateway build-services

build-gateway:
	go build -o /dev/null ./...

build-services:
	cd auth-service && go build -o /dev/null ./...
	cd parts-service && go build -o /dev/null ./...
	cd orders-service && go build -o /dev/null ./...
	cd messaging-service && go build -o /dev/null ./...
	cd export-service && go build -o /dev/null ./...
	@echo "All services build OK"

# Сборка бинарников для запуска в контейнерах (без компиляции внутри Docker)
build-binaries:
	@echo "Building gateway static binary..."
	mkdir -p bin && CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/gateway .
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
	@echo "All static binaries built successfully."

# Установка всех Go-зависимостей
deps:
	go mod download
	cd auth-service && go mod download
	cd parts-service && go mod download
	cd orders-service && go mod download
	cd messaging-service && go mod download
	cd export-service && go mod download
	cd gen && go mod download
	@echo "All dependencies downloaded"

# Запуск всех сервисов в фоне (для разработки)
start-dev:
	go run . main.go gateway.go ai.go metrics.go &
	cd auth-service && go run . &
	cd parts-service && go run . &
	cd orders-service && go run . &
	cd messaging-service && go run . &
	cd export-service && go run . &
	@echo "All services started"

# Запуск всех юнит-тестов во всех сервисах
test:
	@echo "Running unit tests for all services..."
	cd auth-service && go test ./...
	cd parts-service && go test ./...
	cd messaging-service && go test ./...
	cd export-service && go test ./...
	@echo "All unit tests passed!"

# Запуск интеграционных тестов (требуется запущенная БД)
test-integration:
	@echo "Running integration tests..."
	cd auth-service && go test -tags=integration ./...
	@echo "Integration tests passed!"
