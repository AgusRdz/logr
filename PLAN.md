# logr - PLAN.md

## Overview

`logr` is a JSON log beautifier and filter for the patterns software engineers encounter 80% of the time.
`jq` is powerful but its syntax is a barrier for quick log inspection. `logr` covers the common cases - filter by level, time range, field value, and follow a file - with zero syntax to learn.

**Domain:** `logr.run`
**Repository:** `AgusRdz/logr`
**Language:** Go 1.24 (`github.com/mattn/go-isatty` is the only external dep - no Cobra, no log libraries)
**Primary use:** Lambda, CloudWatch, Express/Pino, Winston, structured JSON logs from any service
**Build environment:** Docker only - never rely on host Go installation

---

## Goals

- Auto-detect log format and pretty-print with color and human timestamps
- Filter by level, time range, and arbitrary field values in one command
- Follow a file like `tail -f` with live filtering
- Pipe-friendly: reads from stdin or file
- Handles malformed lines gracefully (non-JSON lines are passed through, never dropped)

## Non-Goals

- Not a log aggregation or storage system
- No remote log fetching (use `awslogs`, `kubectl logs` to pipe into `logr`)
- No log writing or routing
- No regex capture groups or complex transforms

---

## CLI Interface

```
# From stdin (pipe)
cat app.log | logr
aws logs get-log-events ... | logr
kubectl logs my-pod | logr

# From file
logr app.log
logr --file app.log

# Filtering
logr app.log --level error
logr app.log --level warn,error
logr app.log --since 10m
logr app.log --since 1h
logr app.log --since "2025-04-12T14:00:00"
logr app.log --until "2025-04-12T15:00:00"
logr app.log --field requestId=abc123
logr app.log --field userId=42 --field level=error   # AND logic
logr app.log --contains "payment failed"             # substring in any field

# Format hints (auto-detected by default)
logr app.log --format pino
logr app.log --format winston
logr app.log --format lambda
logr app.log --format cloudwatch
logr app.log --format generic                        # any JSON with level+msg+ts

# Output control
logr app.log --no-color
logr app.log --compact              # one line per entry, no pretty-print
logr app.log --fields ts,level,msg,requestId   # show only these fields
logr app.log --hide ts,traceId      # hide specific fields
logr app.log --raw                  # pass through unchanged, just filter

# Follow mode
logr --follow app.log
logr --follow app.log --level error
logr --follow app.log --field service=payments

# Stats
logr app.log --stats
logr app.log --stats --since 1h

# Meta
logr version
logr update
```

### Output (pretty mode)

```
14:23:05.123  INFO   GET /users → 200 OK                    requestId=abc123  latency=42ms
14:23:05.891  WARN   DB pool exhausted, retrying            service=users     attempt=2
14:23:07.002  ERROR  Payment gateway timeout                 requestId=xyz789  orderId=99
  → cause: "connection timeout after 30s"
  → stack: PaymentService.charge (payment.js:142)
```

- Timestamps: human-readable, respects system timezone
- Level colors: `DEBUG` dim, `INFO` default, `WARN` yellow, `ERROR` red, `FATAL` red+bold
- Primary message field is prominent; secondary fields are dim on the same line
- Multi-line fields (stack traces, nested objects) are indented below the main line

---

## Architecture

### Package structure

```
logr/
├── main.go                # CLI dispatch (os.Args + flag package - no Cobra)
├── color.go               # TTY detection + ANSI helpers (see color.go pattern)
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
├── docker-compose.yml
├── cliff.toml
├── install.sh
├── install.ps1
├── public_key.pem         # EdDSA public key (committed - used to verify releases)
├── .github/
│   └── workflows/
│       ├── ci.yml
│       ├── release.yml
│       └── stale-check.yml
├── detect/
│   └── detect.go          # auto-detect log format from sample lines
├── formats/
│   ├── format.go          # interface: Format{ Parse(line) → Entry, bool }
│   ├── lambda.go          # AWS Lambda JSON logs
│   ├── cloudwatch.go      # CloudWatch structured logs
│   ├── pino.go            # Pino (Node.js)
│   ├── winston.go         # Winston (Node.js)
│   ├── generic.go         # any JSON with level + message + timestamp fields
│   └── passthrough.go     # non-JSON lines: pass through unchanged
├── filter/
│   ├── filter.go          # filter chain: level, time range, fields, contains
│   └── time.go            # --since/--until parsing (relative + absolute)
├── render/
│   ├── pretty.go          # colored multi-line output
│   ├── compact.go         # one-line output
│   └── raw.go             # pass through unchanged
├── follow/
│   └── follow.go          # tail -f implementation
├── stats/
│   └── stats.go           # count by level, time distribution
├── updater/
│   └── updater.go         # self-update from GitHub releases
├── config/
│   └── config.go          # user config (~/.config/logr/config.yml)
└── testdata/
    ├── lambda.log
    ├── pino.log
    ├── winston.log
    ├── generic.log
    ├── mixed.log
    └── malformed.log
```

