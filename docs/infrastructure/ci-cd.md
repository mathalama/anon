# CI/CD

Document the delivery pipeline here.

Minimum stages should usually be:

- test
- build
- container image creation
- deployment

If the pipeline changes, document:

- what runs on pull request
- what runs on merge
- what runs during deployment

## Current K8s Flow

The production deployment flow is server-side:

1. Terraform creates the VM, network, and security group.
2. `Provision Server` installs Docker, k3s, and the deploy user.
3. `Deploy Base (DB & WebRTC)` updates the compose-based infra layer.
4. `Bootstrap Kubernetes` creates namespace, configmap, and secret.
5. `Deploy Application Stack` brings up the k8s app workloads.
6. Push-based deploy workflows update only the services that changed.
7. `kubectl rollout status` waits for readiness.

This flow avoids external image registries for application images. The
K3s cluster is installed with Docker runtime support so locally built images
are visible to Kubernetes without GHCR.

`workflow_run` connects the first four steps so a successful provision can
automatically continue into infra, bootstrap, and app deployment.
