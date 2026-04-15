package formats

import "testing"

func TestSpringBootProbe(t *testing.T) {
	s := SpringBoot{}

	// valid Spring Boot JSON
	if !s.Probe([]byte(`{"@timestamp":"2025-04-12T14:23:05.123Z","@version":"1","message":"Started","level":"INFO"}`)) {
		t.Error("expected true for valid Spring Boot JSON")
	}

	// missing @timestamp
	if s.Probe([]byte(`{"level":"INFO","message":"Started"}`)) {
		t.Error("expected false for missing @timestamp")
	}

	// missing message
	if s.Probe([]byte(`{"@timestamp":"2025-04-12T14:23:05Z","level":"INFO"}`)) {
		t.Error("expected false for missing message")
	}

	// invalid JSON
	if s.Probe([]byte(`not json`)) {
		t.Error("expected false for invalid JSON")
	}
}

func TestSpringBootParse(t *testing.T) {
	s := SpringBoot{}

	e := s.Parse([]byte(`{"@timestamp":"2025-04-12T14:23:05.123Z","@version":"1","message":"Application started","logger_name":"com.example.App","thread_name":"main","level":"INFO","level_value":20000}`))
	if e.ParseErr {
		t.Fatal("unexpected ParseErr")
	}
	if e.Level != "INFO" {
		t.Errorf("level: got %q, want INFO", e.Level)
	}
	if e.Message != "Application started" {
		t.Errorf("msg = %q", e.Message)
	}
	if e.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}

	// noise fields removed
	for _, k := range []string{"@timestamp", "@version", "message", "level", "level_value"} {
		if _, ok := e.Fields[k]; ok {
			t.Errorf("field %q should be removed", k)
		}
	}

	// logger_name kept
	if _, ok := e.Fields["logger_name"]; !ok {
		t.Error("logger_name should be in fields")
	}

	// WARN
	e2 := s.Parse([]byte(`{"@timestamp":"2025-04-12T14:23:06Z","message":"pool exhausted","level":"WARN"}`))
	if e2.Level != "WARN" {
		t.Errorf("WARN: got %q", e2.Level)
	}
}
