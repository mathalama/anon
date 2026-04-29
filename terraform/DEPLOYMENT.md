# Project Deployment and Infrastructure Guide

This document provides a comprehensive guide for deploying and managing the Nektokz microservices ecosystem.

## 1. Prerequisites
Before deployment, ensure you have the following installed:
*   Terraform (>= 1.0.0)
*   Docker and Docker Compose (v2 or higher)
*   OCI CLI (configured with appropriate credentials)
*   Git

## 2. Infrastructure Provisioning (Terraform)
The infrastructure is hosted on Oracle Cloud Infrastructure (OCI). To provision the resources:

1.  Ensure you are in the `terraform/` directory.
2.  Initialize the workspace:
    ```bash
    terraform init
    ```
3.  Preview the changes:
    ```bash
    terraform plan
    ```
4.  Apply the configuration:
    ```bash
    terraform apply -auto-approve
    ```
5.  Note the `public_ip` outputted by Terraform. This is your server address.

## 3. Server Configuration
The server is automatically provisioned using a `user_data` script (`setup.sh`) which installs Docker, Docker Compose, and configures the necessary firewall rules.

### Manual Firewall Setup (if required)
Ensure the following ports are open in the Cloud Console:
*   80: API Gateway (Public traffic)
*   81: Nginx Proxy Manager (Admin UI)
*   3000: Grafana Dashboards
*   9090: Prometheus UI
*   22: SSH Access

## 4. Application Deployment
The application is containerized using Docker Compose. The deployment is split into three layers:

### 4.1 Infrastructure Layer
Start the databases and messaging systems:
```bash
docker compose -f docker-compose.infra.yml up -d
```

### 4.2 Application Services Layer
Deploy the Go-based microservices:
```bash
docker compose -f docker-compose.services.yml up -d
```

### 4.3 Monitoring Layer
Deploy the observability stack:
```bash
docker compose -f docker-compose.monitoring.yml up -d
```

## 5. Service Overview
*   **API Gateway**: `http://<public_ip>:8080`
*   **Prometheus**: `http://<public_ip>:9090`
*   **Grafana**: `http://<public_ip>:3000`
*   **Nginx Proxy Manager**: `http://<public_ip>:81`

## 6. Database Migrations
Migrations are handled automatically by each service upon startup. The services wait for the database to be ready and then execute SQL scripts located in their respective `/migrations` directories.

## 7. Troubleshooting
### Viewing Logs
To inspect service logs for errors:
```bash
docker compose -f docker-compose.services.yml logs -f <service_name>
```

### Database Access
To access the PostgreSQL database directly:
```bash
docker exec -it nektokz-postgres psql -U mathalama -d users_db
```

### Monitoring Targets
Check `http://<public_ip>:9090/targets` to ensure all microservices are being successfully scraped by Prometheus.
