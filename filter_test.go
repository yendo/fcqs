package fcqs_test

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yendo/fcqs"
)

func TestFilterWriter(t *testing.T) {
	tests := []struct {
		input        string
		output       string
		isRemoveHead bool
	}{
		{"a\nb\n", "a\nb\n", false},
		{"\na\nb\n\n", "a\nb\n", false},
		{"a\n\n\nb\n", "a\n\nb\n", false},
		{"a\n\n\n\nb\n", "a\n\nb\n", false},
		{"# a\n\n\n\nb\n", "# a\n\nb\n", false},

		{"a\nb\n", "b\n", true},
		{"\na\nb\n\n", "a\nb\n", true},
		{"a\n\n\nb\n", "b\n", true},
		{"a\n\n\n\nb\n", "b\n", true},
		{"# a\n\n\n\nb\n", "b\n", true},
	}
	for _, tc := range tests {
		tc := tc

		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			var Buf bytes.Buffer
			f := fcqs.NewFilter(&Buf, tc.isRemoveHead)

			file := strings.NewReader(tc.input)
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				f.Write(line)
			}
			f.Write("")

			assert.Equal(t, tc.output, Buf.String())
		})
	}
}
