# Kafka CLI

A lightweight command-line tool for managing Apache Kafka clusters. It supports topic and consumer group administration, message producing/consuming, TLS/SASL authentication, and SSH tunneling.

> **Trademark Notice:** Apache Kafka and Kafka are registered trademarks of the Apache Software Foundation.

## Installation

```shell
go install github.com/aaronjheng/kafka-cli/cmd/kafka@latest
```

If you need the newest `master` commit immediately (without relying on branch-resolution cache), install by resolved commit SHA:

```shell
go install github.com/aaronjheng/kafka-cli/cmd/kafka@$(git ls-remote https://github.com/aaronjheng/kafka-cli.git refs/heads/master | cut -f1)
```

## Configuration

See [kafka.example.yaml](contrib/kafka.example.yaml) for reference.

### Config File Search Order

1. Path specified by `--config` / `-f` flag
2. `./kafka.yaml` (current directory)
3. `$XDG_CONFIG_HOME/kafka/kafka.yaml`
   - Linux: `~/.config/kafka/kafka.yaml`
   - macOS: `~/Library/Application Support/kafka/kafka.yaml`

If no config file is found, a default configuration is used with a `default` cluster pointing to `127.0.0.1:9092`.

### Config Fields

| Field | Description |
|-------|-------------|
| `default_cluster` | Default cluster name used when `--cluster` is not specified |
| `clusters` | Map of cluster name to cluster config |

Each cluster supports the following fields:

| Field | Required | Description |
|-------|----------|-------------|
| `brokers` | Yes | List of Kafka broker addresses (e.g. `host:port`) |
| `tls` | No | TLS configuration |
| `sasl` | No | SASL authentication configuration |
| `ssh` | No | SSH proxy/tunnel configuration |

#### `tls`

| Field | Description |
|-------|-------------|
| `insecure` | Skip TLS certificate verification (for self-signed/dev clusters) |
| `cafile` | Path to CA certificate file (PEM format) |

#### `sasl`

| Field | Description |
|-------|-------------|
| `mechanism` | SASL mechanism: `PLAIN`, `SCRAM-SHA-256`, or `SCRAM-SHA-512` |
| `username` | Authentication username |
| `password` | Authentication password |

#### `ssh`

| Field | Required | Description |
|-------|----------|-------------|
| `host` | Yes | SSH bastion host address |
| `port` | Yes | SSH port (e.g. `22`) |
| `user` | Yes | SSH user name |
| `identity_file` | No | Path to SSH private key. If not specified, the following default keys are tried in order: `~/.ssh/id_ed25519`, `~/.ssh/id_ecdsa`, `~/.ssh/id_dsa`, `~/.ssh/id_rsa` |

## Usage

All commands support the following global flags:

- `--config` / `-f` — Specify config file path
- `--cluster` / `-c` — Specify cluster name (defaults to `default_cluster` in config)

### `kafka config cat`

Print the effective configuration file contents.

```shell
kafka config cat
```

### `kafka cluster describe`

Show cluster overview (controller ID, number of brokers/topics, and broker details).

```shell
kafka cluster describe
```

### `kafka topic list`

List all topics in the cluster.

```shell
kafka topic list
```

### `kafka topic create TOPIC`

Create a new topic.

```shell
# Create a topic with default settings (3 partitions, replication factor 1)
kafka topic create my-topic

# Specify partitions and replication factor
kafka topic create my-topic --partitions 6 --replication-factor 3
```

### `kafka topic alter TOPIC`

Alter the partition count of a topic.

```shell
kafka topic alter my-topic --partitions 6
```

### `kafka topic delete TOPIC [TOPIC...]`

Delete one or more topics.

```shell
kafka topic delete my-topic
kafka topic delete topic-a topic-b
```

### `kafka topic describe TOPIC`

Show details of a topic (partitions, replicas, ISR, etc.).

```shell
kafka topic describe my-topic
```

### `kafka topic get-offsets TOPIC`

Show the oldest and newest offsets for each partition of a topic.

```shell
kafka topic get-offsets my-topic
```

### `kafka topic consume TOPIC`

Consume messages from a topic in real time. By default all partitions are consumed; use `--partition` / `-p` to consume a specific partition.

> **Note:** The consumer starts from the latest offset (`LastOffset`), not the beginning of the topic.

```shell
# Consume all partitions
kafka topic consume my-topic

# Consume a specific partition
kafka topic consume my-topic -p 0
```

### `kafka topic produce TOPIC`

Produce messages to a topic by reading from stdin. Empty lines are skipped. Press `Ctrl+D` to finish.

```shell
# Interactive input
kafka topic produce my-topic
hello world
foo bar
^D

# Pipe data from a file
cat messages.txt | kafka topic produce my-topic

# Produce key-value messages using a separator
kafka topic produce my-topic --key-separator ":"
key1:value1
key2:value2
^D
```

### `kafka group list`

List all consumer groups in the cluster.

```shell
kafka group list
```

### `kafka group describe GROUP`

Show details of a consumer group (members, offsets, lag, etc.).

```shell
kafka group describe my-group
```

### `kafka group delete GROUP [GROUP...]`

Delete one or more consumer groups.

```shell
kafka group delete my-group
```

### `kafka version`

Print version information.

```shell
kafka version
```

### `kafka completion [bash|zsh|fish|powershell]`

Generate shell completion script. See `kafka completion --help` for setup instructions.

## License

Kafka CLI is licensed under the [BSD-3-Clause License](https://opensource.org/licenses/BSD-3-Clause). See [LICENSE](LICENSE) for more details.
