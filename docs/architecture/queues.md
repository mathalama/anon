# Queues

The system uses async queues and event fanout to avoid coupling request latency
to downstream side effects.

Current patterns:

- Redis-backed matchmaking queues
- NATS JetStream for async events and notifications

Queue guidance:

- keep queue ownership explicit
- document retry and timeout behavior
- keep workers idempotent where possible
