# Database Architecture

PostgreSQL is the durable store for data that must survive restarts and
instance replacement.

Use PostgreSQL for:

- user records
- chat metadata
- reports
- other durable application state

Use service-owned migrations so each bounded context keeps its schema changes
close to the code.

Database guidance:

- keep transactional state in PostgreSQL
- keep ephemeral queue state out of PostgreSQL unless it must be durable
- document indexes with the feature they support
