# ADR-002: Use Redis

## Context

The project needs fast shared state for queues, fanout, and coordination.

## Decision

Use Redis for ephemeral coordination and caching.

## Consequences

- low-latency queue and pub/sub support
- simple local and containerized deployment
- Redis must not be treated as the only source of truth for durable data
