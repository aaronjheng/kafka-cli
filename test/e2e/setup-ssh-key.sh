#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SSH_KEY="${SCRIPT_DIR}/ssh_key"
SSH_PUB_KEY_FILE="${SCRIPT_DIR}/ssh_public_key"

rm -f "${SSH_KEY}" "${SSH_KEY}.pub" "${SSH_PUB_KEY_FILE}"

ssh-keygen -t ed25519 -f "${SSH_KEY}" -N "" -C "e2e-test" 2>/dev/null

cp "${SSH_KEY}.pub" "${SSH_PUB_KEY_FILE}"
chmod 600 "${SSH_KEY}"

echo "SSH key generated at ${SSH_KEY}"
echo "SSH public key at ${SSH_PUB_KEY_FILE}"
