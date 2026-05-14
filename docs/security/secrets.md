# Secrets

Secrets must not live in plain text runtime config.

Document here:

- secret ownership
- secret storage
- rotation process
- local example values

Rule of thumb:

- `ConfigMap` for non-sensitive config
- `Secret` or external secret storage for credentials
