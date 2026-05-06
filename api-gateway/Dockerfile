FROM golang:1.25-alpine AS builder
WORKDIR /app

# Copy work and mod files for all services to cache dependencies
COPY go.work go.work.sum ./
COPY api-gateway/go.mod api-gateway/go.sum ./api-gateway/
COPY chat-service/go.mod chat-service/go.sum ./chat-service/
COPY matchmaking-service/go.mod matchmaking-service/go.sum ./matchmaking-service/
COPY moderation-service/go.mod moderation-service/go.sum ./moderation-service/
COPY notification-service/go.mod notification-service/go.sum ./notification-service/
COPY user-service/go.mod user-service/go.sum ./user-service/
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY proto/go.mod proto/go.sum ./proto/

RUN go mod download

# Copy the rest
COPY . .

WORKDIR /app/api-gateway
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/api-gateway/docs ./docs
CMD ["./main"]
