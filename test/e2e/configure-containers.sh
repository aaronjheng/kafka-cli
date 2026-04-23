#!/usr/bin/env bash
set -euo pipefail

configure_sasl_container() {
  local container="$1"
  docker exec "${container}" bash -c 'cat > /tmp/client-sasl.properties <<EOF
security.protocol=SASL_PLAINTEXT
sasl.mechanism=PLAIN
sasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required username="admin" password="admin-secret";
EOF'
}

configure_tls_container() {
  local container="$1"
  docker exec "${container}" bash -c 'cat > /tmp/client-ssl.properties <<EOF
security.protocol=SSL
ssl.truststore.location=/etc/kafka/secrets/kafka.truststore.jks
ssl.truststore.password=changeit
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

echo "Configuring kafka-sasl..."
configure_sasl_container kafka-sasl

echo "Configuring kafka-tls..."
configure_tls_container kafka-tls

echo "Configuring kafka-sasl-ssl..."
configure_tls_container kafka-sasl-ssl
configure_sasl_ssl_container kafka-sasl-ssl

echo "All containers configured."
