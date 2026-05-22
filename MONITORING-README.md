# Мониторинг Avtoplaneta

## Обзор

Проект Avtoplaneta теперь имеет полную систему мониторинга на базе Prometheus и Grafana. Все микросервисы оснащены унифицированными метриками для отслеживания производительности, latency, throughput и ошибок.

## Архитектура Мониторинга

```
Сервисы -> Prometheus -> Grafana
    |           |          |
  Метрики   Сбор данных  Визуализация
```

## Метрики

### Стандартные HTTP Метрики (все сервисы)
- `avtoplaneta_http_requests_total{method, endpoint, status}` - Счетчик общего количества HTTP запросов
- `avtoplaneta_http_request_duration_seconds{method, endpoint}` - Гистограмма времени выполнения запросов
- `avtoplaneta_http_requests_errors_total{method, endpoint, status}` - Счетчик HTTP ошибок (4xx/5xx)

### Специфические Метрики Messaging Service
- `avtoplaneta_db_errors_total{operation, service}` - Ошибки базы данных
- `avtoplaneta_business_operations_total{operation, service, status}` - Бизнес операции

## Установка и Настройка

### 1. Prometheus

```bash
# Скачайте и установите Prometheus
wget https://github.com/prometheus/prometheus/releases/download/v2.45.0/prometheus-2.45.0.linux-amd64.tar.gz
tar xvfz prometheus-2.45.0.linux-amd64.tar.gz
cd prometheus-2.45.0.linux-amd64/

# Запустите с нашей конфигурацией
./prometheus --config.file=/path/to/avtoplaneta/prometheus.yml
```

Prometheus будет доступен на http://localhost:9090

### 2. Grafana

```bash
# Установите Grafana
# Для Ubuntu/Debian:
sudo apt-get install grafana

# Запустите сервис
sudo systemctl start grafana-server

# Или запустите вручную
grafana-server
```

Grafana будет доступна на http://localhost:3000 (admin/admin)

### 3. Настройка Grafana

1. Войдите в Grafana (admin/admin)
2. Добавьте источник данных:
   - Configuration → Data Sources → Add data source
   - Выберите Prometheus
   - URL: http://localhost:9090
   - Сохраните
3. Импортируйте дашборд:
   - Dashboards → Import
   - Загрузите файл `grafana-dashboard.json`

## Эндпоинты Метрик

| Сервис | URL | Порт | Метрики |
|--------|-----|------|---------|
| API Gateway | http://localhost:8080/metrics | 8080 | HTTP метрики |
| Messaging Service | http://localhost:8084/metrics | 8084 | HTTP + бизнес метрики |
| Parts Service | http://localhost:8081/metrics | 8081 | HTTP метрики |
| Orders Service | http://localhost:8082/metrics | 8082 | HTTP метрики |

## Запуск Сервисов

### Порядок запуска:

1. **Базы данных** (PostgreSQL, Redis)
2. **Сервисы** в любом порядке
3. **Prometheus**
4. **Grafana**

```bash
# Пример запуска сервисов
cd backend

# Gateway
go run gateway.go &

# Messaging Service
cd messaging-service && go run . &

# Parts Service
cd ../parts-service && go run . &

# Orders Service
cd ../orders-service && go run . &
```

## Мониторинг Дашборда

Дашборд включает следующие панели:

1. **HTTP Request Rate** - Количество запросов в секунду по методам и эндпоинтам
2. **HTTP Request Duration** - 50-й и 95-й перцентили latency
3. **HTTP Error Rate** - Количество ошибок по типам
4. **Database Errors** - Ошибки базы данных (Messaging Service)
5. **Business Operations** - Бизнес операции с статусами

## Алертинг

Для настройки алертов в Prometheus добавьте правила в `prometheus.yml`:

```yaml
rule_files:
  - "alert_rules.yml"

# alert_rules.yml
groups:
  - name: avtoplaneta
    rules:
      - alert: HighErrorRate
        expr: rate(avtoplaneta_http_requests_errors_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
```

## Troubleshooting

### Метрики не отображаются
- Проверьте, что сервисы запущены и отвечают на /metrics
- Проверьте конфигурацию scrape в prometheus.yml
- Проверьте логи Prometheus

### Grafana не подключается к Prometheus
- Убедитесь, что Prometheus запущен на правильном порту
- Проверьте настройки datasource в Grafana

### Высокая latency
- Проверьте нагрузку на сервисы
- Мониторьте использование CPU/памяти
- Проверьте подключения к базам данных

## Производительность

Метрики собираются каждые 15 секунд без значительного влияния на производительность сервисов. Гистограммы используют стандартные бакеты Prometheus для оптимального сжатия данных.