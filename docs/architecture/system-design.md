# System Design

NektoKZ uses a microservices architecture with one public HTTP edge and several
internal services.

High-level path:

```text
Browser / Next.js app
        |
        v
api-gateway  <--- public HTTP entry point and BFF
   |   |   |
   |   |   +--> chat-service (HTTP + WebSocket)
   |   +------> matchmaking-service (HTTP + gRPC)
   +----------> user-service (HTTP + gRPC)
        |
        +------> moderation-service (HTTP)
        +------> notification-service (internal workers)

Shared infra:
- PostgreSQL
- Redis
- NATS JetStream
- Prometheus / Grafana / Jaeger
```

Design rules:

- keep the gateway thin unless a BFF aggregation is intentional
- keep durable data behind the service that owns it
- use gRPC for typed internal contracts
- use WebSocket only where the connection is long-lived and interactive
- use Redis for coordination, not as the only source of truth
