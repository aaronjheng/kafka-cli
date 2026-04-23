#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CERTS_DIR="${SCRIPT_DIR}/certs"

configure_tls_container() {
  local container="$1"

  docker exec "${container}" mkdir -p /etc/kafka/secrets

  docker cp "${CERTS_DIR}/kafka.keystore.jks" "${container}:/etc/kafka/secrets/kafka.keystore.jks"
  docker cp "${CERTS_DIR}/kafka.truststore.jks" "${container}:/etc/kafka/secrets/kafka.truststore.jks"
  docker cp "${CERTS_DIR}/ca.crt" "${container}:/etc/kafka/secrets/ca.crt"

  docker exec "${container}" bash -c 'cat > /tmp/client-ssl.properties <<EOF
security.protocol=SSL
ssl.truststore.location=/etc/kafka/secrets/kafka.truststore.jks
ssl.truststore.password=changeit
EOF'
}

configure_sasl_container() {
  local container="$1"

  docker exec "${container}" bash -c 'cat > /tmp/client-sasl.properties <<EOF
security.protocol=SASL_PLAINTEXT
sasl.mechanism=PLAIN
sasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required username="admin" password="admin-secret";
EOF'
}

configure_sasl_ssl_container() {
  local container="$1"

  docker exec "${container}" bash -c 'cat > /tmp/client-sasl-ssl.properties <<EOF
security.protocol=SASL_SSL
sasl.mechanism=PLAIN
sasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required username="admin" password="admin-secret";
ssl.truststore.location=/etc/kafka/secrets/kafka.truststore.jks
ssl.truststore.password=changeit
EOF'
}

echo "Configuring kafka-tls container..."
configure_tls_container kafka-tls

echo "Configuring kafka-sasl container..."
configure_sasl_container kafka-sasl

echo "Configuring kafka-sasl-ssl container..."
configure_tls_container kafka-sasl-ssl
configure_sasl_ssl_container kafka-sasl-ssl

echo "All containers configured."
