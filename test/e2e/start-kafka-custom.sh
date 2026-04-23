#!/usr/bin/env bash
set -euo pipefail

# Custom startup for apache/kafka with TLS/SASL_SSL.
# Bypasses the buggy /etc/kafka/docker/configure script.

PROPS_FILE="/etc/kafka/server.properties"

# Determine the inter-broker listener name from KAFKA_LISTENERS.
# For a single-node KRaft cluster with only one non-controller listener,
# we set inter.broker.listener.name to that listener.
INTER_BROKER_LISTENER=""
if [ -n "${KAFKA_LISTENERS:-}" ]; then
  # Extract the first non-CONTROLLER listener name
  for listener in ${KAFKA_LISTENERS//,/ }; do
    name="${listener%%://*}"
    if [ "${name}" != "CONTROLLER" ]; then
      INTER_BROKER_LISTENER="${name}"
      break
    fi
  done
fi

cat > "${PROPS_FILE}" <<PROPS
node.id=1
process.roles=broker,controller
controller.quorum.voters=1@$(hostname):9093
controller.listener.names=CONTROLLER
listeners=${KAFKA_LISTENERS}
advertised.listeners=${KAFKA_ADVERTISED_LISTENERS}
listener.security.protocol.map=${KAFKA_LISTENER_SECURITY_PROTOCOL_MAP}
offsets.topic.replication.factor=1
transaction.state.log.replication.factor=1
transaction.state.log.min.isr=1
group.initial.rebalance.delay.ms=0
auto.create.topics.enable=false
PROPS

if [ -n "${INTER_BROKER_LISTENER}" ]; then
  echo "inter.broker.listener.name=${INTER_BROKER_LISTENER}" >> "${PROPS_FILE}"
fi

if [ -n "${KAFKA_SASL_ENABLED_MECHANISMS:-}" ]; then
  cat >> "${PROPS_FILE}" <<PROPS
sasl.enabled.mechanisms=${KAFKA_SASL_ENABLED_MECHANISMS}
sasl.mechanism.inter.broker.protocol=${KAFKA_SASL_MECHANISM_INTER_BROKER_PROTOCOL:-PLAIN}
sasl.mechanism.controller.protocol=${KAFKA_SASL_MECHANISM_CONTROLLER_PROTOCOL:-PLAIN}
${KAFKA_SASL_JAAS_CONFIG_LINE:-}
PROPS
fi

if [ -n "${KAFKA_SSL_KEYSTORE_LOCATION:-}" ]; then
  cat >> "${PROPS_FILE}" <<PROPS
ssl.keystore.location=${KAFKA_SSL_KEYSTORE_LOCATION}
ssl.keystore.password=${KAFKA_SSL_KEYSTORE_PASSWORD}
ssl.key.password=${KAFKA_SSL_KEY_PASSWORD}
ssl.truststore.location=${KAFKA_SSL_TRUSTSTORE_LOCATION}
ssl.truststore.password=${KAFKA_SSL_TRUSTSTORE_PASSWORD}
ssl.client.auth=${KAFKA_SSL_CLIENT_AUTH:-none}
PROPS
fi

CLUSTER_ID="${KAFKA_CLUSTER_ID:-$(/opt/kafka/bin/kafka-storage.sh random-uuid)}"
LOG_DIR="/var/lib/kafka/data"

if [ ! -f "${LOG_DIR}/meta.properties" ]; then
  /opt/kafka/bin/kafka-storage.sh format -t "${CLUSTER_ID}" -c "${PROPS_FILE}"
fi

exec /opt/kafka/bin/kafka-server-start.sh "${PROPS_FILE}"
