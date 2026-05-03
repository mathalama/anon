# Log-Based Troubleshooting Automation

## Overview

Centralized logging infrastructure using **Loki + Promtail** for collecting, storing, and analyzing container logs with automated pattern detection.

## Components

### 1. Loki
- **Purpose**: Log aggregation and indexing
- **Storage**: BoltDB (local filesystem)
- **Retention**: 30 days
- **Config**: [monitoring/loki/loki-config.yml](loki-config.yml)

### 2. Promtail
- **Purpose**: Log collector and shipper
- **Sources**: Docker container logs via socket
- **Config**: [monitoring/promtail/promtail-config.yml](../promtail/promtail-config.yml)

### 3. Grafana Integration
- **Datasource**: Loki automatically configured as Loki datasource
- **Exploration**: Use Grafana's Logs Explore UI (bottom left menu)
- **Dashboards**: Create custom dashboards with log panels

## Alert Rules

Automated detection of critical patterns:

### Database Issues
- `DatabaseConnectionFailed` - Connection refused/timeout
- `DatabaseQueryTimeout` - Slow/hanging queries

### Service Health
- `ServiceRestartLoop` - Container restarting frequently (>3 times/5min)
- `HealthCheckFailure` - Service health checks failing

### Application Errors
- `HighErrorRateInLogs` - >100 ERROR logs in 5 minutes
- `StackTraceDetected` - Panic/crash detected
- `OutOfMemoryError` - OOMKilled or memory allocation failures

### Security
- `AuthenticationFailures` - >20 auth failures in 5 minutes

**Config**: [monitoring/loki/loki-alert-rules.yml](loki-alert-rules.yml)

## Usage

### View Logs in Grafana

1. Open Grafana (http://localhost:3000)
2. Click "Explore" (bottom left)
3. Select "Loki" datasource
4. Use LogQL queries:

```logql
# All Docker container logs
{job="docker"}

# API Gateway errors
{job="docker", container_name="api-gateway"} |= "ERROR"

# Database connection failures
{job="docker"} |= "connection refused"

# Specific service
{job="docker", compose_service="user-service"} 

# Errors from last 1 hour
{job="docker"} |= "error" [1h]

# Error rate for a service
sum(rate({job="docker", container_name="api-gateway"} |= "ERROR" [5m]))
```

### Query Patterns (LogQL)

```logql
# Filter by label
{service="api-gateway"}

# Filter by log content
{job="docker"} |= "ERROR"
{job="docker"} != "DEBUG"

# Regex matching
{job="docker"} |~ "connection.*failed"

# JSON parsing
{job="docker"} | json | status >= 500

# Line format
{job="docker"} | logfmt | level=error

# Metric aggregation
rate({job="docker"} |= "ERROR" [5m])
count_over_time({job="docker"} |= "ERROR" [5m])
```

## Docker Compose

**File**: `docker-compose.monitoring.yml`

```bash
# Start monitoring stack (includes Loki + Promtail)
docker-compose -f docker-compose.infra.yml \
               -f docker-compose.services.yml \
               -f docker-compose.monitoring.yml up -d
```

## Architecture

```
┌─────────────────────────────────────┐
│     Docker Containers               │
│  (api-gateway, user-service, etc)   │
└────────────────┬────────────────────┘
                 │ (logs via socket)
                 ▼
        ┌─────────────────┐
        │   Promtail      │
        │  Log Collector  │
        └────────┬────────┘
                 │ (push logs)
                 ▼
        ┌─────────────────┐
        │     Loki        │
        │  Log Storage    │
        └────────┬────────┘
                 │ (query)
                 ▼
        ┌─────────────────┐
        │    Grafana      │
        │   Visualization │
        │   + Alerting    │
        └─────────────────┘
```

## Benefits

✅ **Centralized Logging** - All container logs in one place
✅ **Pattern Detection** - Auto-detect database failures, restarts, OOM
✅ **Full-Text Search** - Find issues fast with powerful queries
✅ **Grafana Integration** - Explore logs and create dashboards
✅ **Alerting** - Automatic alerts on critical patterns
✅ **Long Retention** - Keep 30 days of logs
✅ **Low Overhead** - BoltDB is lightweight (no external dependencies)

## Performance Tips

1. **High Log Volume**: Consider increasing Loki `ingestion_rate_mb`
2. **Long Queries**: Use time ranges to reduce data scanned
3. **Retention**: Adjust `retention_period` based on disk space
4. **Memory**: For production, use shared storage backend (S3, GCS, etc)

## Troubleshooting

### No logs appearing in Grafana
```bash
# Check Promtail is running and collecting logs
docker logs promtail

# Check Loki is receiving logs
docker logs loki | grep "POST /loki/api/v1/push"

# Verify Grafana datasource connection
# Settings > Data Sources > Loki > Test
```

### High memory usage
```bash
# Reduce log retention
# Loki config: retention_period: 168h  # 1 week instead of 30 days

# Limit ingestion rate
# Loki config: ingestion_rate_mb: 512  # Instead of 1024
```

### Logs not being collected
```bash
# Verify Docker socket is mounted
docker inspect promtail | grep docker.sock

# Check Promtail config
docker logs promtail
```
