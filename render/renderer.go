package render

import (
	"github.com/AgusRdz/logr/formats"
)

// Config controls output behavior. Passed to all renderers.
type Config struct {
	Color  bool     // emit ANSI codes
	Fields []string // show only these fields (empty = show all)
	Hide   []string // hide these fields
}

// Renderer writes a single Entry to output.
type Renderer interface {
	Write(e formats.Entry)
}
