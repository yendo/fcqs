package fcqs

import (
	"fmt"
	"io"
)

type Filter struct {
	WriteBuf     io.Writer
	IsRemoveHead bool
	isDoneTitle  bool
	prevLine     *string
}

// Write writes filtered lines.
func (f *Filter) Write(text string) {
	if f.prevLine == nil {
		// Remove first non blank lines.
		if f.IsRemoveHead && !f.isDoneTitle {
			f.isDoneTitle = true
			return
		}

		// Remove first blank lines.
		if text != "" {
			f.prevLine = &text
		}
		return
	}

	// Remove consecutive blank lines
	if *f.prevLine == "" && text == "" {
		return
	}

	fmt.Fprintln(f.WriteBuf, *f.prevLine)
	f.prevLine = &text
}

func NewFilter(w io.Writer, isRemoveHead bool) Filter {
	return Filter{WriteBuf: w, IsRemoveHead: isRemoveHead}
}
