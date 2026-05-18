#!/usr/bin/env bash
set -euo pipefail

SSH_DIR=$(mktemp -d)
chmod 755 "${SSH_DIR}"

ssh-keygen -t ed25519 -f "${SSH_DIR}/id_ed25519" -N ""

echo "ssh-key-path=${SSH_DIR}/id_ed25519" >> "$GITHUB_OUTPUT"
echo "ssh-key-dir=${SSH_DIR}" >> "$GITHUB_OUTPUT"