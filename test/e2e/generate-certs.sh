#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CERTS_DIR="${SCRIPT_DIR}/certs"

rm -rf "${CERTS_DIR}"
mkdir -p "${CERTS_DIR}"

CA_KEY="${CERTS_DIR}/ca.key"
CA_CERT="${CERTS_DIR}/ca.crt"

openssl req -new -x509 -keyout "${CA_KEY}" -out "${CA_CERT}" -days 365 -nodes \
  -subj "/CN=TestCA" 2>/dev/null

KAFKA_KEY="${CERTS_DIR}/kafka.key"
KAFKA_CSR="${CERTS_DIR}/kafka.csr"
KAFKA_CERT="${CERTS_DIR}/kafka.crt"

openssl req -new -keyout "${KAFKA_KEY}" -out "${KAFKA_CSR}" -nodes \
  -subj "/CN=localhost" 2>/dev/null

cat > "${CERTS_DIR}/ext.cnf" <<'EOF'
subjectAltName=DNS:localhost,IP:127.0.0.1
extendedKeyUsage=serverAuth,clientAuth
EOF

openssl x509 -req -in "${KAFKA_CSR}" -CA "${CA_CERT}" -CAkey "${CA_KEY}" \
  -CAcreateserial -out "${KAFKA_CERT}" -days 365 -extfile "${CERTS_DIR}/ext.cnf" 2>/dev/null

openssl pkcs12 -export -in "${KAFKA_CERT}" -inkey "${KAFKA_KEY}" \
  -out "${CERTS_DIR}/kafka.p12" -password pass:changeit 2>/dev/null

keytool -importkeystore -destkeystore "${CERTS_DIR}/kafka.keystore.jks" \
  -srckeystore "${CERTS_DIR}/kafka.p12" -srcstoretype PKCS12 \
  -srcstorepass changeit -deststorepass changeit -noprompt 2>/dev/null

keytool -import -alias ca -file "${CA_CERT}" \
  -keystore "${CERTS_DIR}/kafka.truststore.jks" \
  -storepass changeit -noprompt 2>/dev/null

chmod 644 "${CERTS_DIR}"/*

echo "Certificates generated in ${CERTS_DIR}"
