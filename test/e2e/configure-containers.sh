#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CERTS_DIR="${SCRIPT_DIR}/certs"

configure_sasl_container() {
  local container="$1"

  docker exec "${container}" bash -c 'cat > /opt/bitnami/kafka/config/client.properties <<EOF
security.protocol=SASL_PLAINTEXT
sasl.mechanism=PLAIN
sasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required username="admin" password="admin-secret";
EOF'
}

configure_tls_container() {
  local container="$1"

  docker exec "${container}" bash -c 'cat > /tmp/client-ssl.properties <<EOF
security.protocol=SSL
ssl.truststore.location=/opt/bitnami/kafka/config/certs/kafka.truststore.jks
ssl.truststore.password=changeit
EOF'
}

configure_sasl_ssl_container() {
  local container="$1"

  docker exec "${container}" bash -c 'cat > /tmp/client-sasl-ssl.properties <<EOF
security.protocol=SASL_SSL
sasl.mechanism=PLAIN
sasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required username="admin" password="admin-secret";
ssl.truststore.location=/opt/bitnami/kafka/config/certs/kafka.truststore.jks
ssl.truststore.password=changeit
EOF'
}

echo "Configuring kafka-sasl container..."
configure_sasl_container kafka-sasl

echo "Configuring kafka-tls container..."
configure_tls_container kafka-tls

echo "Configuring kafka-sasl-ssl container..."
configure_tls_container kafka-sasl-ssl
configure_sasl_ssl_container kafka-sasl-ssl

echo "All containers configured."