### go.mod

```
module github.com/AgusRdz/logr

go 1.24.0

require github.com/mattn/go-isatty v0.0.20

require golang.org/x/sys v0.x.x // indirect
```

`go-isatty` is the only external dependency. Everything else is stdlib.

### color.go

Exact same pattern as chop - copy verbatim, package main, add `colorRed` and `colorRedBold` for ERROR/FATAL:

```go
package main

import (
    "os"
    "github.com/mattn/go-isatty"
)

func isTTY() bool {
    return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

const (
    colorReset    = "\033[0m"
    colorBold     = "\033[1m"
    colorDim      = "\033[2m"
    colorCyan     = "\033[36m"
    colorYellow   = "\033[33m"
    colorRed      = "\033[31m"
    colorRedBold  = "\033[1;31m"
)

func bold(s string) string   { if !isTTY() { return s }; return colorBold + s + colorReset }
func dim(s string) string    { if !isTTY() { return s }; return colorDim + s + colorReset }
func cyan(s string) string   { if !isTTY() { return s }; return colorCyan + s + colorReset }
func yellow(s string) string { if !isTTY() { return s }; return colorYellow + s + colorReset }
func red(s string) string    { if !isTTY() { return s }; return colorRed + s + colorReset }
func fatal(s string) string  { if !isTTY() { return s }; return colorRedBold + s + colorReset }
```

Color is suppressed automatically when piped (not a TTY). `--no-color` flag overrides and always suppresses.

### main.go dispatch pattern

No Cobra. Plain `os.Args` switch for subcommands, `flag` package for flags:

```go
var version = "dev" // injected at build time via -ldflags

func main() {
    if len(os.Args) > 1 {
        switch os.Args[1] {
        case "version", "--version", "-v":
            fmt.Printf("logr %s\n", version)
            return
        case "update":
            updater.Run(version)
            return
        case "help", "--help", "-h":
            printHelp()
            return
        }
    }
    runLogr()
}
```

`runLogr()` parses all flags with the `flag` package and runs the pipeline.

### Core Entry type

```go
// formats/format.go
// The canonical internal representation of a log line.
// All formats parse into this; the renderer reads from this.
type Entry struct {
    Timestamp time.Time
    Level     string          // normalized: DEBUG INFO WARN ERROR FATAL
    Message   string
    Fields    map[string]any  // all other fields
    Raw       []byte          // original line, preserved for --raw mode
    ParseErr  bool            // true if line was not valid JSON
}
```

### Processing pipeline

```
input (stdin or file)
  └── scanner.Scan() → raw lines
        └── detect.Detect(sample) → Format
              └── format.Parse(line) → Entry
                    └── filter.Apply(Entry, filters) → bool
                          └── render.Write(Entry)
```

Each stage is stateless and composable. Filters are ANDed by default.

### Format detection

```
detect.Detect(firstNLines []string) Format →
  try each format's Probe(line) bool:
    pino:        has "v" field (Pino version number) + "pid"
    winston:     has "level" + "message" + "timestamp" (Winston defaults)
    lambda:      has "timestamp" + "message" fields, no "v" field
    cloudwatch:  has "logEvents" wrapper or CloudWatch metadata fields
    generic:     has any of: level/lvl/severity + msg/message + ts/time/timestamp
    passthrough: none of the above
  → return first matching format
  → if ambiguous, prefer more specific (pino > generic)
```

---

## Format specifications

### Lambda

```json
{"timestamp":"2025-04-12T14:23:05.123Z","level":"INFO","message":"handler invoked","requestId":"abc123","functionName":"payments-prod"}
```
Key fields: `timestamp`, `level`, `message`, `requestId`

### Pino

```json
{"level":30,"time":1712930585123,"pid":42,"hostname":"app-1","msg":"server started","port":3000}
```
Level is a number: 10=trace, 20=debug, 30=info, 40=warn, 50=error, 60=fatal

### Winston

```json
{"level":"info","message":"server started","timestamp":"2025-04-12 14:23:05","service":"api"}
```

### Generic

Any JSON object with at least a level-like field and a message-like field:
- Level aliases: `level`, `lvl`, `severity`, `log_level`
- Message aliases: `msg`, `message`, `text`, `body`
- Timestamp aliases: `ts`, `time`, `timestamp`, `at`, `created_at`

---

## Filter logic

