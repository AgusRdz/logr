package formats

import "testing"

func TestPinoProbe(t *testing.T) {
	p := Pino{}

	// has v+pid
	if !p.Probe([]byte(`{"v":1,"pid":123,"level":30,"msg":"hi","time":1700000000000}`)) {
		t.Error("expected true for v+pid")
	}

	// missing v
	if p.Probe([]byte(`{"pid":123,"level":30,"msg":"hi"}`)) {
		t.Error("expected false for missing v")
	}

	// missing pid
	if p.Probe([]byte(`{"v":1,"level":30,"msg":"hi"}`)) {
		t.Error("expected false for missing pid")
	}

	// invalid JSON
	if p.Probe([]byte(`not json`)) {
		t.Error("expected false for invalid JSON")
	}
}

func TestPinoParse(t *testing.T) {
	p := Pino{}

	// level 30 → INFO
	e := p.Parse([]byte(`{"v":1,"pid":42,"hostname":"box","level":30,"msg":"hello","time":1700000000000}`))
	if e.ParseErr {
		t.Fatal("unexpected ParseErr")
	}
	if e.Level != "INFO" {
		t.Errorf("level 30: got %q, want INFO", e.Level)
	}
	if e.Message != "hello" {
		t.Errorf("msg = %q, want hello", e.Message)
	}
	if e.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}
	if e.Timestamp.Unix() != 1700000000 {
		t.Errorf("timestamp unix = %d, want 1700000000", e.Timestamp.Unix())
	}
	// noise fields removed
	for _, k := range []string{"v", "pid", "hostname"} {
		if _, ok := e.Fields[k]; ok {
			t.Errorf("field %q should be removed", k)
		}
	}

	// level 40 → WARN
	e2 := p.Parse([]byte(`{"v":1,"pid":1,"level":40,"msg":"warn","time":1700000000000}`))
	if e2.Level != "WARN" {
		t.Errorf("level 40: got %q, want WARN", e2.Level)
	}

	// level 50 → ERROR
	e3 := p.Parse([]byte(`{"v":1,"pid":1,"level":50,"msg":"err","time":1700000000000}`))
	if e3.Level != "ERROR" {
		t.Errorf("level 50: got %q, want ERROR", e3.Level)
	}

	// level 60 → FATAL
	e4 := p.Parse([]byte(`{"v":1,"pid":1,"level":60,"msg":"fatal","time":1700000000000}`))
	if e4.Level != "FATAL" {
		t.Errorf("level 60: got %q, want FATAL", e4.Level)
	}

	// level 20 → DEBUG
	e5 := p.Parse([]byte(`{"v":1,"pid":1,"level":20,"msg":"debug","time":1700000000000}`))
	if e5.Level != "DEBUG" {
		t.Errorf("level 20: got %q, want DEBUG", e5.Level)
	}
}
