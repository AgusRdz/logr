package stats

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/AgusRdz/logr/formats"
)

func makeEntry(level, msg string, ts time.Time) formats.Entry {
	return formats.Entry{
		Level:     level,
		Message:   msg,
		Timestamp: ts,
	}
}

func TestStats_Add(t *testing.T) {
	s := New()
	now := time.Now()

	s.Add(makeEntry("INFO", "started", now))
	s.Add(makeEntry("INFO", "request", now.Add(time.Minute)))
	s.Add(makeEntry("ERROR", "Payment gateway timeout", now.Add(2*time.Minute)))
	s.Add(makeEntry("ERROR", "Payment gateway timeout", now.Add(3*time.Minute)))
	s.Add(makeEntry("FATAL", "DB pool exhausted", now.Add(4*time.Minute)))
	s.Add(makeEntry("WARN", "slow query", now.Add(5*time.Minute)))

	if s.Total != 6 {
		t.Errorf("expected Total=6, got %d", s.Total)
	}
	if s.ByLevel["INFO"] != 2 {
		t.Errorf("expected INFO=2, got %d", s.ByLevel["INFO"])
	}
	if s.ByLevel["ERROR"] != 2 {
		t.Errorf("expected ERROR=2, got %d", s.ByLevel["ERROR"])
	}
	if s.ByLevel["FATAL"] != 1 {
		t.Errorf("expected FATAL=1, got %d", s.ByLevel["FATAL"])
	}
	if s.ByLevel["WARN"] != 1 {
		t.Errorf("expected WARN=1, got %d", s.ByLevel["WARN"])
	}
	if s.errMsgs["Payment gateway timeout"] != 2 {
		t.Errorf("expected error msg count=2, got %d", s.errMsgs["Payment gateway timeout"])
	}
	if s.errMsgs["DB pool exhausted"] != 1 {
		t.Errorf("expected error msg count=1, got %d", s.errMsgs["DB pool exhausted"])
	}
}

func TestStats_Print(t *testing.T) {
	s := New()
	now := time.Now()

	for i := 0; i < 1000; i++ {
		s.Add(makeEntry("INFO", "ok", now))
	}
	for i := 0; i < 10; i++ {
		s.Add(makeEntry("ERROR", "Payment gateway timeout", now))
	}
	for i := 0; i < 50; i++ {
		s.Add(makeEntry("WARN", "slow", now))
	}

	var buf bytes.Buffer
	s.Print(&buf, false)
	out := buf.String()

	if !strings.Contains(out, "total") {
		t.Error("missing 'total' in output")
	}
	if !strings.Contains(out, "1,060") {
		t.Errorf("expected formatted total '1,060', got:\n%s", out)
	}
	if !strings.Contains(out, "ERROR") {
		t.Error("missing ERROR level in output")
	}
	if !strings.Contains(out, "WARN") {
		t.Error("missing WARN level in output")
	}
	if !strings.Contains(out, "INFO") {
		t.Error("missing INFO level in output")
	}
	if !strings.Contains(out, "█") {
		t.Error("missing progress bar in output")
	}
	if !strings.Contains(out, "top errors") {
		t.Error("missing top errors section")
	}
	if !strings.Contains(out, "Payment gateway timeout") {
		t.Error("missing top error message in output")
	}
}
