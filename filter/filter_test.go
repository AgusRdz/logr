package filter

import (
	"testing"
	"time"

	"github.com/AgusRdz/logr/formats"
)

func entry(level, msg string) formats.Entry {
	return formats.Entry{
		Level:   level,
		Message: msg,
		Fields:  map[string]any{},
	}
}

func TestLevelFilter(t *testing.T) {
	f := NewLevelFilter("INFO,ERROR")

	// match
	if !f.Match(entry("INFO", "hello")) {
		t.Error("INFO should match")
	}
	if !f.Match(entry("ERROR", "fail")) {
		t.Error("ERROR should match")
	}

	// no-match
	if f.Match(entry("DEBUG", "debug")) {
		t.Error("DEBUG should not match")
	}
	if f.Match(entry("WARN", "warn")) {
		t.Error("WARN should not match")
	}

	// ParseErr always passes
	errEntry := formats.Entry{ParseErr: true, Raw: []byte("raw"), Message: "raw"}
	if !f.Match(errEntry) {
		t.Error("ParseErr entry should always pass LevelFilter")
	}
}

func TestTimeFilter(t *testing.T) {
	now := time.Now()
	since := now.Add(-1 * time.Hour)
	until := now.Add(1 * time.Hour)
	f := NewTimeFilter(since, until)

	// within range
	e := formats.Entry{Timestamp: now, Fields: map[string]any{}}
	if !f.Match(e) {
		t.Error("now should be within range")
	}

	// before since
	e2 := formats.Entry{Timestamp: now.Add(-2 * time.Hour), Fields: map[string]any{}}
	if f.Match(e2) {
		t.Error("2h ago should be before since")
	}

	// after until
	e3 := formats.Entry{Timestamp: now.Add(2 * time.Hour), Fields: map[string]any{}}
	if f.Match(e3) {
		t.Error("2h in future should be after until")
	}

	// zero timestamp always passes
	e4 := formats.Entry{Timestamp: time.Time{}, Fields: map[string]any{}}
	if !f.Match(e4) {
		t.Error("zero timestamp should always pass")
	}

	// zero since = unbounded lower
	fNoSince := NewTimeFilter(time.Time{}, until)
	e5 := formats.Entry{Timestamp: now.Add(-10 * 24 * time.Hour), Fields: map[string]any{}}
	if !fNoSince.Match(e5) {
		t.Error("no since: 10d ago should pass")
	}

	// zero until = unbounded upper
	fNoUntil := NewTimeFilter(since, time.Time{})
	e6 := formats.Entry{Timestamp: now.Add(365 * 24 * time.Hour), Fields: map[string]any{}}
	if !fNoUntil.Match(e6) {
		t.Error("no until: far future should pass")
	}
}

func TestFieldFilter(t *testing.T) {
	f, err := NewFieldFilter("service=api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// match
	e := formats.Entry{Fields: map[string]any{"service": "api"}}
	if !f.Match(e) {
		t.Error("service=api should match")
	}

	// no-match value
	e2 := formats.Entry{Fields: map[string]any{"service": "web"}}
	if f.Match(e2) {
		t.Error("service=web should not match service=api")
	}

	// missing key
	e3 := formats.Entry{Fields: map[string]any{"other": "api"}}
	if f.Match(e3) {
		t.Error("missing key should not match")
	}

	// invalid kv → error
	_, err2 := NewFieldFilter("no-equals")
	if err2 == nil {
		t.Error("expected error for missing '='")
	}

	// ParseErr always passes
	errEntry := formats.Entry{ParseErr: true, Raw: []byte("raw"), Message: "raw"}
	if !f.Match(errEntry) {
		t.Error("ParseErr entry should always pass FieldFilter")
	}
}

func TestContainsFilter(t *testing.T) {
	f := NewContainsFilter("hello")

	// found in message
	e := formats.Entry{Message: "say hello world", Fields: map[string]any{}}
	if !f.Match(e) {
		t.Error("should match in message")
	}

	// found in field
	e2 := formats.Entry{Message: "other", Fields: map[string]any{"key": "hello there"}}
	if !f.Match(e2) {
		t.Error("should match in field")
	}

	// not found
	e3 := formats.Entry{Message: "goodbye", Fields: map[string]any{"key": "world"}}
	if f.Match(e3) {
		t.Error("should not match when absent")
	}

	// ParseErr: check Raw
	eRaw := formats.Entry{ParseErr: true, Raw: []byte("raw hello line"), Message: "raw hello line"}
	if !f.Match(eRaw) {
		t.Error("ParseErr: should match in Raw")
	}

	eRawMiss := formats.Entry{ParseErr: true, Raw: []byte("raw line"), Message: "raw line"}
	if f.Match(eRawMiss) {
		t.Error("ParseErr: should not match when not in Raw")
	}
}

func TestChain(t *testing.T) {
	level := NewLevelFilter("INFO")
	contains := NewContainsFilter("important")

	chain := Chain{level, contains}

	// both match
	e := formats.Entry{Level: "INFO", Message: "this is important", Fields: map[string]any{}}
	if !chain.Match(e) {
		t.Error("both filters match: chain should pass")
	}

	// level fails
	e2 := formats.Entry{Level: "DEBUG", Message: "this is important", Fields: map[string]any{}}
	if chain.Match(e2) {
		t.Error("level fails: chain should reject")
	}

	// contains fails
	e3 := formats.Entry{Level: "INFO", Message: "nothing here", Fields: map[string]any{}}
	if chain.Match(e3) {
		t.Error("contains fails: chain should reject")
	}

	// empty chain always passes
	emptyChain := Chain{}
	if !emptyChain.Match(e2) {
		t.Error("empty chain should always pass")
	}
}
