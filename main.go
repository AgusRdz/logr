package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/AgusRdz/logr/config"
	"github.com/AgusRdz/logr/detect"
	"github.com/AgusRdz/logr/filter"
	"github.com/AgusRdz/logr/follow"
	"github.com/AgusRdz/logr/formats"
	"github.com/AgusRdz/logr/render"
	"github.com/AgusRdz/logr/stats"
	"github.com/AgusRdz/logr/updater"

	"bufio"
)

var version = "dev"

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
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "logr: %v\n", err)
		os.Exit(1)
	}
}

// multiFlag allows --field to be specified multiple times.
type multiFlag []string

func (m *multiFlag) String() string  { return strings.Join(*m, ", ") }
func (m *multiFlag) Set(v string) error { *m = append(*m, v); return nil }

func run() error {
	fs := flag.NewFlagSet("logr", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // we handle errors ourselves

	var (
		levelFlag    string
		sinceFlag    string
		untilFlag    string
		fieldFlags   multiFlag
		containsFlag string
		formatFlag   string
		noColorFlag  bool
		colorFlag    bool
		compactFlag  bool
		rawFlag      bool
		fieldsFlag   string
		hideFlag     string
		followFlag   bool
		statsFlag    bool
		fileFlag     string
	)

	fs.StringVar(&levelFlag, "level", "", "filter by level (e.g., error or warn,error)")
	fs.StringVar(&sinceFlag, "since", "", "show entries after this time (e.g., 10m, 1h, 2025-04-12T14:00)")
	fs.StringVar(&untilFlag, "until", "", "show entries before this time")
	fs.Var(&fieldFlags, "field", "filter by field=value (repeatable)")
	fs.StringVar(&containsFlag, "contains", "", "filter entries containing substring")
	fs.StringVar(&formatFlag, "format", "", "force format (pino|winston|zap|bunyan|serilog|log4j|monolog|python|syslog|logfmt|clf|generic|<custom>)")
	fs.BoolVar(&noColorFlag, "no-color", false, "disable color output")
	fs.BoolVar(&colorFlag, "color", false, "force color output even when piped")
	fs.BoolVar(&compactFlag, "compact", false, "one-line output per entry")
	fs.BoolVar(&rawFlag, "raw", false, "pass through original lines, only filter")
	fs.StringVar(&fieldsFlag, "fields", "", "show only these fields (comma-separated)")
	fs.StringVar(&hideFlag, "hide", "", "hide these fields (comma-separated)")
	fs.BoolVar(&followFlag, "follow", false, "follow file like tail -f")
	fs.BoolVar(&statsFlag, "stats", false, "print stats summary instead of streaming output")
	fs.StringVar(&fileFlag, "file", "", "input file path")

	// First non-flag arg (before any --flag) is the file path.
	args := os.Args[1:]
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		// Subcommands handled above; anything else is a file path.
		switch args[0] {
		case "version", "update", "help":
			// already handled
		default:
			fileFlag = args[0]
			args = args[1:]
		}
	}

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("%v\n\nRun 'logr help' for usage", err)
	}

	// Remaining positional arg after flags is also a valid file path.
	if fs.NArg() > 0 && fileFlag == "" {
		fileFlag = fs.Arg(0)
	}

	// --follow requires a file
	if followFlag && fileFlag == "" {
		return fmt.Errorf("--follow requires a file path")
	}

	// Load config (fills in defaults; flags always win)
	cfg := config.Load()

	// Resolve color: flag > no-color > tty detection > config
	colorEnabled := isTTY()
	if cfg.NoColor {
		colorEnabled = false
	}
	if noColorFlag {
		colorEnabled = false
	}
	if colorFlag {
		colorEnabled = true
	}

	// Apply config defaults for compact
	if cfg.Compact && !compactFlag {
		compactFlag = true
	}

	// Build filter chain
	var filters filter.Chain

	if levelFlag != "" {
		filters = append(filters, filter.NewLevelFilter(levelFlag))
	}

	var sinceTime, untilTime time.Time
	if sinceFlag != "" {
		t, err := filter.ParseSince(sinceFlag)
		if err != nil {
			return fmt.Errorf("--since: %w", err)
		}
		sinceTime = t
	}
	if untilFlag != "" {
		t, err := filter.ParseUntil(untilFlag)
		if err != nil {
			return fmt.Errorf("--until: %w", err)
		}
		untilTime = t
	}
	if !sinceTime.IsZero() || !untilTime.IsZero() {
		filters = append(filters, filter.NewTimeFilter(sinceTime, untilTime))
	}

	if containsFlag != "" {
		filters = append(filters, filter.NewContainsFilter(containsFlag))
	}

	for _, kv := range fieldFlags {
		ff, err := filter.NewFieldFilter(kv)
		if err != nil {
			return err
		}
		filters = append(filters, ff)
	}

	// Build render config
	var showFields, hideFields []string
	if fieldsFlag != "" {
		for _, f := range strings.Split(fieldsFlag, ",") {
			if s := strings.TrimSpace(f); s != "" {
				showFields = append(showFields, s)
			}
		}
	}
	// Merge hide flags: command-line + config
	hideAll := append(cfg.HideFields, []string{}...)
	if hideFlag != "" {
		for _, f := range strings.Split(hideFlag, ",") {
			if s := strings.TrimSpace(f); s != "" {
				hideAll = append(hideAll, s)
			}
		}
	}
	for _, f := range hideAll {
		if s := strings.TrimSpace(f); s != "" {
			hideFields = append(hideFields, s)
		}
	}

	renderCfg := render.Config{
		Color:  colorEnabled,
		Fields: showFields,
		Hide:   hideFields,
	}

	// --- Follow mode ---
	if followFlag {
		return runFollow(fileFlag, filters, renderCfg, rawFlag, compactFlag)
	}

	// --- Open input ---
	var r io.Reader
	if fileFlag != "" {
		f, err := os.Open(fileFlag)
		if err != nil {
			return err
		}
		defer f.Close()
		r = f
	} else {
		r = os.Stdin
	}

	// --- Read sample lines for format detection ---
	const sampleSize = 20
	var sampleLines [][]byte

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 10*1024*1024)
	for len(sampleLines) < sampleSize && sc.Scan() {
		line := sc.Bytes()
		cp := make([]byte, len(line))
		copy(cp, line)
		sampleLines = append(sampleLines, cp)
	}
	if err := sc.Err(); err != nil {
		return err
	}

	// Resolve custom format if --format names a user-defined format.
	// Done before declaring the fmt variable to keep fmt package accessible.
	var customFmt formats.Format
	if formatFlag != "" {
		if def, ok := cfg.CustomFormats[formatFlag]; ok {
			c, err := formats.NewCustom(def)
			if err != nil {
				return fmt.Errorf("custom format %q: %w", formatFlag, err)
			}
			customFmt = c
		}
	}

	// Build auto-detectable custom extras (those with a probe_field or pattern).
	var customExtras []formats.Format
	for _, def := range cfg.CustomFormats {
		if def.ProbeField != "" || def.Pattern != "" {
			if c, err := formats.NewCustom(def); err == nil {
				customExtras = append(customExtras, c)
			}
		}
	}

	// Detect or override format
	var selectedFmt formats.Format
	if customFmt != nil {
		selectedFmt = customFmt
	} else if formatFlag != "" {
		selectedFmt = detect.ByName(formatFlag)
	} else {
		selectedFmt = detect.DetectWithExtras(sampleLines, customExtras)
	}

	// Replay sample + rest of reader through the pipeline
	var sampleBuf bytes.Buffer
	for _, line := range sampleLines {
		sampleBuf.Write(line)
		sampleBuf.WriteByte('\n')
	}
	combined := io.MultiReader(&sampleBuf, r)

	// --- Stats mode ---
	if statsFlag {
		return runStats(combined, selectedFmt, filters, colorEnabled)
	}

	// --- Build renderer ---
	var renderer render.Renderer
	switch {
	case rawFlag:
		renderer = render.NewRaw(os.Stdout)
	case compactFlag:
		renderer = render.NewCompact(os.Stdout, renderCfg)
	default:
		renderer = render.NewPretty(os.Stdout, renderCfg)
	}

	// --- Stream pipeline ---
	return streamLines(combined, selectedFmt, filters, renderer)
}

