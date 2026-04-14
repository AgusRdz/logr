package render

import (
	"fmt"
	"io"

	"github.com/AgusRdz/logr/formats"
)

// Raw passes through the original log line unchanged.
type Raw struct {
	w io.Writer
}

func NewRaw(w io.Writer) *Raw {
	return &Raw{w: w}
}

func (r *Raw) Write(e formats.Entry) {
	fmt.Fprintf(r.w, "%s\n", e.Raw)
}
