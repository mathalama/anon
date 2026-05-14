# Kubernetes

Kubernetes manifests describe the cluster-style deployment shape.

Document here:

- namespaces
- config maps
- secrets
- deployments
- service exposure

Security note:

- keep credentials out of `ConfigMap`
- use `Secret` or an external secret manager for sensitive values