func streamLines(r io.Reader, fmt formats.Format, filters filter.Chain, renderer render.Renderer) error {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 10*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		entry := fmt.Parse(line)
		if len(filters) > 0 && !filters.Match(entry) {
			continue
		}
		renderer.Write(entry)
	}
	return sc.Err()
}

func runStats(r io.Reader, fmt formats.Format, filters filter.Chain, colorEnabled bool) error {
	s := stats.New()
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 10*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		entry := fmt.Parse(line)
		if len(filters) > 0 && !filters.Match(entry) {
			continue
		}
		s.Add(entry)
	}
	if err := sc.Err(); err != nil {
		return err
	}
	s.Print(os.Stdout, colorEnabled)
	return nil
}

func runFollow(path string, filters filter.Chain, renderCfg render.Config, rawFlag, compactFlag bool) error {
	var renderer render.Renderer
	switch {
	case rawFlag:
		renderer = render.NewRaw(os.Stdout)
	case compactFlag:
		renderer = render.NewCompact(os.Stdout, renderCfg)
	default:
		renderer = render.NewPretty(os.Stdout, renderCfg)
	}

	out := make(chan []byte, 64)
	done := make(chan struct{})
	defer close(done)

	go follow.Follow(path, out, done)

	// Read enough lines to detect format, then replay + stream
	const sampleSize = 5
	var sample [][]byte
	var detected formats.Format

	for line := range out {
		if detected == nil {
			sample = append(sample, line)
			if len(sample) >= sampleSize {
				detected = detect.Detect(sample)
				for _, s := range sample {
					processLine(s, detected, filters, renderer)
				}
				sample = nil
			}
			continue
		}
		processLine(line, detected, filters, renderer)
	}

	// Flush remaining sample if file had fewer than sampleSize lines
	if detected == nil && len(sample) > 0 {
		detected = detect.Detect(sample)
		for _, s := range sample {
			processLine(s, detected, filters, renderer)
		}
	}

	return nil
}

