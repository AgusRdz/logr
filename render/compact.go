package render

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/AgusRdz/logr/formats"
)

// Compact renders one log entry per line in a concise format.
type Compact struct {
	w   io.Writer
	cfg Config
}

func NewCompact(w io.Writer, cfg Config) *Compact {
	return &Compact{w: w, cfg: cfg}
}

func (c *Compact) Write(e formats.Entry) {
	if e.ParseErr {
		fmt.Fprintf(c.w, "%s\n", e.Raw)
		return
	}

	level := levelLabel(e.Level, 5, c.cfg.Color)

	ts := ""
	if !e.Timestamp.IsZero() {
		ts = e.Timestamp.Format(time.RFC3339)
	}

	var sb strings.Builder
	sb.WriteString(level)
	if ts != "" {
		sb.WriteString("  ")
		sb.WriteString(ts)
	}
	sb.WriteString("  ")
	sb.WriteString(e.Message)

	fields := visibleFields(e.Fields, c.cfg.Fields, c.cfg.Hide)
	for _, k := range fields {
		v := e.Fields[k]
		var valStr string
		switch val := v.(type) {
		case map[string]any:
			b, _ := json.Marshal(val)
			valStr = string(b)
		case []any:
			b, _ := json.Marshal(val)
			valStr = string(b)
		case string:
			if strings.ContainsAny(val, " \t\n") {
				valStr = fmt.Sprintf("%q", val)
			} else {
				valStr = val
			}
		default:
			valStr = fmt.Sprintf("%v", v)
		}
		sb.WriteString("  ")
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(valStr)
	}

	fmt.Fprintln(c.w, sb.String())
}
