# Nektokz Microservices Ecosystem

## Overview
Nektokz is a modular, high-performance microservices-based application designed for real-time communication and matchmaking. The system is built using Go, PostgreSQL, Redis, and is orchestrated via Docker Compose and Terraform for automated infrastructure management.

## System Architecture
The project follows a decoupled microservices architecture to ensure scalability, fault isolation, and independent maintainability.

### Core Services
*   **API Gateway**: The central entry point for all client requests, handling routing and security.
*   **User Service**: Manages user lifecycle, authentication, and profile data.
*   **Chat Service**: Facilitates real-time messaging using WebSockets with persistent message storage.
*   **Matchmaking Service**: Connects users based on specific criteria using low-latency Redis caching.
*   **Moderation Service**: Implements safety protocols, word filtering, and user reporting.
*   **Notification Service**: Handles internal system alerts and cross-service communication.

## Technology Stack
*   **Backend**: Go 1.25 with `chi` router.
*   **Database**: PostgreSQL 16 (Relational) and Redis 7 (Cache).
*   **WebRTC**: Coturn for STUN/TURN media traversal.
*   **Infrastructure**: Terraform (Oracle Cloud Infrastructure).
*   **Containerization**: Docker & Docker Compose.
*   **Monitoring**: Prometheus (Metrics) and Grafana (Visualization).

## Infrastructure & Deployment

### Terraform
The infrastructure is provisioned as code using Terraform. It manages the Compute Instance (VM), Virtual Cloud Network (VCN), and Security Rules for the cloud environment.

### Docker Orchestration
The system is divided into three logical layers:
1.  **Infrastructure Layer**: PostgreSQL, Redis, Coturn.
2.  **Service Layer**: All Go-based microservices.
3.  **Monitoring Layer**: Prometheus, Grafana, Node Exporter.

## Monitoring and Observability
All services expose a `/metrics` endpoint for Prometheus scraping.
*   **Prometheus**: Collects system and application-level metrics.
*   **Grafana**: Provides real-time dashboards for service health and performance monitoring.
*   **Health Checks**: Each service implements a custom `/health` endpoint to monitor dependency status (e.g., database connectivity).

## Deployment Guide
Refer to the `DEPLOYMENT.md` file for step-by-step instructions on provisioning the cloud environment and launching the application stack.

## Development
To run the project locally:
1.  Initialize the infrastructure: `docker compose -f docker-compose.infra.yml up -d`
2.  Start the services: `docker compose -f docker-compose.services.yml up -d`
3.  Access the API Gateway at `http://localhost:8080`

## Database Migrations
Migrations are automatically applied on service startup. Ensure that SQL scripts are placed in the respective `migrations/` directory of each service.

## License
This project is for academic and professional demonstration purposes. All rights reserved.
