FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app

COPY go.work go.work.sum ./
COPY services/moderation-service/go.mod services/moderation-service/go.sum ./services/moderation-service/
COPY libs/shared/go.mod libs/shared/go.sum ./libs/shared/
COPY libs/proto/go.mod libs/proto/go.sum ./libs/proto/

ENV GOPROXY=https://proxy.golang.org,direct
ENV GODEBUG=http2client=0

WORKDIR /app/services/moderation-service
RUN GOWORK=off go mod download

WORKDIR /app
COPY libs/shared ./libs/shared
COPY libs/proto ./libs/proto
COPY services/moderation-service ./services/moderation-service

WORKDIR /app/services/moderation-service
RUN --mount=type=cache,target=/root/.cache/go-build \
    GOWORK=off CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/main ./cmd/main.go

FROM alpine:3.21
RUN apk add --no-cache ca-certificates && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/services/moderation-service/migrations ./migrations

RUN chown -R appuser:appgroup /app
USER appuser

CMD ["./main"]
