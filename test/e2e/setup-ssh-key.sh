#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SSH_KEY="${SCRIPT_DIR}/ssh_key"

rm -f "${SSH_KEY}" "${SSH_KEY}.pub"

ssh-keygen -t ed25519 -f "${SSH_KEY}" -N "" -C "e2e-test" 2>/dev/null

SSH_PROXY_CONTAINER="ssh-proxy"

for i in $(seq 1 30); do
  if docker exec "${SSH_PROXY_CONTAINER}" test -d /config 2>/dev/null; then
    break
  fi
  sleep 1
done

docker cp "${SSH_KEY}.pub" "${SSH_PROXY_CONTAINER}:/config/ssh_key.pub"

chmod 600 "${SSH_KEY}"

echo "SSH key generated at ${SSH_KEY}"
