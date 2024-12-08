package fcqs

import (
	_ "embed"
	"fmt"
	"io"
)

//go:embed shell.bash
var BashTemplate string

func WriteBashScript(w io.Writer) {
	fmt.Fprint(w, BashTemplate)
}
