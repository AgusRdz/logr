# Custom Format Examples

This directory shows how to define **custom formats** for apps with non-standard log schemas.

For well-known frameworks (Spring Boot, Rails, Nginx JSON, Pino, Zap, Serilog, etc.) use the built-in `--format` names — no config needed. Custom formats are for cases where the built-ins don't fit.

Copy entries from [`config.json`](config.json) into `~/.config/logr/config.json` to activate them.

---

## myapp — custom JSON field names

Your service logs JSON but uses field names like `log_time`, `log_level`, `log_message`.

**Config:**
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

**Sample log** (`logs/myapp.log`):
```
{"log_time":"2025-04-12T14:23:05.123Z","log_level":"INFO","log_message":"server started","port":8080}
{"log_time":"2025-04-12T14:23:07.002Z","log_level":"ERROR","log_message":"payment gateway timeout","orderId":99}
```

**Usage:**
```sh
logr --format myapp logs/myapp.log

# Auto-detected because probe_field is set
cat logs/myapp.log | logr
```

---

## myapp-numeric-levels — integer level codes

Same as above but level is a numeric code (like Pino's 30/40/50).

**Config:**
```json
{
  "custom_formats": {
    "myapp-numeric-levels": {
      "probe_field": "log_ts",
      "ts_field":    "log_ts",
      "level_field": "log_level",
      "msg_field":   "log_msg",
      "level_map": {
        "30": "INFO",
        "40": "WARN",
        "50": "ERROR",
        "60": "FATAL"
      }
    }
  }
}
```

**Usage:**
```sh
logr --format myapp-numeric-levels app.log
```

---

## legacy-worker — bracketed timestamp text logs

Old-style worker logs with `[timestamp] LEVEL: message` format.

**Config:**
```json
{
  "custom_formats": {
    "legacy-worker": {
      "pattern": "^\\[(?P<ts>[^\\]]+)\\] (?P<level>[A-Z]+): (?P<msg>.+)$"
    }
  }
}
```

**Sample log** (`logs/legacy-worker.log`):
```
[2025-04-12T14:23:05.123Z] INFO: Worker started, queue=email-notifications
[2025-04-12T14:23:06.890Z] WARN: Retry attempt 2 of 3, job_id=8821
[2025-04-12T14:23:07.234Z] ERROR: Job failed permanently, job_id=8821
```

**Usage:**
```sh
logr --format legacy-worker logs/legacy-worker.log
logr --format legacy-worker logs/legacy-worker.log --level warn,error
```

---

## tagged-service — text logs with an inline service tag

Logs that embed a service name in brackets: `2025-04-12T14:23:05Z [auth] WARN message`.

**Config:**
```json
{
  "custom_formats": {
    "tagged-service": {
      "pattern": "^(?P<ts>\\S+) \\[(?P<service>[^\\]]+)\\] (?P<level>[A-Z]+) (?P<msg>.+)$"
    }
  }
}
```

The `service` named group becomes a field in the output:
```
14:23:05.000  WARN  login attempt failed  service=auth
```

---

## Regex named group reference

| Group | Maps to |
|---|---|
| `ts` | Timestamp — supports RFC3339, Unix epoch, and other common formats |
| `level` | Level — normalized to `DEBUG INFO WARN ERROR FATAL` |
| `msg` | Message body |
| any other name | Added to the entry's fields, visible in output |
