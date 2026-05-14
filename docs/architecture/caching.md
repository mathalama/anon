# Caching

Redis is used for fast, short-lived coordination:

- matchmaking queues
- room occupancy tracking
- pub/sub fanout across instances
- rate limiting

Caching guidance:

- set explicit TTLs for ephemeral data
- do not rely on Redis as the only source of truth for durable state
- document cache keys and invalidation rules alongside the feature
