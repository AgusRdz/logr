package render

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/AgusRdz/logr/formats"
)

const (
	colorDim    = "\033[2m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorRedBold = "\033[1;31m"
	colorReset  = "\033[0m"
)

// Pretty renders one log entry per line with optional ANSI color.
type Pretty struct {
	w   io.Writer
	cfg Config
}

func NewPretty(w io.Writer, cfg Config) *Pretty {
	return &Pretty{w: w, cfg: cfg}
}

func (p *Pretty) Write(e formats.Entry) {
	if e.ParseErr {
		fmt.Fprintf(p.w, "%s\n", e.Raw)
		return
	}

	// Timestamp
	ts := ""
	if !e.Timestamp.IsZero() {
		ts = e.Timestamp.Local().Format("15:04:05.000")
	}

	// Level (padded to 5, colored)
	level := levelLabel(e.Level, 5, p.cfg.Color)

	// Message
	msg := e.Message

	// Build first line
	var sb strings.Builder
	if ts != "" {
		sb.WriteString(ts)
		sb.WriteString("  ")
	}
	sb.WriteString(level)
	sb.WriteString("  ")
	sb.WriteString(msg)

	// Collect visible fields
	fields := visibleFields(e.Fields, p.cfg.Fields, p.cfg.Hide)

	var inline []string
	var nested []string

	for _, k := range fields {
		v := e.Fields[k]
		switch val := v.(type) {
		case map[string]any:
			b, _ := json.Marshal(val)
			nested = append(nested, fmt.Sprintf("  \u2192 %s: %s", k, string(b)))
		case []any:
			b, _ := json.Marshal(val)
			nested = append(nested, fmt.Sprintf("  \u2192 %s: %s", k, string(b)))
		case string:
			if strings.ContainsAny(val, " \t\n") {
				inline = append(inline, fmt.Sprintf("%s=%q", k, val))
			} else {
				inline = append(inline, fmt.Sprintf("%s=%s", k, val))
			}
		default:
			inline = append(inline, fmt.Sprintf("%s=%v", k, v))
		}
	}

	if len(inline) > 0 {
		inlineStr := strings.Join(inline, "  ")
		if p.cfg.Color {
			sb.WriteString("    ")
			sb.WriteString(colorDim)
			sb.WriteString(inlineStr)
			sb.WriteString(colorReset)
		} else {
			sb.WriteString("    ")
			sb.WriteString(inlineStr)
		}
	}

	fmt.Fprintln(p.w, sb.String())

	for _, line := range nested {
		if p.cfg.Color {
			fmt.Fprintf(p.w, "%s%s%s\n", colorDim, line, colorReset)
		} else {
			fmt.Fprintln(p.w, line)
		}
	}
}

// levelLabel returns the level string padded to width, with optional color.
func levelLabel(level string, width int, color bool) string {
	padded := fmt.Sprintf("%*s", width, level)
	if !color {
		return padded
	}
	switch level {
	case "DEBUG":
		return colorDim + padded + colorReset
	case "WARN":
		return colorYellow + padded + colorReset
	case "ERROR":
		return colorRed + padded + colorReset
	case "FATAL":
		return colorRedBold + padded + colorReset
	default:
		return padded
	}
}

// visibleFields returns sorted field keys after applying show/hide filters.
func visibleFields(fields map[string]any, show []string, hide []string) []string {
	hideSet := make(map[string]bool, len(hide))
	for _, h := range hide {
		hideSet[h] = true
	}

	var keys []string
	if len(show) > 0 {
		showSet := make(map[string]bool, len(show))
		for _, s := range show {
			showSet[s] = true
		}
		for k := range fields {
			if showSet[k] && !hideSet[k] {
				keys = append(keys, k)
			}
		}
	} else {
		for k := range fields {
			if !hideSet[k] {
				keys = append(keys, k)
			}
		}
	}
	sort.Strings(keys)
	return keys
}
