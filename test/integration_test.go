package test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const defaultNotesFile = "fcnotes.md"

type stdBuf struct {
	stdout bytes.Buffer
	stderr bytes.Buffer
}

func (b *stdBuf) newTestCmd(args ...string) *exec.Cmd {
	cmd := exec.Command("./fcqs-cli", args...)

	cmd.Env = append(os.Environ(), "GOCOVERDIR=../coverdir")
	cmd.Stdout = &b.stdout
	cmd.Stderr = &b.stderr

	return cmd
}

func TestCmdSuccess(t *testing.T) {
	t.Setenv("FCQS_CONTENTS_NO_TITLE", "")
	os.Unsetenv("FCQS_CONTENTS_NO_TITLE")
	t.Setenv("FCQS_NOTES_FILE", TestNotesFile)

	tests := []struct {
		title   string
		options []string
		stdout  string
	}{
		{
			title:   "with version flag",
			options: []string{"-v"},
			stdout:  "0.0.0-test\n",
		},
		{
			title:   "with url flag and an arg",
			options: []string{"-u", "URL"},
			stdout:  "http://github.com/yendo/fcqs/\n",
		},
		{
			title:   "with cmd flag and an arg",
			options: []string{"-c", "command-line"},
			stdout:  "ls -l | nl\n",
		},
		{
			title:   "with location flag and an arg",
			options: []string{"-l", "title"},
			stdout:  fmt.Sprintf("%q 1\n", TestNotesFile),
		},
		{
			title:   "without args",
			options: []string{},
			stdout:  GetExpectedTitles(),
		},
		{
			title:   "with an arg",
			options: []string{"There can be no blank line"},
			stdout:  "# There can be no blank line\ncontents\n", // next line is only "#".
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			buf := &stdBuf{}
			cmd := buf.newTestCmd(tc.options...)

			err := cmd.Run()

			assert.NoError(t, err)
			assert.Equal(t, tc.stdout, buf.stdout.String())
			assert.Empty(t, buf.stderr.String())
		})
	}
}

func TestCmdWriteContentsWithoutTitle(t *testing.T) {
	t.Setenv("FCQS_NOTES_FILE", TestNotesFile)
	t.Setenv("FCQS_CONTENTS_NO_TITLE", "1")
	buf := &stdBuf{}
	cmd := buf.newTestCmd("There can be no blank line")

	err := cmd.Run()

	assert.NoError(t, err)
	assert.Equal(t, "contents\n", buf.stdout.String())
	assert.Empty(t, buf.stderr.String())
}

func TestCmdFail(t *testing.T) {
	t.Setenv("FCQS_NOTES_FILE", TestNotesFile)

	tests := []struct {
		title   string
		options []string
		stderr  string
	}{
		{
			title:   "with url flag and no arg",
			options: []string{"-u"},
			stderr:  "invalid number of arguments\n",
		},
		{
			title:   "with cmd flag and no arg",
			options: []string{"-c"},
			stderr:  "invalid number of arguments\n",
		},
		{
			title:   "with location flag and no arg",
			options: []string{"-l"},
			stderr:  "invalid number of arguments\n",
		},
		{
			title:   "with two args",
			options: []string{"title", "other"},
			stderr:  "invalid number of arguments\n",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			buf := &stdBuf{}
			cmd := buf.newTestCmd(tc.options...)

			err := cmd.Run()

			assert.Error(t, err)
			assert.Empty(t, buf.stdout.String())
			assert.Equal(t, tc.stderr, buf.stderr.String())
		})
	}
}

func TestCmdNotesLocation(t *testing.T) {
	t.Setenv("FCQS_NOTES_FILE", TestLocationFile)
	buf := &stdBuf{}
	cmd := buf.newTestCmd("-l", "5th Line")

	err := cmd.Run()

	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%q 5\n", TestLocationFile), buf.stdout.String())
	assert.Empty(t, buf.stderr.String())
}

func TestUserHomeDirNotExists(t *testing.T) {
	t.Setenv("FCQS_NOTES_FILE", "")
	t.Setenv("HOME", "")
	buf := &stdBuf{}
	cmd := buf.newTestCmd()

	err := cmd.Run()

	assert.Error(t, err)
	assert.Empty(t, buf.stdout.String())
	assert.Equal(t, "cannot get notes file name: cannot access user home directory: $HOME is not defined\n", buf.stderr.String())
}

func TestNotesNotExists(t *testing.T) {
	t.Setenv("FCQS_NOTES_FILE", "not_exists")
	buf := &stdBuf{}
	cmd := buf.newTestCmd()

	err := cmd.Run()

	assert.Error(t, err)
	assert.Empty(t, buf.stdout.String())
	assert.Equal(t, "cannot access notes file: open not_exists: no such file or directory\n", buf.stderr.String())
}

func TestDefaultNoteExists(t *testing.T) {
	// Arrange
	t.Setenv("FCQS_NOTES_FILE", "")

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	if _, err = os.Stat(filepath.Join(home, defaultNotesFile)); err != nil {
		t.Skipf("the default %s does not exist", defaultNotesFile)
	}

	buf := &stdBuf{}
	cmd := buf.newTestCmd()

	// Act
	err = cmd.Run()

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, buf.stderr.String())
}

func TestBashScript(t *testing.T) {
	t.Parallel()

	// Arrange
	fileName := "../shell.bash"
	data, err := os.ReadFile(fileName)
	require.NoError(t, err)
	expected := string(data)

	buf := &stdBuf{}
	cmd := buf.newTestCmd("--bash")

	// Act
	err = cmd.Run()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expected, buf.stdout.String())
	assert.Empty(t, buf.stderr.String())
}