```go
type Filter interface {
    Match(Entry) bool
}

// LevelFilter: match if entry level is in the allowed set
// TimeFilter: match if entry timestamp is within [since, until]
// FieldFilter: match if Fields[key] == value (string coercion)
// ContainsFilter: match if any string field contains the substring

// Chain: all filters must match (AND)
type FilterChain []Filter
func (c FilterChain) Match(e Entry) bool {
    for _, f := range c {
        if !f.Match(e) { return false }
    }
    return true
}
```

Filter order (cheapest first):
1. `LevelFilter` - O(1) map lookup
2. `TimeFilter` - cheap timestamp compare
3. `ContainsFilter` - string scan
4. `FieldFilter` - map traversal (most expensive)

### Time parsing (`--since`, `--until`)

```
"10m"                → time.Now().Add(-10 * time.Minute)
"1h"                 → time.Now().Add(-1 * time.Hour)
"7d"                 → time.Now().Add(-7 * 24 * time.Hour)
"2025-04-12T14:00"   → parsed as RFC3339 / common datetime formats
"14:00:00"           → today at 14:00:00 in local timezone
```

---

## Follow mode

```go
// follow/follow.go
// tail -f: read to EOF, then poll for new content every 100ms.
//
// Rotation detection:
//   - Linux/macOS: compare syscall.Stat_t.Ino (inode number)
//   - Windows: use os.SameFile(fi1, fi2) - syscall.Stat_t is not available on Windows

func Follow(path string, out chan<- []byte) {
    f, _ := os.Open(path)
    f.Seek(0, io.SeekEnd)
    origInfo, _ := f.Stat()
    for {
        line, err := reader.ReadBytes('\n')
        if err == io.EOF {
            time.Sleep(100 * time.Millisecond)
            // Check for log rotation: if the file at path is now a different file, reopen
            if newInfo, err := os.Stat(path); err == nil && !os.SameFile(origInfo, newInfo) {
                f.Close()
                f, _ = os.Open(path)
                origInfo = newInfo
            }
            continue
        }
        out <- line
    }
}
```

`os.SameFile` is the portable cross-platform approach - it compares inodes on Unix and file IDs on Windows.

---

## Stats mode (`--stats`)

```
logr app.log --stats --since 1h

log stats - last 1h

  total:  1,243 entries
  errors: 12  (0.97%)   ████░░░░░░░░░░░░░░░░
  warns:  89  (7.2%)    ████████░░░░░░░░░░░░
  info:   1,142 (91.9%)

  by minute (last 10m):
  14:23 ██████████ 42
  14:24 ████████   33
  14:25 ████████████████ 67   ← spike
  14:26 ████████   35

  top error messages:
  1. "Payment gateway timeout"   (8 times)
  2. "DB pool exhausted"         (3 times)
  3. "Invalid JWT signature"     (1 time)
```

---

## Docker setup

### Dockerfile

```dockerfile
FROM golang:1.24-alpine

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download 2>/dev/null || true

COPY . .

CMD ["go", "build", "-o", "bin/logr", "."]
```

### docker-compose.yml

```yaml
services:
  dev:
    build: .
    volumes:
      - .:/app
      - go-cache:/go/pkg
      - go-build-cache:/root/.cache/go-build
    working_dir: /app

volumes:
  go-cache:
    name: go-pkg-cache
  go-build-cache:
    name: go-build-cache
```

**All `make` targets run inside Docker.** Never require Go on the host.

---

## Makefile

