package formats

import "testing"

func TestSerilogProbe(t *testing.T) {
	s := Serilog{}

	// valid CLEF
	if !s.Probe([]byte(`{"@t":"2024-04-12T14:23:05.0000000Z","@mt":"Server started on {Port}","Port":8080}`)) {
		t.Error("expected true for valid CLEF")
	}

	// missing @mt
	if s.Probe([]byte(`{"@t":"2024-04-12T14:23:05Z","@l":"Info","msg":"hi"}`)) {
		t.Error("expected false for missing @mt")
	}

	// missing @t
	if s.Probe([]byte(`{"@mt":"Hello {Name}","Name":"world"}`)) {
		t.Error("expected false for missing @t")
	}

	// invalid JSON
	if s.Probe([]byte(`not json`)) {
		t.Error("expected false for invalid JSON")
	}
}

func TestSerilogParse(t *testing.T) {
	s := Serilog{}

	// omitted @l → INFO
	e := s.Parse([]byte(`{"@t":"2024-04-12T14:23:05.0000000Z","@mt":"Server started on {Port}","Port":8080}`))
	if e.ParseErr {
		t.Fatal("unexpected ParseErr")
	}
	if e.Level != "INFO" {
		t.Errorf("level: got %q, want INFO", e.Level)
	}
	if e.Message != "Server started on {Port}" {
		t.Errorf("msg = %q", e.Message)
	}
	if e.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}

	// @l Warning → WARN
	e2 := s.Parse([]byte(`{"@t":"2024-04-12T14:23:06Z","@mt":"Something wrong","@l":"Warning"}`))
	if e2.Level != "WARN" {
		t.Errorf("Warning: got %q, want WARN", e2.Level)
	}

	// @l Fatal → FATAL
	e3 := s.Parse([]byte(`{"@t":"2024-04-12T14:23:07Z","@mt":"Crash","@l":"Fatal"}`))
	if e3.Level != "FATAL" {
		t.Errorf("Fatal: got %q, want FATAL", e3.Level)
	}

	// @mt removed from fields
	if _, ok := e.Fields["@mt"]; ok {
		t.Error("@mt should be removed from fields")
	}
	// @t removed from fields
	if _, ok := e.Fields["@t"]; ok {
		t.Error("@t should be removed from fields")
	}
	// template vars kept in fields
	if _, ok := e.Fields["Port"]; !ok {
		t.Error("Port template var should be in fields")
	}
}
