#!/usr/bin/env bash
set -euo pipefail

max_retries=30
interval=3

: "${SSH_KEY_PATH:?}"

echo "Checking SSH..."
for i in $(seq 1 "$max_retries"); do
  if ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
     -i "${SSH_KEY_PATH}" -p 2222 \
     kafkauser@127.0.0.1 echo "ready" 2>/dev/null; then
    echo "SSH is ready"
    exit 0
  fi
  echo "  retry $i..."
  sleep "$interval"
done

echo "SSH did not become ready after $max_retries retries"
exit 1