```makefile
.PHONY: build test coverage clean cross install changelog release release-patch release-minor release-major

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

build:
	docker compose run --rm dev go build -ldflags="$(LDFLAGS)" -o bin/logr .

test:
	docker compose run --rm dev go test ./... -v

coverage:
	docker compose run --rm dev go test -coverprofile=coverage.out ./...
	docker compose run --rm dev go tool cover -func=coverage.out

clean:
	rm -rf bin/

# Platform detection
UNAME_S := $(shell uname -s)
ifeq ($(findstring MINGW,$(UNAME_S)),MINGW)
  GOOS ?= windows
else ifeq ($(findstring MSYS,$(UNAME_S)),MSYS)
  GOOS ?= windows
else ifeq ($(findstring Darwin,$(UNAME_S)),Darwin)
  GOOS ?= darwin
else
  GOOS ?= linux
endif
GOARCH ?= $(if $(filter arm64 aarch64,$(shell uname -m)),arm64,amd64)
EXT := $(if $(filter windows,$(GOOS)),.exe,)
BINARY := bin/logr$(EXT)

ifeq ($(GOOS),windows)
  INSTALL_DIR ?= $(LOCALAPPDATA)/Programs/logr
else
  INSTALL_DIR ?= $(HOME)/.local/bin
endif

install:
	docker compose run --rm dev sh -c "CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags='$(LDFLAGS)' -o $(BINARY) ."
	@mkdir -p "$(INSTALL_DIR)"
	cp $(BINARY) "$(INSTALL_DIR)/logr$(EXT)"
	@echo "installed logr $(VERSION) ($(GOOS)/$(GOARCH)) to $(INSTALL_DIR)/logr$(EXT)"

cross:
	docker compose run --rm dev sh -c "\
		CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags='$(LDFLAGS)' -o bin/logr-linux-amd64 . && \
		CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -ldflags='$(LDFLAGS)' -o bin/logr-linux-arm64 . && \
		CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags='$(LDFLAGS)' -o bin/logr-darwin-amd64 . && \
		CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags='$(LDFLAGS)' -o bin/logr-darwin-arm64 . && \
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags='$(LDFLAGS)' -o bin/logr-windows-amd64.exe ."

# Changelog - requires git-cliff on host (only needed locally; CI uses the action)
.PHONY: _require-git-cliff
_require-git-cliff:
	@command -v git-cliff >/dev/null 2>&1 || { echo "git-cliff required. See https://git-cliff.org/docs/installation"; exit 1; }

changelog: _require-git-cliff
	git-cliff --output CHANGELOG.md

CURRENT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo v0.0.0)
MAJOR := $(shell echo $(CURRENT_TAG) | sed 's/^v//' | cut -d. -f1)
MINOR := $(shell echo $(CURRENT_TAG) | sed 's/^v//' | cut -d. -f2)
PATCH := $(shell echo $(CURRENT_TAG) | sed 's/^v//' | cut -d. -f3)

release:
	@BUMP=patch; \
	if git log $$(git describe --tags --abbrev=0)..HEAD --format="%s" | grep -qE '^feat(\(.*\))?!:'; then BUMP=major; \
	elif git log $$(git describe --tags --abbrev=0)..HEAD --format="%B" | grep -q 'BREAKING CHANGE'; then BUMP=major; \
	elif git log $$(git describe --tags --abbrev=0)..HEAD --format="%s" | grep -qE '^feat'; then BUMP=minor; fi; \
	echo "detected: $$BUMP"; \
	$(MAKE) release-$$BUMP

release-patch: _require-git-cliff
	@NEXT=v$(MAJOR).$(MINOR).$(shell echo $$(($(PATCH)+1))); \
	echo "$(CURRENT_TAG) -> $$NEXT"; \
	git-cliff --tag $$NEXT --output CHANGELOG.md && \
	git add CHANGELOG.md && \
	git commit -m "chore: update changelog for $$NEXT" && \
	git tag $$NEXT && \
	{ git push origin HEAD $$NEXT && echo "released $$NEXT"; } || { git tag -d $$NEXT; git reset --soft HEAD~1; echo "push failed - rolled back"; exit 1; }

release-minor: _require-git-cliff
	@NEXT=v$(MAJOR).$(shell echo $$(($(MINOR)+1))).0; \
	echo "$(CURRENT_TAG) -> $$NEXT"; \
	git-cliff --tag $$NEXT --output CHANGELOG.md && \
	git add CHANGELOG.md && \
	git commit -m "chore: update changelog for $$NEXT" && \
	git tag $$NEXT && \
	{ git push origin HEAD $$NEXT && echo "released $$NEXT"; } || { git tag -d $$NEXT; git reset --soft HEAD~1; echo "push failed - rolled back"; exit 1; }

release-major: _require-git-cliff
	@NEXT=v$(shell echo $$(($(MAJOR)+1))).0.0; \
	echo "$(CURRENT_TAG) -> $$NEXT"; \
	git-cliff --tag $$NEXT --output CHANGELOG.md && \
	git add CHANGELOG.md && \
	git commit -m "chore: update changelog for $$NEXT" && \
	git tag $$NEXT && \
	{ git push origin HEAD $$NEXT && echo "released $$NEXT"; } || { git tag -d $$NEXT; git reset --soft HEAD~1; echo "push failed - rolled back"; exit 1; }
```

---

## cliff.toml - Changelog automation

