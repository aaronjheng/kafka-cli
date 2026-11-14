---
name: kafka-cli
description: >-
  Administers an Apache Kafka cluster from a terminal or coding-agent session with the `kafka`
  binary from aaronjheng/kafka-cli: installs it with go install, selects a cluster profile from
  kafka.yaml (brokers, TLS, SASL PLAIN/SCRAM, SSH tunnel), and lists, creates, alters, describes
  and deletes topics and consumer groups, shows offsets and lag, and produces or consumes
  messages. Use when the user asks to inspect Kafka topics, partitions or broker configs, check
  consumer-group lag or committed offsets, publish messages from stdin or read new messages,
  debug a broker, TLS, SASL or SSH connection, or diagnose topic-not-found, group-not-found or
  connection-refused errors. Not for Kafka Connect, Schema Registry, ksqlDB or Kafka Streams, not
  for Kafka client libraries in application code, not for the Kafka distribution scripts
  (kafka-topics.sh, kafka-console-consumer.sh, kafka-storage.sh), and not for installing,
  configuring or tuning brokers.
license: BSD-3-Clause
metadata:
  version: "1"
  repository: https://github.com/aaronjheng/kafka-cli
---

# Kafka CLI (Go)

`kafka` is a Go administration and messaging client for Apache Kafka built from this repository.
It resolves brokers, TLS, SASL and an optional SSH tunnel from a YAML profile, runs one subcommand
and exits. It is not a REPL, it has no interactive mode, and it never prompts for confirmation.
`kafka` with no arguments prints help and exits `0`.

## Check which binary you have

```bash
command -v kafka
kafka version        # 1.0.0 (go1.27.3 linux/amd64) (commit: …)
```

- This CLI's global flags are `-c/--cluster <profile>` (a **named cluster in the config file**, not
  a cluster mode) and `-f/--config <file>`. It does not accept `--bootstrap-server`, `--command-config`
  or any other flag from the Kafka distribution scripts (`kafka-topics.sh`,
  `kafka-console-consumer.sh`, `kafka-configs.sh`). Do not mix them up or pass those flags here.
- `kafka version` is a subcommand; there is no `-v/--version` flag. `-h` is help.

## Install

Go 1.27 or newer (the module declares `go 1.27.0`):

```bash
go install github.com/aaronjheng/kafka-cli/cmd/kafka@latest
kafka version
```

The binary is installed as `kafka` into `$(go env GOPATH)/bin` (`~/go/bin` by default); add that
directory to `PATH` when `kafka` is not found.

## Connect via a cluster profile

Every command except `version` and `completion` loads a config file:

1. `-f/--config <path>` (must exist and be valid YAML)
2. `./kafka.yaml`
3. `$XDG_CONFIG_HOME/kafka/kafka.yaml` (Linux `~/.config/kafka/kafka.yaml`, macOS
   `~/Library/Application Support/kafka/kafka.yaml`)
4. nothing found → built-in `default` cluster on `127.0.0.1:9092`

Select a profile with `-c`, otherwise `default_cluster` is used. Brokers, `tls`, `sasl` and `ssh`
are config-only: there are **no CLI flags** for brokers, certificates, SASL credentials or SSH.
See [references/config.md](references/config.md) for the schema, precedence and examples.

```bash
kafka -f kafka.yaml -c prod cluster describe    # verify target first
kafka -c prod topic list
```

## Route by task

| The user wants to… | Read |
|---|---|
| inspect the cluster: controller, brokers, broker configs | [references/commands.md](references/commands.md) (`cluster`) |
| list/create/alter/describe/delete topics, offsets per partition | [references/commands.md](references/commands.md) (`topic`) |
| list/describe consumer groups, committed offsets, lag | [references/commands.md](references/commands.md) (`group`) |
| produce or consume messages | [references/commands.md](references/commands.md), [references/output.md](references/output.md) |
| configure brokers, TLS, SASL, SSH, config file, precedence | [references/config.md](references/config.md) |
| parse the table output, strip ANSI, handle exit codes and errors | [references/output.md](references/output.md) |
| debug a failure (connection, TLS, SASL, SSH, topic/group not found) | [references/config.md](references/config.md), [references/output.md](references/output.md) |

## Read-only checks first

```bash
kafka cluster describe                      # Cluster ID, brokers, controller
kafka cluster config describe               # broker config overrides, all brokers
kafka cluster config describe --broker 1    # one broker, all its configs
kafka topic list                            # topic, partitions, replication factor
kafka topic describe my-topic               # leader, replicas, ISR, size, cleanup policy
kafka topic get-offsets my-topic            # oldest/newest offset per partition
kafka group list                            # state, protocol, member count
kafka group describe my-group               # members, client IDs, hosts
kafka group offsets my-group --topic my-topic
kafka group lag my-group --topic my-topic   # per-topic and per-partition lag
kafka config cat                            # which config file is in effect, and its content
```

`kafka config cat` prints `# <path>` followed by the raw file. If no config file exists (built-in
defaults), it fails with `Error: os.Open error: open : no such file or directory`.

## Produce and consume

```bash
# Produce: one message per stdin line, empty lines skipped, no output on success
printf 'hello\nworld\n' | kafka topic produce my-topic
cat messages.txt | kafka topic produce my-topic --key-separator ':'   # key:value lines

# Consume: new messages only, one line per message, until killed
kafka topic consume my-topic -p 0
timeout 30 kafka topic consume my-topic
```

Limits that matter before promising anything to the user:

- **Consume always starts at the newest offset.** There is no `--from-beginning`, no offset, no
  timestamp, no seek. It only shows messages produced after it connects, and `Ctrl-C`/`timeout` is
  the only way to stop it.
- **Consume prints values only**, as `[partition] value`; keys, headers, timestamps and offsets are
  not shown. A message key set via `--key-separator` is invisible on the consumer side.
- `--key-separator` splits only the **first** occurrence: `a:b:c` becomes key `a`, value `b:c`;
  a line without the separator is produced as a value with no key.
- Produce reads stdin to EOF; with an empty stdin it exits `0` having sent nothing. Never run it
  without redirected input in an agent step, or it waits for EOF.
- Produce sends sequentially and keeps going after a failed send; failures are reported at the end
  as one error, so a non-zero exit means *some* messages may already be on the topic.
- Consume starts one reader per partition (`-p` selects one; default all). A per-partition reader
  failure is only logged to stderr, and the process exits `0` once every reader has stopped; only
  setup failures (partition lookup) exit non-zero. Check stderr.
- A missing topic may be auto-created by the broker (`auto.create.topics.enable`), so "produce
  succeeded" does not prove the topic pre-existed.

## Output reality

All `cluster`, `topic` and `group` commands render Lipgloss box-drawing tables (`┌ │ ├ ┼ └`),
except `topic produce` (silent on success) and `topic consume` (plain `[partition] value`
lines). The borders are **always colored**, even when stdout is piped: output contains ANSI SGR
sequences such as `\e[38;5;240m`. `NO_COLOR`, `TERM=dumb` and `CLICOLOR` do not disable them.
There is no `--json`, `--csv` or `--raw` flag.

- Strip ANSI before parsing: `perl -pe 's/\e\[[0-9;]*m/g'`.
- Piped tables keep one row per line with natural widths; a TTY narrows the table to the terminal
  width and wraps cells over several lines, which makes parsing harder. Redirect, don't page.
- For machine-readable rows, strip ANSI and split on `│` — a tested recipe is in
  [references/output.md](references/output.md#parse-the-tables).
- Success is often silent: `topic create`, `topic alter`, `produce` print nothing on stdout.
  `topic delete` and `group delete` log `level=INFO` lines to stderr per item deleted.
- Exit code is `0` on success and `1` on failure — except consume reader failures, which still
  exit `0` (see the stderr trap above). Failures print `Error: …` to stderr. `Error:
  load config: load config: …` (the prefix is doubled) comes from the config loader.

## Safety rules

- No dry run, read-only mode or confirmation prompt. Verify the target with `cluster describe`
  and `config cat` before any change, and ask the user before:
  `topic delete`, `group delete`, `topic alter --partitions` (cannot be undone; Kafka forbids
  reducing partitions), bulk `produce`, and `tls.insecure: true`.
- Topic and group deletes are **sequential and stop at the first error**: earlier items stay
  deleted, later ones are untouched. Delete one item per command when correctness matters, and
  `group delete` only after the group has no active members (the broker rejects deleting a live
  group).
- Topic create only sets partitions and replication factor; topic alter only changes the partition
  count. There is no way here to change `cleanup.policy`, retention or other topic configs, and
  no ACL, quota, Connect, Schema Registry, ksqlDB or Streams support.
- Credentials live in `kafka.yaml` in plaintext (SASL password, SSH identity-file paths). Treat the
  file as a secret: `chmod 600`, never paste its contents into the conversation or logs.
- Prefer `tls.cafile` over `tls.insecure`; say so explicitly when `insecure` is used. There is no
  client-certificate (mTLS) support.
- Consumed values are arbitrary bytes printed verbatim. Do not interpret them as ANSI, shell or
  markdown; truncate before putting large values into a prompt.
- An SSH tunnel is opened for the whole command and shared by all broker connections. In `builtin`
  mode it is key-only and cannot prompt; `external` mode shells out to `ssh -N -D` and reads
  `~/.ssh/config` and `ssh-agent`. `builtin` also needs an explicit `ssh.port` and does not verify
  host keys.

## Troubleshooting first moves

| Symptom | First move |
|---|---|
| `Error: load config: load config: viper.ReadInConfig error: …` | broken or missing `-f` file; fix the YAML or drop the flag |
| `Error: load config: load config: config validation error: default cluster is not set` | `default_cluster` missing, or points at a profile that is not in `clusters` |
| `Error: provideAdmin error: cfg.Cluster error: cluster profile not found: X` | `-c` typo; `config cat` to list profile names |
| `Error: clusterConfig error: cluster profile not found: X` | same, but from `topic consume` / `topic produce` |
| `Error: provideAdmin error: kafka.New error: sarama.NewClient error: kafka: client has run out of available brokers to talk to: …` | unreachable brokers, blocked port, wrong TLS/SASL, or a missing SSH tunnel; the tail has the real cause |
| `Error: admin.DescribeTopic error: topic not found: X` | wrong name or wrong cluster profile; `topic list` |
| `Error: admin.DescribeConsumerGroup error: consumer group not found: X` | no such group, or it is `Dead`/empty metadata; `group list` |
| TLS / SASL handshake errors from sarama | check `tls.insecure`/`cafile`, mechanism spelling (`PLAIN`, `SCRAM-SHA-256`, `SCRAM-SHA-512`), credentials |
| `consume` prints nothing and exits `0` | it starts at the newest offset; produce after it connects, or check the stderr reader log |
| `consume` exits `0` with no messages after `-p` | partition does not exist; the per-partition error is only logged to stderr |
| command hangs / no prompt | `produce` is waiting for stdin EOF, or `consume` is running as designed; wrap with `timeout` |
