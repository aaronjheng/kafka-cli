# Configuration

`kafka` takes all connection settings from a YAML file. There are no broker, TLS, SASL or SSH
flags; the only global flags are `-f/--config <path>` and `-c/--cluster <profile>`.

## Search order and selection

1. `-f/--config <path>` — must exist and parse.
2. `./kafka.yaml` in the current working directory.
3. `$XDG_CONFIG_HOME/kafka/kafka.yaml` — Linux `~/.config/kafka/kafka.yaml`, macOS
   `~/Library/Application Support/kafka/kafka.yaml`.
4. Nothing found → built-in config: `default_cluster: default`, one cluster `default` with
   broker `127.0.0.1:9092`.

The first match wins; `./kafka.yaml` deliberately shadows the user config (a project-local config
is a real risk when running `kafka` from a repository). The config directory is created (`0755`)
even when no file exists.

`-c/--cluster <name>` selects a profile; without it, `default_cluster` is used. `-c` here means a
config profile, **not** cluster mode and not a controller address.

Configuration is loaded for every command except `version` and `completion` (and `--help`).
A broken config therefore breaks `topic list` even before any network I/O.

## Schema

```yaml
default_cluster: local

clusters:
  local:
    brokers:
      - 127.0.0.1:9092
```

| Field | Required | Meaning |
|---|---|---|
| `default_cluster` | yes (when a file exists) | profile used when `-c` is not given |
| `clusters` | yes | map of profile name to cluster config |

Cluster fields:

| Field | Required | Meaning |
|---|---|---|
| `brokers` | yes | non-empty list of `host:port` bootstrap brokers |
| `tls` | no | TLS block; its presence enables TLS |
| `sasl` | no | SASL block; its presence enables SASL authentication |
| `ssh` | no | SSH tunnel block; its presence routes all broker connections through the bastion |

`tls`:

| Field | Meaning |
|---|---|
| `insecure` | skip certificate and hostname verification (self-signed/dev only) |
| `cafile` | PEM CA file added as a trusted root (required for a private CA) |

There is no client-certificate/key support, so mutual TLS cannot be configured here.

`sasl`:

| Field | Meaning |
|---|---|
| `mechanism` | `PLAIN`, `SCRAM-SHA-256` or `SCRAM-SHA-512` (spelling matters at the wire level) |
| `username`, `password` | credentials; plaintext in the file |

`ssh`:

| Field | Required | Meaning |
|---|---|---|
| `host` | yes | bastion host; in `external` mode also a `Host` alias from `~/.ssh/config` |
| `port` | `builtin`: yes (see below) | bastion port; in `external` mode leave unset to inherit |
| `user` | no | bastion user; in `external` mode leave unset to inherit |
| `identity_file` | no | private key; default keys tried: `~/.ssh/id_ed25519`, `id_ecdsa`, `id_dsa`, `id_rsa` |
| `mode` | no | `builtin` (default, in-process SSH client) or `external` (system `ssh -N -D`) |
| `options` | no | extra `ssh` arguments passed through as-is; `external` mode only |

Gotcha: `builtin` mode does **not** default `port` to 22. Omitting it dials `host:0` and fails
with `ssh.Dial error: dial tcp <host>:0: connect: connection refused`. Set `port: 22` explicitly.
`external` mode omits port and lets `ssh`/`~/.ssh/config` decide.

## Examples

Minimal local cluster:

```yaml
default_cluster: local
clusters:
  local:
    brokers:
      - 127.0.0.1:9092
```

Private CA, no SASL:

```yaml
default_cluster: staging
clusters:
  staging:
    brokers:
      - broker-1.example.com:9093
      - broker-2.example.com:9093
    tls:
      cafile: /path/to/ca.crt
```

SASL with SCRAM, private CA:

```yaml
default_cluster: prod
clusters:
  prod:
    brokers:
      - broker-1.example.com:9093
      - broker-2.example.com:9093
      - broker-3.example.com:9093
    tls:
      cafile: /etc/kafka/ca.crt
    sasl:
      mechanism: SCRAM-SHA-512
      username: svc-kafka-cli
      password: secret
```

SSH tunnel, built-in client (key authentication, port explicit):