`cliff.toml` configures [git-cliff](https://git-cliff.org), a Rust tool that reads conventional commits and generates `CHANGELOG.md` automatically.

**How it works:**
- Locally: `make release-patch/minor/major` calls `git-cliff` to regenerate `CHANGELOG.md`, commits it, tags, and pushes
- In CI: the `release.yml` workflow uses the `orhun/git-cliff-action` GitHub Action - no host tooling needed
- The GitHub Release body is auto-populated with only the current version's changes (`--latest --strip header`)

```toml
[changelog]
header = """# Changelog

All notable changes to logr are documented here.\n
"""
body = """
{%- macro remote_url() -%}https://github.com/AgusRdz/logr{%- endmacro -%}

{% if version -%}
## [{{ version | trim_start_matches(pat="v") }}] - {{ timestamp | date(format="%Y-%m-%d") }}
{% else -%}
## [Unreleased]
{% endif -%}

{% for group, commits in commits | group_by(attribute="group") %}
### {{ group | upper_first }}
{% for commit in commits -%}
- {{ commit.message | split(pat="\n") | first | upper_first | trim }}
  ([{{ commit.id | truncate(length=7, end="") }}]({{ self::remote_url() }}/commit/{{ commit.id }}))
{% endfor -%}
{% endfor %}
"""
footer = ""
trim = true

[git]
conventional_commits = true
filter_unconventional = false
split_commits = false
commit_parsers = [
    { message = "^fix: address PR", skip = true },
    { message = "^fix: resolve merge conflict", skip = true },
    { message = "^chore: update changelog", skip = true },
    { message = "^feat", group = "Features" },
    { message = "^fix", group = "Bug Fixes" },
    { message = "^docs", group = "Documentation" },
    { message = "^refactor", group = "Refactoring" },
    { message = "^test", group = "Testing" },
    { message = "^chore", group = "Miscellaneous" },
    { message = "^ci", group = "CI/CD" },
    { message = "^perf", group = "Performance" },
    { message = ".*", group = "Other" },
]
protect_breaking_commits = false
filter_commits = false
tag_pattern = "v[0-9].*"
sort_commits = "oldest"
```

---

## GitHub Actions

### .github/workflows/ci.yml

Runs on every push to `main` and on all PRs. Builds inside the Go action (CI does not use Docker - the action provides Go directly):

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

env:
  FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: "true"

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      # Pin to SHA - tag-based pins are mutable and can be hijacked.
      - uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683  # v4
      - uses: actions/setup-go@d35c59abb061a4a6fb18e82ac0862c26744d6ab5  # v5
        with:
          go-version: "1.24"
      - run: go test ./... -v -count=1
```

### .github/workflows/release.yml

Triggered on `v*` tags. Builds all platforms, signs checksums, creates GitHub Release with changelog, attests provenance, and optionally updates a Homebrew tap:

```yaml
name: Release
run-name: Release ${{ github.ref_name }}

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write
  id-token: write
  attestations: write

env:
  FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: "true"

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      # Pin actions to commit SHAs to prevent supply chain attacks.
      # To update: replace SHA with latest from the action's releases page.
      - uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683  # v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@d35c59abb061a4a6fb18e82ac0862c26744d6ab5  # v5
        with:
          go-version: "1.24"

      - name: Run tests
        run: go test ./... -count=1

      - name: Validate semver tag
        run: |
          if ! echo "${{ github.ref_name }}" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
            echo "ERROR: tag '${{ github.ref_name }}' is not valid semver"
            exit 1
          fi

      - name: Update CHANGELOG.md
        uses: orhun/git-cliff-action@c93ef52f3d0ddcdcc9bd5447d98d458a11cd4f72
        with:
          config: cliff.toml
          args: --output CHANGELOG.md
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Generate release notes
        uses: orhun/git-cliff-action@c93ef52f3d0ddcdcc9bd5447d98d458a11cd4f72
        id: changelog
        with:
          config: cliff.toml
          args: --latest --strip header
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Build binaries
        run: |
          VERSION=${{ github.ref_name }}
          LDFLAGS="-s -w -X main.version=${VERSION}"
          CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o bin/logr-linux-amd64 .
          CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o bin/logr-linux-arm64 .
          CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o bin/logr-darwin-amd64 .
          CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o bin/logr-darwin-arm64 .
          CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o bin/logr-windows-amd64.exe .

      - name: Generate checksums
        run: cd bin && sha256sum logr-* > checksums.txt

      - name: Sign checksums
        env:
          SIGNING_KEY: ${{ secrets.SIGNING_KEY }}
        run: |
          set -euo pipefail
          echo "$SIGNING_KEY" | base64 -d > /tmp/signing.pem
          chmod 600 /tmp/signing.pem
          openssl pkey -in /tmp/signing.pem -noout
          openssl pkeyutl -sign -inkey /tmp/signing.pem -rawin -in bin/checksums.txt \
            | xxd -p -c 256 | tr -d '\n' > bin/checksums.txt.sig
          rm /tmp/signing.pem
          [ -s bin/checksums.txt.sig ] || { echo "ERROR: checksums.txt.sig is empty"; exit 1; }

      - name: Create GitHub Release
        uses: softprops/action-gh-release@da05d552573ad5aba039eaac05058a918a7bf631  # v2
        with:
          body: ${{ steps.changelog.outputs.content }}
          files: |
            bin/logr-linux-amd64
            bin/logr-linux-arm64
            bin/logr-darwin-amd64
            bin/logr-darwin-arm64
            bin/logr-windows-amd64.exe
            bin/checksums.txt
            bin/checksums.txt.sig

      - name: Attest build provenance
        uses: actions/attest-build-provenance@e8d6784b2e26e3c8b11a3982b43f3d5dad85a8e9  # v2
        with:
          subject-path: bin/logr-*

      # Homebrew tap update - uncomment when ready to publish.
      # Requires: HOMEBREW_TAP_TOKEN secret + AgusRdz/homebrew-tap repo.
      # - name: Update Homebrew formula
      #   env:
      #     HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
      #   run: |
          [ -z "${HOMEBREW_TAP_TOKEN}" ] && echo "HOMEBREW_TAP_TOKEN not set, skipping" && exit 0
          VERSION=${{ github.ref_name }}
          VER=${VERSION#v}
          SHA_DARWIN_ARM64=$(sha256sum bin/logr-darwin-arm64 | awk '{print $1}')
          SHA_DARWIN_AMD64=$(sha256sum bin/logr-darwin-amd64 | awk '{print $1}')
          SHA_LINUX_AMD64=$(sha256sum bin/logr-linux-amd64 | awk '{print $1}')
          SHA_LINUX_ARM64=$(sha256sum bin/logr-linux-arm64 | awk '{print $1}')
          git clone https://x-access-token:${HOMEBREW_TAP_TOKEN}@github.com/AgusRdz/homebrew-tap.git tap
          cat > tap/Formula/logr.rb << FORMULA
          class Logr < Formula
            desc "JSON log beautifier and filter"
            homepage "https://logr.run"
            version "${VER}"
            on_macos do
              if Hardware::CPU.arm?
                url "https://github.com/AgusRdz/logr/releases/download/${VERSION}/logr-darwin-arm64"
                sha256 "${SHA_DARWIN_ARM64}"
              else
                url "https://github.com/AgusRdz/logr/releases/download/${VERSION}/logr-darwin-amd64"
                sha256 "${SHA_DARWIN_AMD64}"
              end
            end
            on_linux do
              if Hardware::CPU.arm?
                url "https://github.com/AgusRdz/logr/releases/download/${VERSION}/logr-linux-arm64"
                sha256 "${SHA_LINUX_ARM64}"
              else
                url "https://github.com/AgusRdz/logr/releases/download/${VERSION}/logr-linux-amd64"
                sha256 "${SHA_LINUX_AMD64}"
              end
            end
            def install
              binary = Dir["logr-*"].first
              chmod 0755, binary
              bin.install binary => "logr"
            end
            test do
              assert_match version.to_s, shell_output("#{bin}/logr version")
            end
          end
          FORMULA
          cd tap && \
          git config user.name "github-actions[bot]" && \
          git config user.email "github-actions[bot]@users.noreply.github.com" && \
          git add Formula/logr.rb && \
          git commit -m "logr ${VERSION}" && \
          git push
```

---

## Signing setup

EdDSA key pair. Private key never committed - stored as a GitHub Actions secret.

```bash
# Generate once, locally
openssl genpkey -algorithm ed25519 -out signing.pem
openssl pkey -in signing.pem -pubout -out public_key.pem

# Encode private key for GitHub secret
# Note: base64 -w0 is Linux-only. This command works on both Linux and macOS:
base64 < signing.pem | tr -d '\n'   # paste output as SIGNING_KEY secret in GitHub repo settings
```

- `public_key.pem` - committed to the repo (users can verify downloads)
- `signing.pem` - add to `.gitignore`, never committed
- `SIGNING_KEY` - base64-encoded private key stored in GitHub Actions secrets

Set before first push to GitHub.

---

## Installation scripts

### install.sh

```sh
#!/bin/sh
set -e

REPO="AgusRdz/logr"

OS="$(uname -s)"
case "$OS" in
  Linux*)  OS="linux" ;;
  Darwin*) OS="darwin" ;;
  MINGW*|MSYS*|CYGWIN*) OS="windows" ;;
  *) echo "unsupported OS: $OS" >&2; exit 1 ;;
