#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SSH_KEY="${SCRIPT_DIR}/ssh_key"

rm -f "${SSH_KEY}" "${SSH_KEY}.pub"

ssh-keygen -t ed25519 -f "${SSH_KEY}" -N "" -C "e2e-test" 2>/dev/null

chmod 600 "${SSH_KEY}"

SSH_PUBLIC_KEY_CONTENT="$(cat "${SSH_KEY}.pub")"

echo "SSH_KEY=${SSH_KEY}" >> "${GITHUB_ENV:-/dev/null}"
echo "SSH_PUBLIC_KEY=${SSH_PUBLIC_KEY_CONTENT}" >> "${GITHUB_ENV:-/dev/null}"

echo "SSH key generated at ${SSH_KEY}"
echo "SSH_PUBLIC_KEY set for docker compose"
