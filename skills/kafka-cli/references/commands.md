# Commands

The command tree, defaults and behaviours that matter when driving `kafka` from an agent or script.
All connection settings come from the config file (see [config.md](config.md)); only `-f` and `-c`
are global flags.

```text
kafka [ -f <config> ] [ -c <profile> ] <group> <subcommand> [args] [flags]
```

| Group | Subcommands |
|---|---|
| `config` | `cat` |
| `cluster` | `describe`, `config describe` |
| `topic` | `list`, `create`, `alter`, `delete`, `describe`, `get-offsets`, `consume`, `produce` |
| `group` | `list`, `describe`, `offsets`, `lag`, `delete` |
| `version` | — |
| `completion` | `bash`, `zsh`, `fish` |

## `cluster`

```bash
kafka cluster describe
kafka cluster config describe                    # overrides for every broker
kafka cluster config describe --broker 1         # all configs of one broker
kafka cluster config describe --all              # all configs of every broker
```

- `describe` prints `Cluster ID: <id>`, a blank line, `Brokers`, then a table with `ID`, `Address`,
  `Rack` (`-` when unset) and `Type` (`controller` or `broker`). It requires a metadata round trip
  to the controller, so it is also a good connectivity probe.
- `config describe` prints one section per broker: `Broker: <id>`, a blank line, then
  `Config Overrides` (default) or `Configs` (`--all` or a single `--broker`), then a table with
  `Name`, `Value`, `Source`, `Read Only`, `Sensitive`.
  - Default (no `--broker`): every broker, overrides only (`Source` not `default`).
  - `--broker <id>`: one broker, **all** of its configs; a negative value means all brokers.
  - `--all` and `--broker` are mutually exclusive.
  - Empty values print `-`, sensitive values print `*****`, and a broker with no overrides prints
    `No config overrides` instead of a table.
  - `Source` is one of `static-broker`, `dynamic-broker`, `dynamic-default-broker`,
    `dynamic-topic` (topics only), `default`, `unknown`.

## `topic`

### list, create, alter, delete, describe, get-offsets

```bash
kafka topic list
kafka topic create my-topic                                  # 3 partitions, RF 1
kafka topic create my-topic --partitions 6 --replication-factor 3
kafka topic alter my-topic --partitions 6                    # total partitions, not a delta
kafka topic delete my-topic
kafka topic delete topic-a topic-b                           # sequential, stops at first error
kafka topic delete -- --topic-starting-with-a-dash
kafka topic describe my-topic
kafka topic get-offsets my-topic
```

- `list` prints a table (`Topic`, `Number of Partitions`, `Replication Factor`) sorted by name,
  then `Total topics: N`. Internal topics such as `__consumer_offsets` may appear.
- `create` only sets partitions and replication factor. There is no way to set `cleanup.policy`,
  retention or any other topic config, and no pre-check: an existing topic is a broker error.
  Defaults come from this CLI, not the broker: `--partitions 3`, `--replication-factor 1`.
- `alter` sets the **total** partition count via `CreatePartitions`. Kafka cannot reduce partitions,
  and the flag defaults to `-1`, so always pass `--partitions <n>`.
- `delete` sends one delete per topic, sequentially; the first error aborts the rest. Deletion is
  asynchronous in Kafka: metadata may linger, and recreating the same name immediately can fail.
  Use `topic list` to confirm. Each item logs `level=INFO` lines (`Delete topic`,
  `Topic deleted`) to stderr.
- `describe` prints `Topic`, `Partitions`, `Replication Factor`, `Cleanup Policy`, `Size`, then a
  partition table (`Partition`, `Leader`, `Replicas`, `ISR`). `Cleanup Policy` and `Size` are
  `unknown` when the extra `DescribeConfigs`/`DescribeLogDirs` calls fail (permissions, ACLs,
  brokers without log-dir reporting). `Size` sums every replica's on-disk bytes for the topic
  across brokers, so with replication factor 3 it is roughly three times the logical data size;
  it is formatted as `B`/`KB`/`MB`/`GB`, not an exact byte count.
- `get-offsets` prints `Partition`, `Oldest Offset`, `Newest Offset`; the difference is the number
  of messages currently readable in that partition.

### consume

```bash
kafka topic consume my-topic              # all partitions
kafka topic consume my-topic -p 0         # one partition
timeout 30 kafka topic consume my-topic   # always bound it in an agent step
```

- Starts every reader at `OffsetNewest`: only messages produced **after** the reader connects.
  There is no `--from-beginning`, no offset/timestamp seek, no limit and no graceful stop; kill it
  or wrap it in `timeout`.
- Output is `[<partition>] <value>` per line, with the partition number padded to the widest
  partition ID. Keys, headers, timestamps and offsets are not printed.