esac

if [ -z "$LOGR_INSTALL_DIR" ]; then
  if [ "$OS" = "windows" ]; then
    INSTALL_DIR="$(cygpath "$LOCALAPPDATA/Programs/logr" 2>/dev/null || echo "$HOME/AppData/Local/Programs/logr")"
  else
    INSTALL_DIR="$HOME/.local/bin"
  fi
else
  INSTALL_DIR="$LOGR_INSTALL_DIR"
fi

ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

EXT=""
[ "$OS" = "windows" ] && EXT=".exe"

BINARY="logr-${OS}-${ARCH}${EXT}"

if [ -z "$LOGR_VERSION" ]; then
  LOGR_VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": *"//;s/".*//')
fi

[ -z "$LOGR_VERSION" ] && echo "failed to determine latest version" >&2 && exit 1

URL="https://github.com/${REPO}/releases/download/${LOGR_VERSION}/${BINARY}"

echo "installing logr ${LOGR_VERSION} (${OS}/${ARCH})..."

mkdir -p "$INSTALL_DIR"

# Download binary + checksums, then verify integrity before installing
TMPDIR="$(mktemp -d)"
curl -fsSL "$URL" -o "${TMPDIR}/logr${EXT}"
curl -fsSL "https://github.com/${REPO}/releases/download/${LOGR_VERSION}/checksums.txt" -o "${TMPDIR}/checksums.txt"

