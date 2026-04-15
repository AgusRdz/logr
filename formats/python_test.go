package formats

import "testing"

func TestPythonProbe(t *testing.T) {
	p := Python{}

	// stdlib logging JSON
	if !p.Probe([]byte(`{"name":"app","levelname":"INFO","msg":"hello","created":1712930585.123}`)) {
		t.Error("expected true for stdlib logging")
	}

	// structlog JSON
	if !p.Probe([]byte(`{"event":"cache miss","level":"debug","timestamp":"2024-04-12T14:23:09Z"}`)) {
		t.Error("expected true for structlog")
	}

	// structlog with severity alias
	if !p.Probe([]byte(`{"event":"started","severity":"INFO","timestamp":"2024-04-12T14:23:05Z"}`)) {
		t.Error("expected true for structlog with severity")
	}

	// event without level — should not match
	if p.Probe([]byte(`{"event":"something happened","data":"value"}`)) {
		t.Error("expected false for event without level")
	}

	// generic JSON without levelname or event
	if p.Probe([]byte(`{"level":"info","msg":"hi"}`)) {
		t.Error("expected false for generic JSON (no levelname/event)")
	}

	// invalid JSON
	if p.Probe([]byte(`not json`)) {
		t.Error("expected false for invalid JSON")
	}
}

func TestPythonParse(t *testing.T) {
	p := Python{}

	// stdlib logging
	e := p.Parse([]byte(`{"name":"app","levelname":"INFO","msg":"Server started","created":1712930585.123,"pathname":"/app/main.py","lineno":42,"process":1234,"thread":140234567890,"funcName":"start"}`))
	if e.ParseErr {
		t.Fatal("unexpected ParseErr")
	}
	if e.Level != "INFO" {
		t.Errorf("level: got %q, want INFO", e.Level)
	}
	if e.Message != "Server started" {
		t.Errorf("msg = %q", e.Message)
	}
	if e.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}

	// noise fields removed
	for _, k := range []string{"name", "pathname", "lineno", "process", "thread", "funcName", "levelname"} {
		if _, ok := e.Fields[k]; ok {
			t.Errorf("field %q should be removed", k)
		}
	}

	// structlog
	e2 := p.Parse([]byte(`{"event":"cache miss","level":"debug","timestamp":"2024-04-12T14:23:09Z","key":"user:42"}`))
	if e2.Level != "DEBUG" {
		t.Errorf("structlog debug: got %q, want DEBUG", e2.Level)
	}
	if e2.Message != "cache miss" {
		t.Errorf("structlog event: got %q, want 'cache miss'", e2.Message)
	}
	if _, ok := e2.Fields["key"]; !ok {
		t.Error("key should be in fields")
	}

	// CRITICAL → FATAL
	e3 := p.Parse([]byte(`{"levelname":"CRITICAL","msg":"crash","created":1712930588.0}`))
	if e3.Level != "FATAL" {
		t.Errorf("CRITICAL: got %q, want FATAL", e3.Level)
	}
}
