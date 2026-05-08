# SRE Strategy: Project Sumdyk

## 1. Service Level Indicators (SLIs)
We track the following metrics for all microservices:
- **Availability**: Ratio of successful requests (non-5xx) to total requests.
- **Latency**: Time taken to process a request (95th percentile).
- **Error Rate**: Percentage of requests resulting in 5xx errors.
- **Saturation**: CPU and Memory utilization of service instances.

## 2. Service Level Objectives (SLOs)
Our target performance levels are:
- **Availability**: >= 99.0% per month.
- **Latency**: 95% of requests processed in < 200ms.
- **Error Rate**: < 1% of total requests.
- **Saturation**: < 80% average CPU/Memory usage.

## 3. Monitoring Stack
- **Prometheus**: Metrics collection and alerting.
- **Grafana**: Visualization and dashboarding.
- **Alertmanager**: Handling alert notifications.
- **Loki/Promtail**: Log aggregation.

## 4. Incident Response Plan
1. **Detection**: Alerts triggered in Prometheus and sent to Slack/Email.
2. **Triaging**: SRE on-call verifies the impact using Grafana dashboards.
3. **Mitigation**: Scale up services, restart pods, or rollback last deployment.
4. **Resolution**: Fix the root cause in the code or infrastructure.
5. **Postmortem**: Document the incident, root cause, and preventive actions.

## 5. Capacity Planning
- **Vertical Scaling**: Increase CPU/RAM limits for database (Postgres) and heavy services (Matchmaking).
- **Horizontal Scaling**: Increase `replicas` in Kubernetes/Swarm for stateless services.
- **Load Testing**: Regularly run `load-tests/` scripts to identify bottlenecks.
