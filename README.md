# logr

JSON log beautifier and filter for the terminal. Auto-detects your log format, colors output, and lets you filter by level, time, field, or substring — no syntax to learn.

```
14:23:05.123   INFO  GET /users                       requestId=abc123  latency=42ms
14:23:05.891   WARN  DB pool exhausted                service=users     attempt=2
14:23:07.002  ERROR  Payment gateway timeout          requestId=xyz789  orderId=99
```

## Why logr

Most log tools are format-specific. `pino-pretty` only reads Pino. Winston transports only help Winston. `jq` works on anything but you need to know the schema upfront — `.msg // .message // .Message` just to get the message field.

logr auto-detects the format and normalizes everything. Pipe it any JSON log and get readable, colored output with no flags required.

It's the tool you reach for when you don't know — or don't care — what format the logs are in.

**Strongest use cases:**
- Switching between projects using different loggers (Pino, Zap, Serilog, Log4j2, Monolog...)
- `kubectl logs my-pod | logr` or `aws logs get-log-events ... | logr` — ugly JSON becomes readable instantly
- `logr --follow app.log --level error` during an incident — no jq gymnastics
- Your internal app uses non-standard field names? Define a custom format once in config and never think about it again

## Install

**macOS / Linux**
```sh
curl -fsSL https://raw.githubusercontent.com/AgusRdz/logr/main/install.sh | sh
```

**Windows (PowerShell)**
```powershell
irm https://raw.githubusercontent.com/AgusRdz/logr/main/install.ps1 | iex
```

