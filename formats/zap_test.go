package formats

import "testing"

func TestZapProbe(t *testing.T) {
	z := Zap{}

	// valid zap
	if !z.Probe([]byte(`{"level":"info","ts":1712930585.123456,"caller":"server/main.go:42","msg":"started"}`)) {
		t.Error("expected true for valid zap")
	}

	// ts as integer string (not a float) — should not match
	if z.Probe([]byte(`{"level":"info","ts":"1712930585","caller":"main.go:1","msg":"hi"}`)) {
		t.Error("expected false for string ts")
	}

	// missing caller
	if z.Probe([]byte(`{"level":"info","ts":1712930585.1,"msg":"hi"}`)) {
		t.Error("expected false for missing caller")
	}

	// caller without colon
	if z.Probe([]byte(`{"level":"info","ts":1712930585.1,"caller":"mainfile","msg":"hi"}`)) {
		t.Error("expected false for caller without colon")
	}

	// missing msg
	if z.Probe([]byte(`{"level":"info","ts":1712930585.1,"caller":"main.go:1"}`)) {
		t.Error("expected false for missing msg")
	}

	// invalid JSON
	if z.Probe([]byte(`not json`)) {
		t.Error("expected false for invalid JSON")
	}
}

func TestZapParse(t *testing.T) {
	z := Zap{}

	e := z.Parse([]byte(`{"level":"info","ts":1712930585.123456,"caller":"server/main.go:42","msg":"server started","port":8080}`))
	if e.ParseErr {
		t.Fatal("unexpected ParseErr")
	}
	if e.Level != "INFO" {
		t.Errorf("level: got %q, want INFO", e.Level)
	}
	if e.Message != "server started" {
		t.Errorf("msg = %q, want 'server started'", e.Message)
	}
	if e.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}

	// caller kept in fields
	if _, ok := e.Fields["caller"]; !ok {
		t.Error("caller should remain in fields")
	}

	// port in fields
	if _, ok := e.Fields["port"]; !ok {
		t.Error("port should be in fields")
	}

	// warn
	e2 := z.Parse([]byte(`{"level":"warn","ts":1712930586.0,"caller":"db/pool.go:99","msg":"pool exhausted"}`))
	if e2.Level != "WARN" {
		t.Errorf("warn: got %q, want WARN", e2.Level)
	}

	// dpanic → treated as FATAL by NormalizeLevel fallback (uppercased)
	e3 := z.Parse([]byte(`{"level":"dpanic","ts":1712930587.0,"caller":"main.go:1","msg":"panic"}`))
	if e3.Level != "DPANIC" {
		t.Errorf("dpanic: got %q, want DPANIC", e3.Level)
	}
}
