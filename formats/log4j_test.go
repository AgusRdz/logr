package formats

import "testing"

func TestLog4jProbe(t *testing.T) {
	f := Log4j{}

	// valid log4j2
	if !f.Probe([]byte(`{"timeMillis":1712930585000,"thread":"main","level":"INFO","loggerName":"com.example.App","message":"started"}`)) {
		t.Error("expected true for valid log4j2")
	}

	// missing loggerName
	if f.Probe([]byte(`{"timeMillis":1712930585000,"level":"INFO","message":"hi"}`)) {
		t.Error("expected false for missing loggerName")
	}

	// missing timeMillis
	if f.Probe([]byte(`{"loggerName":"com.example.App","level":"INFO","message":"hi"}`)) {
		t.Error("expected false for missing timeMillis")
	}

	// invalid JSON
	if f.Probe([]byte(`not json`)) {
		t.Error("expected false for invalid JSON")
	}
}

func TestLog4jParse(t *testing.T) {
	f := Log4j{}

	e := f.Parse([]byte(`{"timeMillis":1712930585000,"thread":"main","level":"INFO","loggerName":"com.example.App","message":"Application started","endOfBatch":false,"loggerFqcn":"org.apache.logging.log4j.spi.AbstractLogger","contextMap":{}}`))
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
	if e.Timestamp.Unix() != 1712930585 {
		t.Errorf("ts unix = %d, want 1712930585", e.Timestamp.Unix())
	}

	// noise fields removed
	for _, k := range []string{"timeMillis", "message", "level", "loggerFqcn", "endOfBatch", "contextMap"} {
		if _, ok := e.Fields[k]; ok {
			t.Errorf("field %q should be removed", k)
		}
	}

	// thread and loggerName kept
	if _, ok := e.Fields["loggerName"]; !ok {
		t.Error("loggerName should be in fields")
	}

	// WARN
	e2 := f.Parse([]byte(`{"timeMillis":1712930586000,"thread":"t","level":"WARN","loggerName":"com.example","message":"warn"}`))
	if e2.Level != "WARN" {
		t.Errorf("WARN: got %q", e2.Level)
	}
}
