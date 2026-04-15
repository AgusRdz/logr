package formats

import "testing"

func TestSyslogProbe(t *testing.T) {
	s := Syslog{}

	// valid RFC 5424
	if !s.Probe([]byte(`<34>1 2024-04-12T14:23:05.003Z mymachine.example.com su - ID47 - BOM message`)) {
		t.Error("expected true for valid RFC 5424")
	}

	// plain text — should not match
	if s.Probe([]byte(`INFO server started`)) {
		t.Error("expected false for plain text")
	}

	// JSON — should not match
	if s.Probe([]byte(`{"level":"info","msg":"hi"}`)) {
		t.Error("expected false for JSON")
	}

	// missing prival
	if s.Probe([]byte(`1 2024-04-12T14:23:05Z host app - - - message`)) {
		t.Error("expected false for missing <prival>")
	}
}

func TestSyslogParse(t *testing.T) {
	s := Syslog{}

	// severity 2 (CRIT) → FATAL
	e := s.Parse([]byte(`<34>1 2024-04-12T14:23:05.003Z mymachine.example.com su - ID47 - BOM 'su root' failed`))
	if e.ParseErr {
		t.Fatal("unexpected ParseErr")
	}
	// prival 34 = facility 4 (auth), severity 2 (CRIT) → FATAL
	if e.Level != "FATAL" {
		t.Errorf("prival 34 (severity 2): got %q, want FATAL", e.Level)
	}
	if e.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}

	// severity 6 (INFO) — prival 165 = facility 20, severity 5 (NOTICE) → INFO
	e2 := s.Parse([]byte(`<165>1 2024-04-12T14:23:06.000Z mymachine.example.com myapp 1234 - - Server started`))
	if e2.Level != "INFO" {
		t.Errorf("prival 165 (severity 5): got %q, want INFO", e2.Level)
	}
	if e2.Message != "Server started" {
		t.Errorf("msg = %q, want 'Server started'", e2.Message)
	}

	// severity 3 (ERR) — prival 131 = facility 16, severity 3 → ERROR
	e3 := s.Parse([]byte(`<131>1 2024-04-12T14:23:07.000Z mymachine.example.com myapp 1234 REQ001 - Request failed`))
	if e3.Level != "ERROR" {
		t.Errorf("prival 131 (severity 3): got %q, want ERROR", e3.Level)
	}

	// severity 7 (DEBUG) — prival 135 = facility 16, severity 7 → DEBUG
	e4 := s.Parse([]byte(`<135>1 2024-04-12T14:23:09.000Z mymachine.example.com myapp 1234 - - Debug checkpoint`))
	if e4.Level != "DEBUG" {
		t.Errorf("prival 135 (severity 7): got %q, want DEBUG", e4.Level)
	}
}
