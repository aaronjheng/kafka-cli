#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SSH_KEY="${SCRIPT_DIR}/ssh_key"
AUTHORIZED_KEYS="${SCRIPT_DIR}/ssh_authorized_keys"

rm -f "${SSH_KEY}" "${SSH_KEY}.pub" "${AUTHORIZED_KEYS}"

ssh-keygen -t ed25519 -f "${SSH_KEY}" -N "" -C "e2e-test" 2>/dev/null

cp "${SSH_KEY}.pub" "${AUTHORIZED_KEYS}"
chmod 600 "${SSH_KEY}"
chmod 644 "${AUTHORIZED_KEYS}"

echo "SSH key generated at ${SSH_KEY}"
echo "Authorized keys at ${AUTHORIZED_KEYS}"