# Verify SHA256 (sha256sum on Linux, shasum -a 256 on macOS)
EXPECTED=$(grep "${BINARY}" "${TMPDIR}/checksums.txt" | awk '{print $1}')
if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL=$(sha256sum "${TMPDIR}/logr${EXT}" | awk '{print $1}')
else
  ACTUAL=$(shasum -a 256 "${TMPDIR}/logr${EXT}" | awk '{print $1}')
fi

if [ "$EXPECTED" != "$ACTUAL" ]; then
  echo "checksum mismatch - aborting" >&2
  rm -rf "$TMPDIR"
  exit 1
fi

cp "${TMPDIR}/logr${EXT}" "${INSTALL_DIR}/logr${EXT}"
chmod +x "${INSTALL_DIR}/logr${EXT}"
rm -rf "$TMPDIR"

echo "installed logr to ${INSTALL_DIR}/logr${EXT}"

case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    if [ "$OS" = "windows" ]; then
      WIN_DIR=$(cygpath -w "$INSTALL_DIR" 2>/dev/null || echo "$INSTALL_DIR")
      powershell.exe -NoProfile -Command "\$p = [Environment]::GetEnvironmentVariable('Path', 'User'); \$d = '${WIN_DIR}'.TrimEnd('\\'); if ((\$p -split ';' | ForEach-Object { \$_.TrimEnd('\\') }) -notcontains \$d) { [Environment]::SetEnvironmentVariable('Path', \"\$d;\$p\", 'User') }"
      export PATH="${INSTALL_DIR}:$PATH"
    else
      SHELL_NAME="$(basename "${SHELL:-}")"
      case "$SHELL_NAME" in
        zsh)  SHELL_RC="$HOME/.zshrc" ;;
        bash) SHELL_RC="$HOME/.bashrc" ;;
        *)    SHELL_RC="" ;;
      esac
      PATH_LINE="export PATH=\"${INSTALL_DIR}:\$PATH\""
      if [ -n "$SHELL_RC" ] && ! grep -qF "$INSTALL_DIR" "$SHELL_RC" 2>/dev/null; then
        printf '\n# logr\n%s\n' "$PATH_LINE" >> "$SHELL_RC"
        echo "Added ${INSTALL_DIR} to PATH in $SHELL_RC"
        echo "Reload your shell: source $SHELL_RC"
      fi
    fi
    ;;
esac

echo ""
echo "Quick start:"
echo "  cat app.log | logr"
echo "  logr app.log --level error --since 30m"
echo "  logr --follow app.log --level error"
```

### install.ps1

```powershell
$ErrorActionPreference = "Stop"

$Repo = "AgusRdz/logr"
$InstallDir = if ($env:LOGR_INSTALL_DIR) { $env:LOGR_INSTALL_DIR } else { "$env:LOCALAPPDATA\Programs\logr" }

