# NektoKZ

Anonymous chat and matchmaking platform.

## Quick Start

1. Clone the repo and enter it.
2. Copy `.env.example` to `.env` and fill in local values.
3. Start infrastructure:

```bash
make up-infra
```

4. Run database migrations:

```bash
make migrate
```

5. Start the application services:

```bash
make up-services
```

6. Start the frontend:

```bash
cd apps/web
npm install
npm run dev
```

## Navigation

- [docs/README.md](docs/README.md) - wiki home
- [docs/overview/vision.md](docs/overview/vision.md) - what the product is
- [docs/architecture/system-design.md](docs/architecture/system-design.md) - system layout and service boundaries
- [docs/development/setup.md](docs/development/setup.md) - local development workflow
- [docs/infrastructure/docker.md](docs/infrastructure/docker.md) - container runtime and compose setup
- [docs/security/secrets.md](docs/security/secrets.md) - secret handling and trust boundaries
- [docs/runbooks/production-incident.md](docs/runbooks/production-incident.md) - incident response

## Code Layout

- `services/` - Go microservices
- `apps/web/` - Next.js frontend
- `libs/` - shared libraries and protobufs
- `infrastructure/` - Docker, Kubernetes, Terraform, Ansible, and monitoring
- `docs/` - engineering wiki
