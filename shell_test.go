package fcqs_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yendo/fcqs"
)

func TestWriteBashScript(t *testing.T) {
	t.Parallel()

	fileName := "shell.bash"
	data, err := os.ReadFile(fileName)
	require.NoError(t, err)
	expected := string(data)

	var buf bytes.Buffer
	fcqs.WriteBashScript(&buf)

	assert.Equal(t, expected, buf.String())
}
