# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .
# Copy migrations if needed (user-service uses them)
COPY --from=builder /app/migrations ./migrations

EXPOSE 8081

CMD ["./main"]
