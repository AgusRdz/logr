package formats

import "testing"

func TestCustomProbe_JSONMode(t *testing.T) {
	c, err := NewCustom(CustomFormatDef{
		ProbeField: "log_time",
		TSField:    "log_time",
		LevelField: "log_level",
		MsgField:   "log_message",
	})
	if err != nil {
		t.Fatal(err)
	}

	// has probe_field
	if !c.Probe([]byte(`{"log_time":"2024-04-12T14:23:05Z","log_level":"INFO","log_message":"hello"}`)) {
		t.Error("expected true when probe_field present")
	}

	// missing probe_field
	if c.Probe([]byte(`{"level":"info","msg":"hello"}`)) {
		t.Error("expected false when probe_field absent")
	}

	// invalid JSON
	if c.Probe([]byte(`not json`)) {
		t.Error("expected false for invalid JSON")
	}
}

func TestCustomParse_JSONMode(t *testing.T) {
	c, err := NewCustom(CustomFormatDef{
		ProbeField: "log_time",
		TSField:    "log_time",
		LevelField: "log_level",
		MsgField:   "log_message",
	})
	if err != nil {
		t.Fatal(err)
	}

	e := c.Parse([]byte(`{"log_time":"2024-04-12T14:23:05Z","log_level":"WARNING","log_message":"something happened","extra":"value"}`))
	if e.ParseErr {
		t.Fatal("unexpected ParseErr")
	}
	if e.Level != "WARN" {
		t.Errorf("level: got %q, want WARN", e.Level)
	}
	if e.Message != "something happened" {
		t.Errorf("msg = %q", e.Message)
	}
	if e.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}
	if _, ok := e.Fields["extra"]; !ok {
		t.Error("extra should remain in fields")
	}
	for _, k := range []string{"log_time", "log_level", "log_message"} {
		if _, ok := e.Fields[k]; ok {
			t.Errorf("field %q should be removed", k)
		}
	}
}

func TestCustomParse_JSONMode_LevelMap(t *testing.T) {
	c, err := NewCustom(CustomFormatDef{
		TSField:    "ts",
		LevelField: "lvl",
		MsgField:   "msg",
		LevelMap:   map[string]string{"30": "INFO", "40": "WARN", "50": "ERROR"},
	})
	if err != nil {
		t.Fatal(err)
	}

	e := c.Parse([]byte(`{"ts":1712930585.0,"lvl":40,"msg":"pool warn"}`))
	if e.Level != "WARN" {
		t.Errorf("numeric level 40 via level_map: got %q, want WARN", e.Level)
	}

	e2 := c.Parse([]byte(`{"ts":1712930585.0,"lvl":50,"msg":"error"}`))
	if e2.Level != "ERROR" {
		t.Errorf("numeric level 50: got %q, want ERROR", e2.Level)
	}
}

func TestCustomProbe_RegexMode(t *testing.T) {
	c, err := NewCustom(CustomFormatDef{
		Pattern: `^(?P<ts>\S+) (?P<level>[A-Z]+) (?P<msg>.+)$`,
	})
	if err != nil {
		t.Fatal(err)
	}

	if !c.Probe([]byte(`2024-04-12T14:23:05Z INFO server started`)) {
		t.Error("expected true for matching line")
	}
	if c.Probe([]byte(`{"level":"info","msg":"json"}`)) {
		t.Error("expected false for non-matching line")
	}
}

func TestCustomParse_RegexMode(t *testing.T) {
	c, err := NewCustom(CustomFormatDef{
		Pattern: `^(?P<ts>\S+) (?P<level>[A-Z]+) (?P<msg>.+)$`,
	})
	if err != nil {
		t.Fatal(err)
	}

	e := c.Parse([]byte(`2024-04-12T14:23:05Z WARN DB pool exhausted`))
	if e.ParseErr {
		t.Fatal("unexpected ParseErr")
	}
	if e.Level != "WARN" {
		t.Errorf("level: got %q, want WARN", e.Level)
	}
	if e.Message != "DB pool exhausted" {
		t.Errorf("msg = %q", e.Message)
	}
	if e.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}
}

func TestCustomParse_RegexMode_ExtraGroups(t *testing.T) {
	c, err := NewCustom(CustomFormatDef{
		Pattern: `^(?P<ts>\S+) \[(?P<service>\w+)\] (?P<level>[A-Z]+) (?P<msg>.+)$`,
	})
	if err != nil {
		t.Fatal(err)
	}

	e := c.Parse([]byte(`2024-04-12T14:23:05Z [auth] ERROR login failed`))
	if e.Level != "ERROR" {
		t.Errorf("level: got %q, want ERROR", e.Level)
	}
	if svc, ok := e.Fields["service"]; !ok || svc != "auth" {
		t.Errorf("service field: got %v, want auth", svc)
	}
}

func TestNewCustom_InvalidPattern(t *testing.T) {
	_, err := NewCustom(CustomFormatDef{Pattern: `(?P<ts>[`})
	if err == nil {
		t.Error("expected error for invalid regex")
	}
}

func TestCustomProbe_NoProbeField(t *testing.T) {
	// No probe_field and no pattern → Probe always returns false
	c, err := NewCustom(CustomFormatDef{
		TSField:    "ts",
		LevelField: "level",
		MsgField:   "msg",
	})
	if err != nil {
		t.Fatal(err)
	}
	if c.Probe([]byte(`{"ts":"2024-04-12T14:23:05Z","level":"info","msg":"hi"}`)) {
		t.Error("expected false when no probe_field and no pattern")
	}
}
