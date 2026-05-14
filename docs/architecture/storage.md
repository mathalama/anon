# Storage

Storage splits into three classes:

- durable application data in PostgreSQL
- ephemeral coordination in Redis
- object or external storage for assets if the product adds them later

If a feature needs persistence, document:

- who owns the data
- how it is migrated
- how it is backed up
- how it is restored
