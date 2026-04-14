package render

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/AgusRdz/logr/formats"
)

func makeEntry() formats.Entry {
	return formats.Entry{
		Timestamp: time.Date(2024, 1, 15, 10, 30, 45, 123000000, time.UTC),
		Level:     "INFO",
		Message:   "hello world",
		Fields: map[string]any{
			"user":    "alice",
			"req_id":  "abc123",
			"latency": 42,
		},
		Raw: []byte(`{"level":"info","msg":"hello world"}`),
	}
}

func TestPrettyWrite_Normal(t *testing.T) {
	var buf bytes.Buffer
	p := NewPretty(&buf, Config{Color: false})
	e := makeEntry()
	p.Write(e)

	out := buf.String()
	if !strings.Contains(out, "10:30:45.123") {
		t.Errorf("missing timestamp, got: %s", out)
	}
	if !strings.Contains(out, "INFO") {
		t.Errorf("missing level, got: %s", out)
	}
	if !strings.Contains(out, "hello world") {
		t.Errorf("missing message, got: %s", out)
	}
	if !strings.Contains(out, "user=alice") {
		t.Errorf("missing field user, got: %s", out)
	}
	if !strings.Contains(out, "req_id=abc123") {
		t.Errorf("missing field req_id, got: %s", out)
	}
}

func TestPrettyWrite_ParseErr(t *testing.T) {
	var buf bytes.Buffer
	p := NewPretty(&buf, Config{})
	e := formats.Entry{
		ParseErr: true,
		Raw:      []byte("unparseable garbage line"),
	}
	p.Write(e)

	out := buf.String()
	if !strings.Contains(out, "unparseable garbage line") {
		t.Errorf("expected raw line, got: %s", out)
	}
}

func TestPrettyWrite_NoColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewPretty(&buf, Config{Color: false})
	e := makeEntry()
	e.Level = "ERROR"
	p.Write(e)

	out := buf.String()
	if strings.Contains(out, "\033[") {
		t.Errorf("unexpected ANSI codes in no-color output: %s", out)
	}
}

func TestPrettyWrite_HideFields(t *testing.T) {
	var buf bytes.Buffer
	p := NewPretty(&buf, Config{Hide: []string{"user"}})
	e := makeEntry()
	p.Write(e)

	out := buf.String()
	if strings.Contains(out, "user=") {
		t.Errorf("hidden field 'user' appeared in output: %s", out)
	}
	if !strings.Contains(out, "req_id=abc123") {
		t.Errorf("visible field missing from output: %s", out)
	}
}

func TestPrettyWrite_ShowOnlyFields(t *testing.T) {
	var buf bytes.Buffer
	p := NewPretty(&buf, Config{Fields: []string{"req_id"}})
	e := makeEntry()
	p.Write(e)

	out := buf.String()
	if !strings.Contains(out, "req_id=abc123") {
		t.Errorf("expected req_id in output: %s", out)
	}
	if strings.Contains(out, "user=") {
		t.Errorf("unexpected field 'user' in output: %s", out)
	}
	if strings.Contains(out, "latency=") {
		t.Errorf("unexpected field 'latency' in output: %s", out)
	}
}
