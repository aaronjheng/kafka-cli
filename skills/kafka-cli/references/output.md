# Output and errors

What stdout and stderr actually contain, what is safe to parse, and why exit `0` is not always
success.

## Admin output is ANSI-colored tables

Every `cluster`, `topic list/describe/get-offsets`, `group list/describe/offsets/lag` and
`cluster config describe` renders a Lipgloss table. Borders use box-drawing characters and the
output contains ANSI SGR sequences **even when stdout is piped**:

```text
\e[38;5;240m┌──────────┬──────────────────────┬────────────────────┐\e[m
\e[38;5;240m│\e[m \e[1mTopic\e[m    \e[38;5;240m│\e[m \e[1mNumber of Partitions\e[m \e[38;5;240m│\e[m …
\e[38;5;240m├──────────┼──────────────────────┼────────────────────┤\e[m
\e[38;5;240m│\e[m orders   \e[38;5;240m│\e[m 3                    \e[38;5;240m│\e[m …
\e[38;5;240m└──────────┴──────────────────────┴────────────────────┘\e[m
Total topics: 1
```

- Color is unconditional: `NO_COLOR`, `TERM=dumb` and `CLICOLOR=0` do not remove it, and there is
  no flag to disable it. Strip ANSI before any parsing or matching.
- There is no `--json`, `--csv`, `--raw` or machine-readable mode. Cells are left-aligned and
  padded with spaces; sensitive broker config values are masked as `*****`; empty values print `-`.
- When stdout is not a terminal, tables keep one row per line at natural width — the friendliest
  form to parse. When it is a terminal, the table is clamped to the terminal width and cells wrap
  onto continuation lines, which breaks line-oriented parsing. Redirect to a file, do not page.
- Some output is outside the tables and interleaved with them: `Cluster ID: …`, `Topic: …`,
  `Consumer Group: …`, `Brokers`, `Configs`/`Config Overrides`, `Partitions`, `Total topics: N`.
  A pure table-to-TSV conversion drops those lines; collect them separately if needed.

### Parse the tables

Tested recipe: strip ANSI, then turn each bordered row into TSV (header row included, border rows
dropped):

```bash
kafka topic list |
  perl -pe 's/\e\[[0-9;]*m//g' |
  awk -F'│' 'NF>2 { row=""; for (i=2;i<NF;i++){ gsub(/^[ \t]+|[ \t]+$/,"",$i); row = row (i>2 ? "\t" : "") $i } print row }'
```

```text
Topic	Number of Partitions	Replication Factor
orders	3	1
payments	6	3
```

The header is the first emitted row. If `perl` is unavailable, `sed $'s/\033\\[[0-9;]*m//g'` does
the same strip in bash, but the Perl one-liner is the portable choice on macOS and Linux. Do not
split on `│` before stripping ANSI: the last column carries a reset sequence after the value.
Cells that themselves contain `│` (rare; not produced by these tables) would split incorrectly.

## Non-table output

| Command | stdout | stderr |
|---|---|---|
| `topic produce` | nothing, even on success | send errors, if any |
| `topic consume` | `[<partition>] <value>` per message | reader failures (`level=ERROR …`) |
| `topic create`, `topic alter` | nothing | nothing |
| `topic delete`, `group delete` | nothing | `level=INFO` lines per item deleted, plus errors |
| `cluster describe` | `Cluster ID: …` + brokers table | — |
| `topic describe` | header lines + partitions table | — |
| `group lag` | summary table + `Partitions` + tables | — |
| `config cat` | `# <path>` + raw config contents | — |
| `version` | `<version> (<go> <os>/<arch>) (commit: …)` | — |
| `--help` | usage text, no ANSI | — |

Log lines use Go `slog` text format, for example
`time=2026-11-14T10:00:00.000+08:00 level=INFO msg="Delete topic" topic=my-topic`. Treat any
non-`Error:` stderr line as diagnostic noise, not as failure.

## Exit codes

| Situation | Exit | Where the message is |
|---|---|---|
| success, including silent writes and an empty table | `0` | — |
| broker returned an error for a command | `1` | stderr, `Error: …` |
| connection/TLS/SASL/SSH setup failure | `1` | `Error: provideAdmin error: kafka.New error: sarama.NewClient error: …` |
| topic not found | `1` | `Error: admin.DescribeTopic error: topic not found: X` |
| consumer group not found / dead | `1` | `Error: admin.DescribeConsumerGroup error: consumer group not found: X` |
| config file problem | `1` | `Error: load config: load config: …` |
| unknown flag | `1` | `Error: unknown flag: --x` |
| unknown command | `1` | `Error: unknown command "x" for "kafka"` + `Run 'kafka --help' for usage.` |
| `--help`, `version`, `completion` | `0` | stdout |
| **consume** reader failure (bad partition, lost broker) | **`0`** once all readers stop | stderr `level=ERROR …`, stdout may be empty |
| **produce** send failure | `1` (after consuming all stdin) | `Error: producer.SendMessage error: …`; earlier messages may be written |

Practical rules:

- Branch on the exit code first; then look for the last `Error:` line. Error chains are prefixed by
  layers (`provideAdmin error: kafka.New error: sarama.NewClient error: …`), and the real cause is
  at the end.
- Do not assume exit `0` from `consume` means messages flowed. A reader error is only logged, and
  no data plus a `level=ERROR` stderr line is a failure.
- Do not assume exit `0` from `produce` means anything was sent: an empty stdin produces nothing
  and still succeeds.
- `topic delete` / `group delete` report what they did only through `level=INFO` lines on stderr;
  the exit code tells you whether the whole batch completed.
- `--help` output is plain text without ANSI, so it is safe to grep for flags.

## Streaming and hanging commands

- `topic consume` never ends by itself: no message count, no timeout, no seek. Kill it or wrap it
  in `timeout` (exit `124` on timeout). Partitions are read in parallel, so output order is only
  guaranteed within a partition.
- `topic produce` reads stdin until EOF. Run it with a pipe or a file; with an inherited open stdin
  it waits forever. Blank lines are skipped.
- The CLI installs no signal handler. `Ctrl-C`/`SIGTERM` aborts immediately: already acknowledged
  synchronous sends are durable, the in-flight line may be lost, and cleanup (closing the producer
  and SSH tunnel) is skipped.
- Long outputs: `cluster config describe --all` can be very large; write it to a file and filter
  after stripping ANSI rather than dumping everything into the conversation.
