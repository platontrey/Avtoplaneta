FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY gen ../gen
COPY pkg ../pkg
COPY parts-service/go.mod parts-service/go.sum ./
RUN go mod download
COPY parts-service/ .

RUN GOMAXPROCS=16 CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /parts-service .

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata su-exec

WORKDIR /app

COPY --from=builder /parts-service .
COPY parts-service/entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

RUN adduser -D -g '' appuser && mkdir -p /app/uploads && chown -R appuser:appuser /app

EXPOSE 8081

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

ENTRYPOINT ["./entrypoint.sh"]
CMD ["./parts-service"]
