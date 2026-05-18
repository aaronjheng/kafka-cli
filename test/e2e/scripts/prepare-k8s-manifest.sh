#!/usr/bin/env bash
set -euo pipefail

: "${KAFKA_VERSION:?}"
: "${SSH_PUBLIC_KEY_FILE:?}"
: "${SSHD_CONFIG_PATH:?}"
: "${SASL_CONFIG_PATH:?}"
: "${TLS_CERTS_PATH:?}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
MANIFEST_DIR="${PROJECT_DIR}/test/e2e/kafka"

series=$(echo "${KAFKA_VERSION}" | cut -d. -f1-2)
series_template="${MANIFEST_DIR}/pod-${series}.yaml"

if [[ ! -f "${series_template}" ]]; then
	echo "kafka-cli: no pod template found for Kafka ${KAFKA_VERSION} (expected ${series_template})" >&2
	exit 1
fi

output="${PROJECT_DIR}/test/e2e/kafka/pod.yaml"
cp "${series_template}" "${output}"

SSH_PUBLIC_KEY=$(cat "${SSH_PUBLIC_KEY_FILE}")

sed -i "s|PLACEHOLDER_KAFKA_VERSION|${KAFKA_VERSION}|" "${output}"
sed -i "s|PLACEHOLDER_SSH_PUBLIC_KEY|${SSH_PUBLIC_KEY}|" "${output}"
sed -i "s|PLACEHOLDER_SSHD_CONFIG_PATH|${SSHD_CONFIG_PATH}|" "${output}"
sed -i "s|PLACEHOLDER_SASL_CONFIG_PATH|${SASL_CONFIG_PATH}|" "${output}"
sed -i "s|PLACEHOLDER_TLS_CERTS_PATH|${TLS_CERTS_PATH}|" "${output}"