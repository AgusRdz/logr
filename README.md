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
- Switching between projects that use different loggers (Pino, Winston, Lambda, CloudWatch)
- `kubectl logs my-pod | logr` or `aws logs get-log-events ... | logr` - ugly JSON becomes readable instantly
- `logr --follow app.log --level error` during an incident - no jq gymnastics

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

# Filter by time
logr app.log --since 30m
logr app.log --since 1h
logr app.log --since 2025-04-12T14:00

# Filter by field value
logr app.log --field requestId=abc123
logr app.log --field service=payments --field level=error

# Substring search
logr app.log --contains "payment failed"

# Follow mode (like tail -f)
logr --follow app.log
logr --follow app.log --level error

# Stats summary
logr app.log --stats
logr app.log --stats --since 1h
```

## Flags

| Flag | Description |
|---|---|
| `--level` | Filter by level: `debug`, `info`, `warn`, `error`, `fatal` (comma-separated) |
| `--since` | Show entries after: `10m`, `1h`, `7d`, or `2025-04-12T14:00` |
| `--until` | Show entries before (same format as `--since`) |
| `--field` | Filter by `key=value` (repeatable) |
| `--contains` | Filter entries containing a substring |
| `--format` | Force format: `pino`, `winston`, `lambda`, `cloudwatch`, `generic` |
| `--follow` | Follow file like `tail -f` |
| `--stats` | Print stats summary instead of streaming |
| `--compact` | One-line output per entry |
| `--raw` | Pass through original lines, only filter |
| `--fields` | Show only these fields (comma-separated) |
| `--hide` | Hide specific fields (comma-separated) |
| `--no-color` | Disable color output |
| `--color` | Force color even when piped |

## Supported formats

logr auto-detects your log format. You can also force one with `--format`.

| Format | Detection |
|---|---|
| **Pino** | `v` + `pid` fields, integer levels |
| **Winston** | `level` + `message` + `timestamp` fields |
| **Lambda** | `timestamp` + `message`, no `v` field |
| **CloudWatch** | `logEvents` array or `logGroup` + `logStream` |
| **Generic** | Any JSON with a level-like and message-like field |

Non-JSON lines are always passed through unchanged.

## Configuration

Optional config file at `~/.config/logr/config.json`:

```json
{
  "no_color": false,
  "compact": false,
  "hide_fields": ["pid", "hostname"]
}
```

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