func processLine(line []byte, fmt formats.Format, filters filter.Chain, renderer render.Renderer) {
	entry := fmt.Parse(line)
	if len(filters) > 0 && !filters.Match(entry) {
		return
	}
	renderer.Write(entry)
}

func printHelp() {
	fmt.Printf(`%s

%s

  cat app.log | logr
  logr app.log
  logr app.log --level error
  logr app.log --since 30m --level warn,error
  logr app.log --field requestId=abc123
  logr app.log --contains "payment failed"
  logr --follow app.log --level error
  logr app.log --stats --since 1h

%s
  --level      filter by level: debug, info, warn, error, fatal (comma-separated)
  --since      show entries after: 10m, 1h, 7d, or 2025-04-12T14:00
  --until      show entries before (same format as --since)
  --field      filter by field=value (repeatable: --field a=1 --field b=2)
  --contains   filter entries containing substring
  --format     force format: pino, winston, zap, bunyan, serilog, log4j, monolog, python, syslog, logfmt, clf, generic, or a custom name from config
  --follow     tail -f mode (file only)
  --stats      print stats summary instead of streaming
  --compact    one-line output per entry
  --raw        pass through original lines, only filter
  --fields     show only these fields (comma-separated)
  --hide       hide these fields (comma-separated)
  --no-color   disable color output
  --color      force color even when piped

%s
  logr version   print version
  logr update    update to latest release
  logr help      show this help

`, bold("logr"), cyan("usage:"), cyan("flags:"), cyan("commands:"))
}