**Manual** — download the binary for your platform from the [releases page](https://github.com/AgusRdz/logr/releases), place it in your `PATH`.

**Update**
```sh
logr update
```

## Usage

```sh
# Pipe from any command
cat app.log | logr
kubectl logs my-pod | logr
aws logs get-log-events ... | logr

# Read a file
logr app.log

# Filter by level
logr app.log --level error
logr app.log --level warn,error

# Filter by time window
logr app.log --since 30m
logr app.log --since 1h --until 30m
logr app.log --since 2025-04-12T14:00

# Filter by field value
logr app.log --field requestId=abc123
logr app.log --field service=payments --field status=500

# Substring search
logr app.log --contains "payment failed"

# Control output
logr app.log --compact               # one line per entry
logr app.log --fields ts,level,msg   # show only these fields
logr app.log --hide pid,hostname     # hide specific fields
logr app.log --raw                   # pass lines through, only filter

# Follow mode (like tail -f)
logr --follow app.log
logr --follow app.log --level error

# Stats summary
logr app.log --stats
logr app.log --stats --since 1h

# Force a specific format
logr app.log --format zap
logr app.log --format serilog
logr app.log --format myapp          # user-defined custom format
```

## Flags

| Flag | Description |
|---|---|
| `--level` | Filter by level: `debug`, `info`, `warn`, `error`, `fatal` (comma-separated) |
| `--since` | Show entries after: `10m`, `1h`, `7d`, or `2025-04-12T14:00` |
| `--until` | Show entries before (same format as `--since`) |
| `--field` | Filter by `key=value` (repeatable) |
| `--contains` | Filter entries containing a substring |
| `--format` | Force format: see [supported formats](#supported-formats) or a custom name from config |
| `--follow` | Follow file like `tail -f` |
| `--stats` | Print stats summary instead of streaming |
| `--compact` | One-line output per entry |
| `--raw` | Pass through original lines, only filter |
| `--fields` | Show only these fields (comma-separated) |
| `--hide` | Hide specific fields (comma-separated) |
| `--no-color` | Disable color output |
| `--color` | Force color even when piped |

## Supported formats

logr auto-detects your log format from the first few lines. You can also force a format with `--format <name>`.

### Node.js

| Format | `--format` name | Detection |
|---|---|---|
| **Pino** | `pino` | `v` + `pid` fields, numeric levels (30=INFO, 40=WARN...) |
| **Bunyan** | `bunyan` | `v` + `name` + `pid`, ISO8601 `time` string |
| **Winston** | `winston` | `level` + `message` + `timestamp` fields |

### Go

| Format | `--format` name | Detection |
|---|---|---|
| **Zap** | `zap` | `ts` (float) + `caller` (file:line) + `msg` |
| **Logfmt** | `logfmt` | `key=value` pairs with recognized level/msg/time fields |

### Java / JVM

| Format | `--format` name | Detection |
|---|---|---|
| **Log4j2** | `log4j` or `log4j2` | `timeMillis` + `loggerName` fields |
| **Spring Boot** | `springboot` or `spring-boot` | `@timestamp` + `level` + `message` (logstash-logback-encoder) |

### .NET

| Format | `--format` name | Detection |
|---|---|---|
| **Serilog (CLEF)** | `serilog` | `@t` (timestamp) + `@mt` (message template) |

### PHP

| Format | `--format` name | Detection |
|---|---|---|
| **Monolog** | `monolog` | `level_name` + `channel` + `datetime` fields |

### Python

| Format | `--format` name | Detection |
|---|---|---|
| **stdlib logging** | `python` | `levelname` + `msg` fields |
| **structlog** | `python` | `event` + level field |

### AWS

| Format | `--format` name | Detection |
|---|---|---|
| **Lambda** | `lambda` | `timestamp` + `message`, structured JSON |
| **CloudWatch** | `cloudwatch` | `logEvents` array or `logGroup` + `logStream` |

### Ruby

| Format | `--format` name | Detection |
|---|---|---|
| **Rails** | `rails` | `X, [TIMESTAMP #PID]  LEVEL -- : message` (production.log) |

### Infrastructure

| Format | `--format` name | Detection |
|---|---|---|
| **Syslog (RFC 5424)** | `syslog` | `<priority>version timestamp hostname...` |
| **Nginx JSON** | `nginxjson` or `nginx-json` | `remote_addr` + `request` + `status`; status code mapped to level |
| **CLF / Combined** | `clf` | Apache/Nginx Common Log Format and Combined Log Format |

### Catch-all

| Format | `--format` name | Detection |
|---|---|---|
| **Generic** | `generic` | Any JSON with a level-like field (`level`, `lvl`, `severity`...) and a message-like field (`msg`, `message`, `text`...) |
| **textline** | `textline` | Lines starting with an ISO timestamp or a level keyword |

Unrecognized lines are always passed through unchanged.

## Custom formats

If your app uses field names that don't match any built-in format, define a custom format in `~/.config/logr/config.json`.

Two modes are available: **JSON field mapping** and **regex pattern**. See [`examples/custom-formats/`](examples/custom-formats/) for ready-to-use examples.

### JSON field mapping

Map your app's field names to logr's timestamp / level / message:

```json
{
  "custom_formats": {
    "myapp": {
      "probe_field": "log_time",
      "ts_field":    "log_time",
      "level_field": "log_level",
      "msg_field":   "log_message"
    }
  }
}
```

```sh
# Explicit
cat app.log | logr --format myapp

# Auto-detected (because probe_field is set)
cat app.log | logr
```

**Field reference:**

| Key | Type | Description |
|---|---|---|
| `probe_field` | string | A field that must be present for auto-detection. Omit to disable auto-detection (explicit `--format` only). |
| `ts_field` | string | Field name for the timestamp |
| `level_field` | string | Field name for the log level |
| `msg_field` | string | Field name for the message |
| `level_map` | object | Map raw level values to canonical names. Works for both string and numeric values. |

**Using `level_map`** — useful when levels are integers or non-standard strings:

```json
{
  "custom_formats": {
    "myapp": {
      "probe_field": "log_time",
      "ts_field":    "log_time",
      "level_field": "log_level",
      "msg_field":   "log_message",
      "level_map": {
        "30":       "INFO",
        "40":       "WARN",
        "50":       "ERROR",
        "NOTICE":   "INFO",
        "CRITICAL": "FATAL"
      }
    }
  }
}
```

### Regex pattern

For non-JSON text logs, provide a regex with named capture groups. Groups named `ts`, `level`, and `msg` map to the standard fields. Any other named group is added to the entry's fields.

```json
{
  "custom_formats": {
    "rails": {
      "pattern": "^(?P<ts>[A-Z], \\[\\S+ #\\d+\\]\\s+)(?P<level>[A-Z]+) -- \\w*: (?P<msg>.+)$"
    }
  }
}
```

```sh
cat production.log | logr --format rails
```

Another example — timestamped text with a service tag:

```json
{
  "custom_formats": {
    "myservice": {
      "pattern": "^(?P<ts>\\S+) \\[(?P<service>[^\\]]+)\\] (?P<level>[A-Z]+) (?P<msg>.+)$"
    }
  }
}
```

Input: `2025-04-12T14:23:05Z [auth] WARN login attempt failed`
Output: `14:23:05.000  WARN  login attempt failed  service=auth`

### Multiple custom formats

You can define as many formats as you need. All of them with a `probe_field` or `pattern` participate in auto-detection:

```json
{
  "custom_formats": {
    "orders-api": {
      "probe_field": "order_ts",
      "ts_field":    "order_ts",
      "level_field": "severity",
      "msg_field":   "event"
    },
    "legacy-worker": {
      "pattern": "^\\[(?P<ts>[^\\]]+)\\] (?P<level>\\w+) (?P<msg>.+)$"
    }
  }
}
```

## Configuration

Optional config file at `~/.config/logr/config.json`:

```json
{
  "no_color":    false,
  "compact":     false,
  "hide_fields": ["pid", "hostname"],
  "custom_formats": {
    "myapp": {
      "probe_field": "log_time",
      "ts_field":    "log_time",
      "level_field": "log_level",
      "msg_field":   "log_message"
    }
  }
}
```

| Key | Type | Default | Description |
|---|---|---|---|
| `no_color` | bool | `false` | Disable color output globally |
| `compact` | bool | `false` | Use compact (one-line) output by default |
| `hide_fields` | string[] | `[]` | Fields to hide from output globally |
| `custom_formats` | object | `{}` | User-defined format definitions (see above) |

## Verify a release

Releases are signed with an EdDSA key. To verify:

```sh
# Download checksums and signature
curl -fsSL https://github.com/AgusRdz/logr/releases/download/v0.0.1/checksums.txt -o checksums.txt
curl -fsSL https://github.com/AgusRdz/logr/releases/download/v0.0.1/checksums.txt.sig -o checksums.txt.sig

# Verify signature using the public key in this repo
xxd -r -p checksums.txt.sig > checksums.txt.sig.bin
openssl pkeyutl -verify -inkey public_key.pem -pubin -rawin -in checksums.txt -sigfile checksums.txt.sig.bin
```

## License

MIT