- `-p/--partition` outside the topic's range, or a topic that disappears, is logged to stderr and
  the command still exits `0` once all readers stop. Do not treat exit `0` as "all good"; inspect
  stderr. Partition listing failure (unknown topic) happens before the readers start and exits `1`.
- Messages arrive from independent partition readers, so output ordering is per-partition only.

### produce

```bash
printf 'hello\nworld\n' | kafka topic produce my-topic
cat messages.ndjson | kafka topic produce my-topic
printf 'k1:v1\nk2:v2\n' | kafka topic produce my-topic --key-separator ':'
```

- One message per stdin line, blank lines skipped, until EOF. A line longer than 1 MiB fails the
  scanner (`Error: scan stdin error: bufio.Scanner: token too long`).
- `--key-separator <s>` splits the line at the **first** occurrence: `a:b:c` → key `a`, value `b:c`.
  A line without the separator becomes a value-only message; the whole line is the value.
- Sequential synchronous sends. A failed send is recorded and production continues; at the end the
  joined errors are returned, so `1` means *some* messages may already be written.
- Nothing is printed on success. There is no `--count`, no file argument, no headers, no
  compression and no timestamp control.
- A missing topic may be auto-created by the broker (`auto.create.topics.enable`), so success does
  not prove the topic was pre-existing; `timed out` style errors point at a broker that rejects
  auto-creation.

## `group`

```bash
kafka group list
kafka group describe my-group
kafka group offsets my-group
kafka group offsets my-group --topic my-topic
kafka group lag my-group
kafka group lag my-group --topic my-topic
kafka group delete my-group
```

- `list` prints `Consumer Group`, `State`, `Protocol Type`, `Protocol`, `Members`, sorted by name.
- `describe` prints `Consumer Group`, `State`, `Protocol Type`, `Protocol`, `Members`, then a
  `Members` table (`Member ID`, `Client ID`, `Client Host`) when the group has members. Groups
  reported with an empty or `Dead` state and no error are treated as not found.
- `offsets` prints `Consumer Group: <g>`, then one table per topic:
  `Partition`, `Current Offset`, `Log End Offset`, `Lag`, `Consumer ID`, `Host`. Partitions with
  no committed offset (offset `< 0`) are skipped entirely. `Consumer ID`/`Host` are filled from
  the current member assignment and stay empty for inactive groups.
- `lag` prints a summary table (`Topic`, `Current Offset`, `Log End Offset`, `Lag`) aggregated
  over the committed partitions, then blank, `Partitions`, then the per-topic partition tables.
  Lag is clamped at zero (`max(logEnd - committed, 0)`).
- `--topic` limits output and requires the topic to exist; without it every topic with committed
  offsets is shown.
- `offsets`/`lag` describe the group first, so they fail for `Dead`/unknown groups and for groups
  that the coordinator cannot describe.
- `delete` deletes one group at a time, sequentially, stopping at the first error; it logs
  `level=INFO` lines to stderr. Kafka rejects deleting a group with active members: stop the
  consumers first, and use `group describe` to confirm `Empty`.

## `config`, `version`, `completion`

```bash
kafka config cat                 # "# <path>" + the raw config file
kafka version                    # 1.0.0 (go… os/arch) (commit: … when built from git)
kafka completion bash|zsh|fish
```

`config cat` prints the path exactly as it was resolved (`-f` value or the discovered file) and
then the file content. It fails when the built-in default config is in effect, because there is no
file to print. `version` and `completion` skip config loading and work with a broken config.

## Risk table

| Command | Risk |
|---|---|
| `topic delete`, `group delete` | irreversible; sequential with partial success on error |
| `topic alter --partitions` | irreversible increase; a decrease is rejected by Kafka |
| `topic create` | creates a topic with this CLI's defaults (3 partitions, RF 1) that may not match cluster policy |
| `topic produce` | writes real data; partially succeeds on error; can auto-create a topic |
| `topic consume` | unbounded stream; no time or count limit |
| `cluster config describe` | read-only, but can expose sensitive settings (masked as `*****` only when the broker flags them) and `--all` output is large |

## Not implemented

There is no support for: ACLs, quotas, SCRAM/ACL user management, topic config changes
(`cleanup.policy`, retention, `max.message.bytes`), partition reassignment, preferred leaders,
consumer-group offset reset/deletion by topic, consumer-group creation, `kafka-configs.sh`-style
dynamic config writes, transactions, headers, compression selection, Schema Registry, Kafka
Connect, ksqlDB, Streams, or any JSON/CSV output. For those, use the Kafka distribution tools or an
admin client, and say so instead of improvising flags here.
