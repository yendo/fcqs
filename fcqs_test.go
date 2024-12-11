package fcqs_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yendo/fcqs"
	"github.com/yendo/fcqs/test"
)

func TestWriteTitles(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	file := test.OpenTestNotesFile(t, test.TestNotesFile)
	fcqs.WriteTitles(&buf, file)

	assert.Equal(t, test.GetExpectedTitles(), buf.String())
}

func TestWriteContents(t *testing.T) {
	tests := []struct {
		title    string
		contents string
	}{
		{"# title\n", "contents\n"},
		{"# Long title and contents have lines\n", "line 1\n\nline 2\n"},
		{"# Regular expression meta chars in the title are ignored $\n", "contents\n"},
		{"# Consecutive blank lines are combined into a single line\n", "line 1\n\nline 2\n"},
		{"# same title\n", "Contents with the same title are combined into one.\n\n" +
			"# same title\n\n2nd\n\n" + "# same title\n\n3rd\n"},
		{"## Heading levels and structures are ignored\n", "contents\n"},
		{"# Trailing spaces in the title are ignored  \n", "The contents have trailing spaces.  \n"},
		{"# Notes without content output the title only", ""},
		{"#   Spaces before the title are ignored\n", "contents\n"},
		{"# Headings in fenced code blocks are ignored\n", "```\n" + "# fenced heading\n" + "```\n"},
		{"# There can be no blank line", "contents\n"},
		{"#\n", "no title contents are combined into one.\n\n" + "#  \n\n" + "title is only spaces\n"},
		{"# Titles without a space after the # are not recognized\n", "#no_space_title\n\n" + "contents\n\n" +
			"  # Titles with spaces before the # are not recognized\n\n" + "contents\n"},
		{"# URL\n", "fcqs: http://github.com/yendo/fcqs/\n" + "github: http://github.com/\n"},
		{"# command-line\n", "```sh\n" + "ls -l | nl\n" + "```\n"},
		{"# command-line with $\n", "```console\n" + "$ date\n" + "```\n"},
	}

	t.Run("contents with title", func(t *testing.T) {
		t.Setenv("FCQS_CONTENTS_NO_TITLE", "")
		os.Unsetenv("FCQS_CONTENTS_NO_TITLE")

		for _, tc := range tests {
			tc := tc

			t.Run(tc.title, func(t *testing.T) {
				t.Parallel()

				file := test.OpenTestNotesFile(t, test.TestNotesFile)
				var buf bytes.Buffer
				title := strings.TrimRight(strings.TrimLeft(tc.title, "# "), "\n")

				fcqs.WriteContents(&buf, file, title)

				assert.Equal(t, tc.title+"\n"+tc.contents, buf.String())
			})
		}
	})

	t.Run("contents without title", func(t *testing.T) {
		t.Setenv("FCQS_CONTENTS_NO_TITLE", "1")

		for _, tc := range tests {
			tc := tc

			t.Run(tc.title, func(t *testing.T) {
				t.Parallel()

				file := test.OpenTestNotesFile(t, test.TestNotesFile)
				var buf bytes.Buffer
				title := strings.TrimRight(strings.TrimLeft(tc.title, "# "), "\n")

				fcqs.WriteContents(&buf, file, title)

				assert.Equal(t, tc.contents, buf.String())
			})
		}
	})
}

func TestWriteNoContents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc  string
		title string
	}{
		{
			desc:  "Titles without a space after the `#` are not recognized as title",
			title: "#no_space_title",
		},
		{
			desc:  "Titles with spaces before the # are not recognized as title",
			title: "  # Titles with spaces before the # are not recognized",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			file := test.OpenTestNotesFile(t, test.TestNotesFile)

			fcqs.WriteContents(&buf, file, strings.TrimLeft(tc.title, "#"))

			assert.Empty(t, buf.String())
		})
	}
}

func TestWriteFirstURL(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	file := test.OpenTestNotesFile(t, test.TestNotesFile)

	fcqs.WriteFirstURL(&buf, file, "URL")

	assert.Equal(t, "http://github.com/yendo/fcqs/\n", buf.String())
}

func TestWriteFirstCmdLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		title  string
		output bool
	}{
		{"shell 1", true},
		{"shell 2", true},
		{"shell 3", true},
		{"sh", true},
		{"shell-script", true},
		{"bash", true},
		{"zsh", true},
		{"powershell", true},
		{"posh", true},
		{"pwsh", true},
		{"shellsession", true},
		{"bash session", true},
		{"console", true},
		{"go", false},
		{"no identifier", false},
		{"other identifier", false},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			file := test.OpenTestNotesFile(t, test.TestShellBlockFile)

			fcqs.WriteFirstCmdLineBlock(&buf, file, tc.title)

			expected := map[bool]string{true: "ls -l | nl\n", false: ""}
			assert.Equal(t, expected[tc.output], buf.String())
		})
	}
}

func TestWriteNoteLocation(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	testFile := test.OpenTestNotesFile(t, test.TestLocationFile)
	fcqs.WriteNoteLocation(&buf, testFile, "5th Line")

	assert.Equal(t, fmt.Sprintf("%q 5\n", testFile.Name()), buf.String())
}

func TestGetFcqsFile(t *testing.T) {
	t.Run("cannot access user home directory", func(t *testing.T) {
		t.Setenv("FCQS_NOTES_FILE", "")
		t.Setenv("HOME", "")

		fileName, err := fcqs.GetNotesFileName()

		assert.Empty(t, fileName)
		assert.Error(t, err)
		assert.EqualError(t, err, "cannot access user home directory: $HOME is not defined")
	})

	t.Run("set from environment variable", func(t *testing.T) {
		expectedFileName := "test_file_from_env.md"
		t.Setenv("FCQS_NOTES_FILE", expectedFileName)

		fileName, err := fcqs.GetNotesFileName()

		assert.NoError(t, err)
		assert.Equal(t, expectedFileName, fileName)
	})

	t.Run("default filename", func(t *testing.T) {
		t.Setenv("FCQS_NOTES_FILE", "")
		home, err := os.UserHomeDir()
		require.NoError(t, err)

		filename, err := fcqs.GetNotesFileName()

		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(home, fcqs.DefaultNotesFile), filename)
	})
}
