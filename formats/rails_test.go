package formats

import "testing"

func TestRailsProbe(t *testing.T) {
	r := Rails{}

	if !r.Probe([]byte(`I, [2025-04-12T14:23:05.123456 #1234]  INFO -- : Started GET "/users" for 127.0.0.1`)) {
		t.Error("expected true for INFO line")
	}
	if !r.Probe([]byte(`E, [2025-04-12T14:23:06.567890 #1234] ERROR -- : ArgumentError: bad input`)) {
		t.Error("expected true for ERROR line")
	}

	// plain text — should not match
	if r.Probe([]byte(`INFO server started`)) {
		t.Error("expected false for plain text")
	}

	// JSON — should not match
	if r.Probe([]byte(`{"level":"info","msg":"hi"}`)) {
		t.Error("expected false for JSON")
	}

	// short line
	if r.Probe([]byte(`I, []`)) {
		t.Error("expected false for malformed short line")
	}
}

func TestRailsParse(t *testing.T) {
	r := Rails{}

	e := r.Parse([]byte(`I, [2025-04-12T14:23:05.123456 #1234]  INFO -- : Started GET "/users" for 127.0.0.1`))
	if e.ParseErr {
		t.Fatal("unexpected ParseErr")
	}
	if e.Level != "INFO" {
		t.Errorf("level: got %q, want INFO", e.Level)
	}
	if e.Message != `Started GET "/users" for 127.0.0.1` {
		t.Errorf("msg = %q", e.Message)
	}
	if e.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}

	// WARN
	e2 := r.Parse([]byte(`W, [2025-04-12T14:23:05.234567 #1234]  WARN -- : DB pool exhausted`))
	if e2.Level != "WARN" {
		t.Errorf("WARN: got %q", e2.Level)
	}

	// ERROR
	e3 := r.Parse([]byte(`E, [2025-04-12T14:23:06.567890 #1234] ERROR -- : ArgumentError: bad input`))
	if e3.Level != "ERROR" {
		t.Errorf("ERROR: got %q", e3.Level)
	}
	if e3.Message != "ArgumentError: bad input" {
		t.Errorf("msg = %q", e3.Message)
	}

	// FATAL
	e4 := r.Parse([]byte(`F, [2025-04-12T14:23:07.678901 #1234] FATAL -- : Killed`))
	if e4.Level != "FATAL" {
		t.Errorf("FATAL: got %q", e4.Level)
	}
}
