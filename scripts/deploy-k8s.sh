#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd)"

NAMESPACE="${K8S_NAMESPACE:-sumdyk}"
IMAGE_NAMESPACE="${IMAGE_NAMESPACE:-${GITHUB_REPOSITORY_OWNER:-mathalama}/nektokz}"
REGISTRY="${REGISTRY:-ghcr.io}"
IMAGE_TAG="${IMAGE_TAG:-prod}"
ALLOWED_ORIGINS="${ALLOWED_ORIGINS:-}"
JWT_SECRET="${JWT_SECRET:-}"
INTERNAL_TOKEN="${INTERNAL_TOKEN:-}"
USER_DB_URL="${USER_DB_URL:-}"
CHAT_DB_URL="${CHAT_DB_URL:-}"
MODERATION_DB_URL="${MODERATION_DB_URL:-}"
REDIS_URL="${REDIS_URL:-}"
GHCR_USERNAME="${GHCR_USERNAME:-}"
GHCR_TOKEN="${GHCR_TOKEN:-}"
TEMP_DIR="$(mktemp -d)"

cleanup() {
  rm -rf "${TEMP_DIR}"
}

trap cleanup EXIT

kubectl_cmd() {
  sudo k3s kubectl "$@"
}

image_for() {
  local name="$1"
  printf '%s/%s/%s:%s' "${REGISTRY}" "${IMAGE_NAMESPACE}" "${name}" "${IMAGE_TAG}"
}

render_manifest() {
  local src="$1"
  local dst="$2"
  local image="$3"

  sed "s|__IMAGE__|${image}|g" "${src}" > "${dst}"
}

echo "Applying namespace and shared resources..."

sudo ufw allow 5432/tcp
sudo ufw allow 6379/tcp

kubectl_cmd apply -f "${ROOT_DIR}/infrastructure/k8s/namespace.yaml"

if [ -n "${GHCR_USERNAME}" ] && [ -n "${GHCR_TOKEN}" ]; then
  kubectl_cmd create secret docker-registry ghcr-creds \
    -n "${NAMESPACE}" \
    --docker-server="${REGISTRY}" \
    --docker-username="${GHCR_USERNAME}" \
    --docker-password="${GHCR_TOKEN}" \
    --docker-email="github-actions@users.noreply.github.com" \
    --dry-run=client -o yaml | kubectl_cmd apply -f -

  kubectl_cmd patch serviceaccount default \
    -n "${NAMESPACE}" \
    --type=merge \
    -p '{"imagePullSecrets":[{"name":"ghcr-creds"}]}'
fi

if [ -z "${ALLOWED_ORIGINS}" ]; then
  echo "ALLOWED_ORIGINS is required for deployment." >&2
  exit 1
fi

if [ -z "${JWT_SECRET}" ] || [ -z "${INTERNAL_TOKEN}" ] || [ -z "${USER_DB_URL}" ] || [ -z "${CHAT_DB_URL}" ] || [ -z "${MODERATION_DB_URL}" ] || [ -z "${REDIS_URL}" ]; then
  echo "JWT_SECRET, INTERNAL_TOKEN, USER_DB_URL, CHAT_DB_URL, MODERATION_DB_URL and REDIS_URL are required for deployment." >&2
  exit 1
fi

kubectl_cmd create configmap sumdyk-config \
  -n "${NAMESPACE}" \
  --from-literal=APP_ENV=production \
  --from-literal=LOG_LEVEL=info \
  --from-literal=ALLOWED_ORIGINS="${ALLOWED_ORIGINS}" \
  --from-literal=DEV_ALLOWED_ORIGINS="http://localhost:3000,http://127.0.0.1:3000" \
  --from-literal=USER_SERVICE_URL="http://user-service:8081" \
  --from-literal=MATCHMAKING_SERVICE_URL="http://matchmaking-service:8082" \
  --from-literal=CHAT_SERVICE_URL="http://chat-service:8083" \
  --from-literal=MODERATION_SERVICE_URL="http://moderation-service:8084" \
  --from-literal=NOTIFICATION_SERVICE_URL="http://notification-service:8085" \
  --dry-run=client -o yaml | kubectl_cmd apply -f -

kubectl_cmd create secret generic sumdyk-secrets \
  -n "${NAMESPACE}" \
  --from-literal=JWT_SECRET="${JWT_SECRET}" \
  --from-literal=INTERNAL_TOKEN="${INTERNAL_TOKEN}" \
  --from-literal=USER_DB_URL="${USER_DB_URL}" \
  --from-literal=CHAT_DB_URL="${CHAT_DB_URL}" \
  --from-literal=MODERATION_DB_URL="${MODERATION_DB_URL}" \
  --from-literal=REDIS_URL="${REDIS_URL}" \
  --dry-run=client -o yaml | kubectl_cmd apply -f -

declare -a deployments=(
  "api-gateway"
  "user-service"
  "matchmaking-service"
  "chat-service"
  "moderation-service"
  "notification-service"
  "frontend"
)

for deployment in "${deployments[@]}"; do
  src="${ROOT_DIR}/infrastructure/k8s/${deployment}.yaml"
  dst="${TEMP_DIR}/${deployment}.yaml"
  image="$(image_for "${deployment}")"

  echo "Applying ${deployment} -> ${image}"
  render_manifest "${src}" "${dst}" "${image}"
  kubectl_cmd apply -f "${dst}"
done

echo "Waiting for rollouts..."
for deployment in "${deployments[@]}"; do
  kubectl_cmd -n "${NAMESPACE}" rollout status "deployment/${deployment}" --timeout=600s
done

echo "Deployment completed successfully."
