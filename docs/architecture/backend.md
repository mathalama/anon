# Backend Architecture

Backend services are organized around bounded contexts:

- `api-gateway`
- `user-service`
- `matchmaking-service`
- `chat-service`
- `moderation-service`
- `notification-service`

Common backend layers:

- delivery or handlers for transport
- usecase or service layer for business logic
- repository for storage
- client for outbound calls
- config for environment wiring

Communication patterns:

- public HTTP through the gateway
- internal HTTP where a service exposes a simple boundary
- gRPC for typed service-to-service calls
- WebSocket for chat and signaling