$Arch = if ([System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture -eq [System.Runtime.InteropServices.Architecture]::Arm64) { "arm64" } else { "amd64" }
$Binary = "logr-windows-$Arch.exe"

if (-not $env:LOGR_VERSION) {
    $Release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
    $env:LOGR_VERSION = $Release.tag_name
}

$Url = "https://github.com/$Repo/releases/download/$($env:LOGR_VERSION)/$Binary"
$ChecksumsUrl = "https://github.com/$Repo/releases/download/$($env:LOGR_VERSION)/checksums.txt"
Write-Host "installing logr $($env:LOGR_VERSION) (windows/$Arch)..."

# Download to temp location and verify checksum before installing
$TmpDir = [System.IO.Path]::GetTempPath() + [System.Guid]::NewGuid().ToString()
New-Item -ItemType Directory -Force -Path $TmpDir | Out-Null
$TmpBinary = Join-Path $TmpDir $Binary
$TmpChecksums = Join-Path $TmpDir "checksums.txt"

Invoke-WebRequest -Uri $Url -OutFile $TmpBinary
Invoke-WebRequest -Uri $ChecksumsUrl -OutFile $TmpChecksums

$Expected = (Get-Content $TmpChecksums | Where-Object { $_ -match $Binary }) -split '\s+' | Select-Object -First 1
$Actual = (Get-FileHash -Algorithm SHA256 $TmpBinary).Hash.ToLower()

if ($Expected -ne $Actual) {
    Remove-Item -Recurse -Force $TmpDir
    Write-Error "checksum mismatch - aborting"
    exit 1
}

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
$Destination = Join-Path $InstallDir "logr.exe"
Move-Item -Force $TmpBinary $Destination
Remove-Item -Recurse -Force $TmpDir
Write-Host "installed logr to $Destination"

$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
$CleanDir = $InstallDir.TrimEnd("\")
if (($UserPath -split ";" | ForEach-Object { $_.TrimEnd("\") }) -notcontains $CleanDir) {
    [Environment]::SetEnvironmentVariable("PATH", "$InstallDir;$UserPath", "User")
    Write-Host "added $InstallDir to PATH"
}
$env:PATH = "$InstallDir;$env:PATH"

# Broadcast PATH change to open terminals (no restart needed)
$HWND_BROADCAST = [IntPtr]0xffff
$WM_SETTINGCHANGE = 0x001a
$MethodDefinition = @'
[DllImport("user32.dll", SetLastError = true, CharSet = CharSet.Auto)]
public static extern IntPtr SendMessageTimeout(IntPtr hWnd, uint Msg, IntPtr wParam, string lParam, uint fuFlags, uint uTimeout, out IntPtr lpdwResult);
'@
$User32 = Add-Type -MemberDefinition $MethodDefinition -Name "User32" -Namespace "Win32" -PassThru
$result = [IntPtr]::Zero
$User32::SendMessageTimeout($HWND_BROADCAST, $WM_SETTINGCHANGE, [IntPtr]::Zero, "Environment", 2, 100, [ref]$result) | Out-Null

Write-Host ""
Write-Host "Quick start:"
Write-Host "  cat app.log | logr"
Write-Host "  logr app.log --level error --since 30m"
Write-Host "  logr --follow app.log --level error"
```

---

## Security

- Reads files and stdin - no network access (except `logr update`)
- Handles malformed UTF-8 gracefully (replace invalid bytes, never panic)
- No eval, no plugin loading
- `--field` values are compared as strings - no injection possible
- Release binaries are signed with EdDSA; checksums are SHA-256

---

## Performance

- Streaming: lines are processed one at a time - never loads the whole file into memory
- Use `bufio.Scanner` for line-by-line reading. **Always call `scanner.Buffer()` to raise the max token size** - the default is 64KB and a single JSON log line with an embedded stack trace or large payload will silently fail:
  ```go
  scanner := bufio.NewScanner(r)
  scanner.Buffer(make([]byte, 64*1024), 10*1024*1024) // allow up to 10MB per line
  ```
- Filter short-circuit: level filter runs first (cheapest), field filter last
- `--follow` uses 100ms poll interval - low CPU usage
- Color output uses direct ANSI escape codes - no color library
- Target: process 1M lines in < 3s on a modern machine

---

## .gitignore

```
bin/
*.exe
signing.pem
coverage.out
.chop.yml
.claude/
```

---

## Testing strategy

```
detect/
  detect_test.go        - sample lines from each format, verify detection
formats/
  lambda_test.go        - parse valid lines, verify Entry fields
  pino_test.go          - level number mapping (30→INFO, 50→ERROR)
  winston_test.go
  generic_test.go
  passthrough_test.go   - non-JSON lines returned with ParseErr=true
filter/
  filter_test.go        - each filter type: match, no-match, edge cases
  time_test.go          - all --since duration formats + absolute timestamps
follow/
  follow_test.go        - write lines to temp file, verify follow sees them
stats/
  stats_test.go         - count by level, minute bucketing
main_test.go            - golden output: input fixture → expected colored output
```

Test fixtures live in `testdata/`:
`lambda.log`, `pino.log`, `winston.log`, `generic.log`, `mixed.log`, `malformed.log`

Run via Docker: `make test`

---

## Implementation notes

- All log processing is streaming - never load the full file into memory
- Non-JSON lines are NEVER dropped. They pass through as `Entry{ParseErr: true, Raw: line}`
- Level normalization happens in each format's `Parse()` function. All internal levels are uppercase: `DEBUG INFO WARN ERROR FATAL`
- ANSI color codes are only emitted when `isTTY()` returns true, or when `--color` is explicitly set. `--no-color` always wins
- Follow mode detects log rotation with `os.SameFile()` - portable across Linux, macOS, and Windows. Do NOT use `syscall.Stat_t.Ino` directly (Unix-only)
- Filters are ANDed and short-circuit: `LevelFilter` first, `TimeFilter` second, `FieldFilter` last
- Always set `scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)` - default 64KB limit silently truncates large log lines
