# ADR-001: Use PostgreSQL

## Context

The project needs durable transactional storage for user, chat, and moderation
data.

## Decision

Use PostgreSQL as the primary durable database.

## Consequences

- strong transactional guarantees
- familiar operational model
- additional schema and migration discipline required
