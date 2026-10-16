# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Копируем общие модули для разрешения локальных replace
COPY gen ../gen
COPY pkg ../pkg
COPY template-service/go.mod template-service/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY template-service/ .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOMAXPROCS=16 CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /template-service .

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /template-service .

RUN adduser -D -g '' appuser && chown -R appuser:appuser /app
USER appuser

EXPOSE 8090 9090

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8090/health || exit 1

CMD ["./template-service"]
