package formats

import "testing"

func TestMonologProbe(t *testing.T) {
	m := Monolog{}

	// valid monolog
	if !m.Probe([]byte(`{"message":"hello","context":{},"level":200,"level_name":"INFO","channel":"app","datetime":"2024-04-12T14:23:05+00:00","extra":{}}`)) {
		t.Error("expected true for valid monolog")
	}

	// missing level_name
	if m.Probe([]byte(`{"message":"hi","channel":"app","datetime":"2024-04-12T14:23:05+00:00"}`)) {
		t.Error("expected false for missing level_name")
	}

	// missing channel
	if m.Probe([]byte(`{"message":"hi","level_name":"INFO","datetime":"2024-04-12T14:23:05+00:00"}`)) {
		t.Error("expected false for missing channel")
	}

	// missing datetime
	if m.Probe([]byte(`{"message":"hi","level_name":"INFO","channel":"app"}`)) {
		t.Error("expected false for missing datetime")
	}

	// invalid JSON
	if m.Probe([]byte(`not json`)) {
		t.Error("expected false for invalid JSON")
	}
}

func TestMonologParse(t *testing.T) {
	m := Monolog{}

	e := m.Parse([]byte(`{"message":"Server started","context":{"port":8080},"level":200,"level_name":"INFO","channel":"app","datetime":"2024-04-12T14:23:05+00:00","extra":{}}`))
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

	// numeric level and message removed from fields
	for _, k := range []string{"level", "level_name", "message", "datetime", "context", "extra"} {
		if _, ok := e.Fields[k]; ok {
			t.Errorf("field %q should be removed or flattened", k)
		}
	}

	// context flattened
	if _, ok := e.Fields["context.port"]; !ok {
		t.Error("context.port should be in fields")
	}

	// WARNING → WARN
	e2 := m.Parse([]byte(`{"message":"warn","context":{},"level":300,"level_name":"WARNING","channel":"app","datetime":"2024-04-12T14:23:06+00:00","extra":{}}`))
	if e2.Level != "WARN" {
		t.Errorf("WARNING: got %q, want WARN", e2.Level)
	}

	// CRITICAL → FATAL
	e3 := m.Parse([]byte(`{"message":"crit","context":{},"level":500,"level_name":"CRITICAL","channel":"app","datetime":"2024-04-12T14:23:07+00:00","extra":{}}`))
	if e3.Level != "FATAL" {
		t.Errorf("CRITICAL: got %q, want FATAL", e3.Level)
	}
}
