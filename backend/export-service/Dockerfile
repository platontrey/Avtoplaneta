FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY gen ../gen
COPY export-service/go.mod export-service/go.sum ./
RUN go mod download
COPY export-service/ .

RUN GOMAXPROCS=16 CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /export-service .

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /export-service .

RUN adduser -D -g '' appuser && mkdir -p /app/uploads && chown -R appuser:appuser /app
USER appuser

EXPOSE 8085

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8085/health || exit 1

CMD ["./export-service"]
