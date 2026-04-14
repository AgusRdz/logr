package render

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/AgusRdz/logr/formats"
)

func TestCompactWrite(t *testing.T) {
	var buf bytes.Buffer
	c := NewCompact(&buf, Config{Color: false})
	e := formats.Entry{
		Timestamp: time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC),
		Level:     "WARN",
		Message:   "disk space low",
		Fields:    map[string]any{"host": "web-01"},
		Raw:       []byte(`{"level":"warn","msg":"disk space low"}`),
	}
	c.Write(e)

	out := buf.String()
	// Should be a single line
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Errorf("expected single line, got %d lines: %s", len(lines), out)
	}
	if !strings.Contains(out, "2024-01-15T10:30:45Z") {
		t.Errorf("missing RFC3339 timestamp, got: %s", out)
	}
	if !strings.Contains(out, "WARN") {
		t.Errorf("missing level, got: %s", out)
	}
	if !strings.Contains(out, "disk space low") {
		t.Errorf("missing message, got: %s", out)
	}
	if !strings.Contains(out, "host=web-01") {
		t.Errorf("missing field, got: %s", out)
	}
}

func TestCompactWrite_ParseErr(t *testing.T) {
	var buf bytes.Buffer
	c := NewCompact(&buf, Config{})
	e := formats.Entry{
		ParseErr: true,
		Raw:      []byte("raw unparseable line"),
	}
	c.Write(e)

	out := buf.String()
	if !strings.Contains(out, "raw unparseable line") {
		t.Errorf("expected raw line, got: %s", out)
	}
}
