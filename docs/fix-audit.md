# Full Fix Audit

## Scope
Repository-wide audit with implementation guidance focused on:
- security
- correctness
- buildability
- deployment hygiene
- service-to-service trust boundaries

## Executive Summary

The project has four high-impact problems:

1. Secrets are stored in a Kubernetes `ConfigMap`.
2. The Go workspace is broken because the proto module replacement points to a non-existent path.
3. Public user lookup leaks more data than it should.
4. Internal service endpoints rely on weak trust assumptions.

I also verified the current build state:
- `go test ./...` at the repo root fails because the workspace/module layout is inconsistent.
- `go test ./...` in each service module fails for the same proto replacement issue.

## Findings

### 1. Secrets are committed in plain text config

Files:
- [`infrastructure/k8s/configmap.yaml`](/home/aikyn/anon/infrastructure/k8s/configmap.yaml)
- [`infrastructure/docker/docker-compose.services.yml`](/home/aikyn/anon/infrastructure/docker/docker-compose.services.yml)
- [`.env.example`](/home/aikyn/anon/.env.example)

Problem:
- `DB_URL` and `REDIS_URL` in the Kubernetes `ConfigMap` contain real-looking credentials.
- A `ConfigMap` is not a secret store.
- Secrets in example files can be copied into production by accident.

Impact:
- credential exposure
- accidental reuse of leaked passwords
- hard-to-audit deployments

Best implementation:
- Move all credential-bearing values to `Secret` or external secret storage.
- Keep only non-sensitive endpoints in `ConfigMap`.
- Make example files clearly fake and non-deployable.
- Rotate credentials after the migration.

### 2. Workspace build is broken

Files:
- [`services/user-service/go.mod`](/home/aikyn/anon/services/user-service/go.mod)
- [`services/api-gateway/go.mod`](/home/aikyn/anon/services/api-gateway/go.mod)
- [`services/chat-service/go.mod`](/home/aikyn/anon/services/chat-service/go.mod)
- [`services/matchmaking-service/go.mod`](/home/aikyn/anon/services/matchmaking-service/go.mod)
- [`services/moderation-service/go.mod`](/home/aikyn/anon/services/moderation-service/go.mod)
- [`services/notification-service/go.mod`](/home/aikyn/anon/services/notification-service/go.mod)
- [`go.work`](/home/aikyn/anon/go.work)

Problem:
- The service modules resolve `github.com/mathalama/nektokz/proto` via `../proto`.
- In this repository, the module exists at `libs/proto`, not `services/proto`.
- The entire build fails before tests can run.

Impact:
- no CI signal
- no reliable local validation
- high chance of drift between generated code and module references

Best implementation:
- Standardize the proto module location and reference it consistently.
- Prefer one of these two patterns:
  - keep the proto module in `libs/proto` and update every `replace` to point there
  - or move the module to `services/proto` and keep the repo structure aligned
- Add a root-level validation target that checks module resolution before test execution.

### 3. User profile lookup leaks unnecessary data

Files:
- [`services/api-gateway/internal/handler/user_handler.go`](/home/aikyn/anon/services/api-gateway/internal/handler/user_handler.go)
- [`services/user-service/internal/delivery/grpc/handler.go`](/home/aikyn/anon/services/user-service/internal/delivery/grpc/handler.go)
- [`services/user-service/internal/domain/user.go`](/home/aikyn/anon/services/user-service/internal/domain/user.go)

Problem:
- `GetByID` is exposed through the gateway without an ownership check.
- The gRPC handler returns a value derived from `Email` as `Username`.
- The response shape mixes public profile data with private account data.

Impact:
- privacy leak
- user enumeration risk
- future auth logic becomes harder to reason about

Best implementation:
- Split user data into:
  - public profile fields
  - private account fields
- Make `GetByID` return only public data unless the caller is the owner or a privileged service.
- Introduce explicit response DTOs for public profile reads.

### 4. Internal trust is too weak

Files:
- [`services/api-gateway/internal/middleware/auth.go`](/home/aikyn/anon/services/api-gateway/internal/middleware/auth.go)
- [`services/user-service/internal/delivery/http/handler.go`](/home/aikyn/anon/services/user-service/internal/delivery/http/handler.go)
- [`services/chat-service/internal/delivery/http/handler.go`](/home/aikyn/anon/services/chat-service/internal/delivery/http/handler.go)
- [`services/moderation-service/internal/delivery/http/handler.go`](/home/aikyn/anon/services/moderation-service/internal/delivery/http/handler.go)

Problem:
- The gateway strips spoofed `X-User-ID`, but internal handlers still trust that header.
- Internal endpoints rely on a shared token rather than explicit service identity.
- The system assumes the Docker/Kubernetes network is fully trusted.

Impact:
- lateral movement risk
- service impersonation if network controls are bypassed
- hard-to-debug authorization behavior

Best implementation:
- Use mTLS or service identity at the network layer for internal traffic.
- Keep `X-Internal-Token` only as a temporary defense-in-depth layer.
- Make internal endpoints refuse requests unless they come through an authenticated service path.
- Consider moving privileged actions behind gRPC with service credentials instead of plain HTTP headers.

## Recommended Design

### Secrets and config

Separate values into three groups:
- public runtime config: ports, feature flags, non-sensitive URLs
- private runtime config: passwords, tokens, API keys
- deploy-time values: domain names, ingress hosts, region settings

Implementation pattern:
- `ConfigMap` for public values
- `Secret` for credentials
- `.env.example` for illustrative placeholders only

### Auth model

Use three distinct trust levels:
- end-user auth: JWT access token
- refresh flow: short handler that only issues new tokens
- service auth: mTLS or signed internal identity, not a generic shared header

### User profile model

Split the API contract into:
- `PublicUserProfile`
- `PrivateUserProfile`
- `AccountState`

That lets you safely expose:
- display name
- avatar
- gender if intended
- matchmaking-facing metadata

And keep private:
- email
- password hash
- internal device linkage
- moderation metadata

### Build layout

The workspace should have one unambiguous source of truth:
- one proto module path
- one `go.work` entry
- one dependency reference per service

Add a check that fails if:
- any `replace` points to a missing directory
- any generated protobuf package imports a stale module path

## Fix Order

### Phase 1: make the repository build again

1. Fix proto module references.
2. Re-run `go test ./...` in each module.
3. Confirm `go.work` resolves cleanly.

### Phase 2: remove secret exposure

1. Move database and Redis credentials out of the `ConfigMap`.
2. Update manifests and compose files to use secret injection.
3. Rotate all affected credentials.

### Phase 3: tighten API data exposure

1. Redesign public user reads.
2. Remove accidental email exposure from public lookups.
3. Add tests for ownership and visibility rules.

### Phase 4: harden internal trust

1. Introduce stronger service identity.
2. Keep shared-token checks only as a transitional guard.
3. Restrict internal endpoints to authenticated internal callers.

## Verification Plan

Run after each phase:
- `go test ./...` in every service module
- `go test ./...` in `libs/proto` and `libs/shared`
- config scan for secrets in `ConfigMap`
- grep for stale `../proto` replacements
- manual API checks:
  - `/api/v1/users/me`
  - `/api/v1/users/{id}`
  - internal ban/report endpoints

## Acceptance Criteria

- repository builds successfully
- no real secrets exist in public config files
- user lookup returns only intended public fields
- internal endpoints have an explicit trust boundary
- build/test commands are stable in CI and locally

## Notes From Current Test Run

I ran the module tests during this audit. They currently fail because the codebase points to a missing `services/proto` module path.

That is not a flaky test issue. It is a structural dependency problem and should be fixed before any further verification can be trusted.
