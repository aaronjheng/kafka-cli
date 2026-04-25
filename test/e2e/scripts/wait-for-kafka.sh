#!/usr/bin/env bash
set -euo pipefail

max_retries=30
interval=5

echo "Checking kafka..."
for i in $(seq 1 "$max_retries"); do
  if podman exec kafka-ssh-kafka-ssh-test /opt/kafka/bin/kafka-topics.sh \
     --bootstrap-server 127.0.0.1:9092 --list 2>/dev/null; then
    echo "kafka is ready"
    exit 0
  fi
  echo "  retry $i..."
  sleep "$interval"
done

echo "kafka did not become ready after $max_retries retries"
exit 1
