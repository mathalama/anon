FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app

# 1. Загрузка зависимостей
COPY go.work go.work.sum ./
COPY services/api-gateway/go.mod services/api-gateway/go.sum ./services/api-gateway/
COPY libs/shared/go.mod libs/shared/go.sum ./libs/shared/
COPY libs/proto/go.mod libs/proto/go.sum ./libs/proto/

ENV GOPROXY=https://proxy.golang.org,direct
ENV GODEBUG=http2client=0

WORKDIR /app/services/api-gateway
RUN GOWORK=off go mod download

# 2. Точечное копирование кода (кэш не сбросится при изменении других сервисов)
WORKDIR /app
COPY libs/shared ./libs/shared
COPY libs/proto ./libs/proto
COPY services/api-gateway ./services/api-gateway

# 3. Сборка с кэшированием (BuildKit) и вырезкой отладочной информации (LDFLAGS)
WORKDIR /app/services/api-gateway
RUN --mount=type=cache,target=/root/.cache/go-build \
    GOWORK=off CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/main ./cmd/main.go

# 4. Финальный образ (Безопасность + Non-root)
FROM alpine:3.21
RUN apk add --no-cache ca-certificates && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/services/api-gateway/docs ./docs

RUN chown -R appuser:appgroup /app
USER appuser

CMD ["./main"]
