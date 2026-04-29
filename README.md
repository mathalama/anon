# Nektokz Microservices Ecosystem

[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![Next.js](https://img.shields.io/badge/Next.js-15-black?style=flat-square&logo=next.js)](https://nextjs.org/)
[![Docker](https://img.shields.io/badge/Docker-24.0+-2496ED?style=flat-square&logo=docker)](https://www.docker.com/)
[![Terraform](https://img.shields.io/badge/Terraform-1.0+-7B42BC?style=flat-square&logo=terraform)](https://www.terraform.io/)

Nektokz is a high-performance, modular microservices ecosystem designed for real-time communication, matchmaking, and social interaction. Built with **Go** (Backend) and **Next.js** (Frontend).

---

## Prerequisites

Ensure you have the following installed:

- **Go 1.25+**
- **Node.js 20+** & **npm**
- **Docker & Docker Compose v2+**
- **Make** (for Windows, ensure `make` is in your PATH or use PowerShell scripts directly)

---

## Quick Start (Backend)

### 1. Configure Environment
```bash
cp .env.example .env
```

### 2. Launch Infrastructure
Start PostgreSQL, Redis, and Coturn:
```bash
make up-infra
```

### 3. Run Migrations
```bash
make migrate
```

### 4. Start Services
```bash
make up-services
```

---

## Frontend Setup

The frontend is a modern Next.js application located in the `/frontend` directory.

### 1. Install Dependencies
```bash
cd frontend
npm install
```

### 2. Configure Environment
```bash
cp .env.local.example .env.local  # If available, or use defaults
```

### 3. Run Development Server
```bash
npm run dev
```
The frontend will be available at [http://localhost:3000](http://localhost:3000).

---

## Project Structure

```text
├── api-gateway/          # Central entry point & routing
├── user-service/         # Authentication & profile management
├── chat-service/         # Real-time WebSocket messaging
├── matchmaking-service/  # User matching algorithms
├── moderation-service/   # Content filtering & reporting
├── notification-service/ # Internal alerts
├── frontend/             # Next.js web application
├── docker/               # Dockerfiles & Compose layers
├── terraform/            # OCI Infrastructure as Code
└── monitoring/           # Prometheus & Grafana
```

---

## Makefile Commands

| Command | Description |
| :--- | :--- |
| `make up` | Start all backend containers |
| `make down` | Stop and remove all containers |
| `make migrate` | Run database migrations |
| `make logs-services` | Tail logs for Go services |
| `make logs-infra` | View logs for DBs & Redis |

---

## Dashboards

- **API Gateway**: [http://localhost:8080](http://localhost:8080)
- **Frontend**: [http://localhost:3000](http://localhost:3000)
- **Grafana**: [http://localhost:3000](http://localhost:3000) (Note: Port collision with Frontend if run on same host; check docker-compose)
- **Prometheus**: [http://localhost:9090](http://localhost:9090)

---

## Deployment

Refer to [DEPLOYMENT.md](file:///c:/Users/Admin/Desktop/sumdyk/DEPLOYMENT.md) for OCI cloud deployment details.

---

## License

This project is for demonstration purposes. All rights reserved.
