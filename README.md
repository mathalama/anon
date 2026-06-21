# Mathalama Talk

<p align="left">
  <img src="https://img.shields.io/badge/Next.js-000000?style=for-the-badge&logo=nextdotjs&logoColor=white" alt="Next.js" />
  <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker" />
  <img src="https://img.shields.io/badge/PostgreSQL-4169E1?style=for-the-badge&logo=postgresql&logoColor=white" alt="PostgreSQL" />
  <img src="https://img.shields.io/badge/Redis-DC382D?style=for-the-badge&logo=redis&logoColor=white" alt="Redis" />
  <img src="https://img.shields.io/badge/NATS-0080FF?style=for-the-badge&logo=nats-io&logoColor=white" alt="NATS" />
  <img src="https://img.shields.io/badge/Kubernetes-326CE5?style=for-the-badge&logo=kubernetes&logoColor=white" alt="Kubernetes" />
  <img src="https://img.shields.io/badge/Ansible-EE0000?style=for-the-badge&logo=ansible&logoColor=white" alt="Ansible" />
  <img src="https://img.shields.io/badge/Terraform-7B42BC?style=for-the-badge&logo=terraform&logoColor=white" alt="Terraform" />
  <img src="https://img.shields.io/badge/Prometheus-E6522C?style=for-the-badge&logo=prometheus&logoColor=white" alt="Prometheus" />
  <img src="https://img.shields.io/badge/Grafana-F46800?style=for-the-badge&logo=grafana&logoColor=white" alt="Grafana" />
  <img src="https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge" alt="License: MIT" />
</p>

**Mathalama Talk** is a modern, secure, and fast platform for instant anonymous communication.

The service lets users find conversation partners based on gender and common interests in a fraction of a second. After a successful match, users can chat in private rooms or start a real-time voice call. All activity is completely private and requires no registration, email, or phone numbers. A session is created automatically based on a temporary device identifier.

---

## Code Layout

The repository is a monorepo with the following structure:

* `apps/web/` : Next.js frontend application.
* `services/` : Go microservices:
  * `api-gateway` : Unified entry point routing API and WebSockets.
  * `user-service` : User session and profile management.
  * `chat-service` : Room management and WebRTC signaling.
  * `matchmaking-service` : Fast matchmaking queue logic.
  * `moderation-service` : Report handling and screening.
  * `notification-service` : Dispatching internal events.
* `libs/` : Shared Go libraries and generated Protobuf files.
* `infrastructure/` : Deployment configurations for Docker, Kubernetes, Ansible, Terraform, and monitoring (Prometheus, Grafana, Loki).
* `docs/` : Engineering wiki and technical documentation.

---

## Navigation

* [docs/README.md](docs/README.md) : Engineering wiki home page
* [docs/overview/vision.md](docs/overview/vision.md) : Product vision and goals
* [docs/architecture/system-design.md](docs/architecture/system-design.md) : System architecture design
* [docs/development/setup.md](docs/development/setup.md) : Local development setup guide
* [docs/infrastructure/docker.md](docs/infrastructure/docker.md) : Container runtime configurations
* [docs/security/secrets.md](docs/security/secrets.md) : Security guidelines and secrets management
* [docs/runbooks/production-incident.md](docs/runbooks/production-incident.md) : Incident response instructions

---

## Quick Start

### 1. Prepare Environment
Clone the repository and create your local configuration file:
```bash
cp .env.example .env
```

### 2. Start Infrastructure
Run the database, message broker, and routing infrastructure (Postgres, Redis, NATS, coturn):
```bash
# Using make utility:
make up-infra

# Or directly with Docker Compose:
docker compose up -d postgres redis nats coturn
```

### 3. Run Database Migrations
Apply schemas to Postgres databases:
```bash
# Using make utility:
make migrate

# Or directly with Docker Compose:
docker compose run --rm migrate-user
docker compose run --rm migrate-chat
docker compose run --rm migrate-moderation
```

### 4. Start Application Services
Build and start all backend services and the frontend client:
```bash
# Using make utility:
make up-services

# Or directly with Docker Compose:
docker compose up -d api-gateway user-service matchmaking-service chat-service moderation-service notification-service web
```
The web application will be accessible at: **http://localhost:3002** (mapped to port 3000 inside the container).

---

## License

This project is licensed under the **MIT License**. Check [LICENSE](LICENSE) for details.
