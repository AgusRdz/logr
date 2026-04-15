package formats

import "testing"

func TestBunyanProbe(t *testing.T) {
	b := Bunyan{}

	// valid bunyan: v + name + pid + time-as-string
	if !b.Probe([]byte(`{"v":0,"level":30,"name":"myapp","hostname":"h","pid":1,"time":"2024-04-12T14:23:05.000Z","msg":"hi"}`)) {
		t.Error("expected true for valid bunyan")
	}

	// pino-style time as number — should not match
	if b.Probe([]byte(`{"v":1,"pid":42,"level":30,"msg":"hi","time":1700000000000}`)) {
		t.Error("expected false for numeric time (pino)")
	}

	// missing name
	if b.Probe([]byte(`{"v":0,"pid":1,"time":"2024-04-12T14:23:05.000Z","msg":"hi"}`)) {
		t.Error("expected false for missing name")
	}

	// missing pid
	if b.Probe([]byte(`{"v":0,"name":"app","time":"2024-04-12T14:23:05.000Z","msg":"hi"}`)) {
		t.Error("expected false for missing pid")
	}

	// invalid JSON
	if b.Probe([]byte(`not json`)) {
		t.Error("expected false for invalid JSON")
	}
}

func TestBunyanParse(t *testing.T) {
	b := Bunyan{}

	e := b.Parse([]byte(`{"v":0,"level":30,"name":"myapp","hostname":"myhost","pid":1234,"time":"2024-04-12T14:23:05.000Z","msg":"server started","port":3000}`))
	if e.ParseErr {
		t.Fatal("unexpected ParseErr")
	}
	if e.Level != "INFO" {
		t.Errorf("level 30: got %q, want INFO", e.Level)
	}
	if e.Message != "server started" {
		t.Errorf("msg = %q, want 'server started'", e.Message)
	}
	if e.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}

	// noise fields removed
	for _, k := range []string{"v", "pid", "hostname", "name"} {
		if _, ok := e.Fields[k]; ok {
			t.Errorf("field %q should be removed", k)
		}
	}

	// port kept in fields
	if _, ok := e.Fields["port"]; !ok {
		t.Error("field 'port' should be in fields")
	}

	// level 40 → WARN
	e2 := b.Parse([]byte(`{"v":0,"name":"app","pid":1,"level":40,"msg":"warn","time":"2024-04-12T14:23:05.000Z"}`))
	if e2.Level != "WARN" {
		t.Errorf("level 40: got %q, want WARN", e2.Level)
	}

	// level 50 → ERROR
	e3 := b.Parse([]byte(`{"v":0,"name":"app","pid":1,"level":50,"msg":"err","time":"2024-04-12T14:23:05.000Z"}`))
	if e3.Level != "ERROR" {
		t.Errorf("level 50: got %q, want ERROR", e3.Level)
	}

	// level 60 → FATAL
	e4 := b.Parse([]byte(`{"v":0,"name":"app","pid":1,"level":60,"msg":"fatal","time":"2024-04-12T14:23:05.000Z"}`))
	if e4.Level != "FATAL" {
		t.Errorf("level 60: got %q, want FATAL", e4.Level)
	}
}
