package filter

import (
	"fmt"
	"strings"
	"time"

	"github.com/AgusRdz/logr/formats"
)

// Filter matches log entries.
type Filter interface {
	Match(e formats.Entry) bool
}

// Chain applies multiple filters with AND logic (short-circuit).
type Chain []Filter

func (c Chain) Match(e formats.Entry) bool {
	for _, f := range c {
		if !f.Match(e) {
			return false
		}
	}
	return true
}

// LevelFilter passes entries whose level is in the allowed set.
type LevelFilter struct {
	allowed map[string]bool
}

func NewLevelFilter(levels string) *LevelFilter {
	allowed := map[string]bool{}
	for _, l := range strings.Split(levels, ",") {
		l = strings.TrimSpace(l)
		if l != "" {
			allowed[strings.ToUpper(l)] = true
		}
	}
	return &LevelFilter{allowed: allowed}
}

func (f *LevelFilter) Match(e formats.Entry) bool {
	if e.ParseErr {
		return true
	}
	return f.allowed[e.Level]
}

// TimeFilter passes entries within the [since, until] window.
type TimeFilter struct {
	since, until time.Time
}

func NewTimeFilter(since, until time.Time) *TimeFilter {
	return &TimeFilter{since: since, until: until}
}

func (f *TimeFilter) Match(e formats.Entry) bool {
	if e.Timestamp.IsZero() {
		return true
	}
	if !f.since.IsZero() && e.Timestamp.Before(f.since) {
		return false
	}
	if !f.until.IsZero() && e.Timestamp.After(f.until) {
		return false
	}
	return true
}

// FieldFilter passes entries where a specific field equals a value.
type FieldFilter struct {
	key, value string
}

func NewFieldFilter(kv string) (*FieldFilter, error) {
	parts := strings.SplitN(kv, "=", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("field filter %q missing '='", kv)
	}
	return &FieldFilter{key: parts[0], value: parts[1]}, nil
}

func (f *FieldFilter) Match(e formats.Entry) bool {
	if e.ParseErr {
		return true
	}
	v, ok := e.Fields[f.key]
	if !ok {
		return false
	}
	return fmt.Sprintf("%v", v) == f.value
}

// ContainsFilter passes entries where the substring appears in message or fields.
type ContainsFilter struct {
	substr string
}

func NewContainsFilter(s string) *ContainsFilter {
	return &ContainsFilter{substr: s}
}

func (f *ContainsFilter) Match(e formats.Entry) bool {
	if e.ParseErr {
		return strings.Contains(string(e.Raw), f.substr)
	}
	if strings.Contains(e.Message, f.substr) {
		return true
	}
	for _, v := range e.Fields {
		if s, ok := v.(string); ok {
			if strings.Contains(s, f.substr) {
				return true
			}
		}
	}
	return false
}
