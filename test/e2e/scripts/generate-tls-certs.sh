#!/usr/bin/env bash
set -euo pipefail

TLS_DIR=$(mktemp -d)
chmod 755 "${TLS_DIR}"

openssl req -new -x509 -keyout "${TLS_DIR}/ca-key" -out "${TLS_DIR}/ca-cert" \
  -days 365 -passout pass:capass \
  -subj "/CN=TestCA"

keytool -keystore "${TLS_DIR}/kafka.server.keystore.jks" \
  -alias localhost -keyalg RSA -validity 365 \
  -storepass keystorepass -keypass keystorepass \
  -dname "CN=localhost" -genkeypair

keytool -keystore "${TLS_DIR}/kafka.server.keystore.jks" \
  -alias localhost -certreq -file "${TLS_DIR}/server-cert-sign-request" \
  -storepass keystorepass -keypass keystorepass

openssl x509 -req -CA "${TLS_DIR}/ca-cert" -CAkey "${TLS_DIR}/ca-key" \
  -in "${TLS_DIR}/server-cert-sign-request" \
  -out "${TLS_DIR}/server-cert-signed" \
  -days 365 -CAcreateserial -passin pass:capass

keytool -keystore "${TLS_DIR}/kafka.server.keystore.jks" \
  -alias CARoot -importcert -file "${TLS_DIR}/ca-cert" \
  -storepass keystorepass -noprompt

keytool -keystore "${TLS_DIR}/kafka.server.keystore.jks" \
  -alias localhost -importcert -file "${TLS_DIR}/server-cert-signed" \
  -storepass keystorepass -keypass keystorepass -noprompt

keytool -keystore "${TLS_DIR}/kafka.server.truststore.jks" \
  -alias CARoot -importcert -file "${TLS_DIR}/ca-cert" \
  -storepass truststorepass -noprompt

printf "keystorepass" > "${TLS_DIR}/keystore_creds"
printf "keystorepass" > "${TLS_DIR}/key_creds"
printf "truststorepass" > "${TLS_DIR}/truststore_creds"

cp "${TLS_DIR}/ca-cert" "${TLS_DIR}/ca-cert.pem"

echo "tls-certs-path=${TLS_DIR}" >> "$GITHUB_OUTPUT"
echo "ca-cert-path=${TLS_DIR}/ca-cert.pem" >> "$GITHUB_OUTPUT"