```yaml
default_cluster: internal
clusters:
  internal:
    brokers:
      - 10.0.0.11:9092
      - 10.0.0.12:9092
    ssh:
      host: bastion.example.com
      port: 22
      user: deploy
      identity_file: ~/.ssh/id_ed25519
```

SSH tunnel through the system `ssh` (uses `~/.ssh/config`, `ssh-agent`, `ProxyJump`):

```yaml
default_cluster: internal
clusters:
  internal:
    brokers:
      - 10.0.0.11:9092
    ssh:
      host: bastion-alias
      mode: external
```

With an alias, only `host` and `mode` are needed. In `external` mode the tunnel is
`ssh -N -D 127.0.0.1:<port> …`; an `INFO starting external ssh tunnel command="ssh …"` line is
logged to stderr (no secrets, but it shows the identity-file path).

## Behaviour and limits

- **Builtin mode** is key-only: no password, keyboard-interactive, `~/.ssh/config` alias or
  passphrase-protected key (an encrypted key fails to parse). It also ignores `known_hosts`
  (`InsecureIgnoreHostKey`), so it does not verify the bastion's host key — use `external` mode if
  host-key verification matters.
- **External mode** requires `ssh` on `PATH`, and gives normal `ssh` semantics: config aliases,
  agent, host-key checks and prompts. If there is no controlling terminal, prompts cannot be
  answered.
- One tunnel is opened per command and shared by all broker connections (seed brokers and nodes
  discovered from metadata). The bastion must be able to reach every broker address that Kafka
  advertises, not just the bootstrap list; unreachable advertised addresses fail with a dial
  error inside the tunnel.
- TLS hostname verification uses the broker address from the config/metadata. If brokers advertise
  IPs but the certificate is for a DNS name, TLS verification fails; there is no SNI override.
- `sasl.mechanism` is handed to sarama and the broker as written, so use the exact supported
  spelling (`PLAIN`, `SCRAM-SHA-256`, `SCRAM-SHA-512`); a wrong case fails the handshake. The
  SCRAM hash client itself is selected case-insensitively.
- No environment variables override any of this. Secrets stay in plaintext in the file; use
  `chmod 600` and never print the file into logs or a conversation.
- `config cat` prints the resolved file: `# <path as given>` followed by its raw contents. With
  built-in defaults (no file), it fails: `Error: os.Open error: open : no such file or directory`.

## Configuration error messages

`load config` is wrapped twice by the root command, so these strings really do repeat:

| Error | Cause |
|---|---|
| `Error: load config: load config: viper.ReadInConfig error: open …: no such file or directory` | `-f` file does not exist |
| `Error: load config: load config: viper.ReadInConfig error: While parsing config: …` | invalid YAML |
| `Error: load config: load config: config validation error: default cluster is not set` | `default_cluster` missing or blank |
| `Error: load config: load config: config validation error: no clusters configured` | `clusters` missing or empty |
| `Error: load config: load config: config validation error: cluster profile not found: X` | `default_cluster` names a missing profile |
| `Error: load config: load config: config validation error: cluster "X": kafka config is required` | a profile is `null` |
| `Error: load config: load config: config validation error: cluster "X": broker is required` | empty or blank `brokers` |
| `Error: provideAdmin error: cfg.Cluster error: cluster profile not found: X` | `-c` names a missing profile (admin commands) |
| `Error: clusterConfig error: cluster profile not found: X` | same for `topic consume` / `topic produce` |

The last two are not config-validation errors: the file loaded fine, but the selected name does not
exist. Run `kafka config cat` to see the available profile names.

## Connection failures at a glance

All admin commands wrap connection setup as `provideAdmin error: kafka.New error:
sarama.NewClient error: …`. The tail is the real cause:

- `kafka: client has run out of available brokers to talk to: dial tcp …: connect: connection refused`
  — nothing listening, wrong port, firewall, or the SSH tunnel never came up.
- `… x509: certificate signed by unknown authority` — missing or wrong `tls.cafile`.
- `… SASL Authentication failed: …` / `invalid credentials` — wrong mechanism, user or password.
- `… configureSSH error: ssh.NewTunnel error: …` — bastion unreachable or key rejected.
