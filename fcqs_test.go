package fcqs_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yendo/fcqs"
	"github.com/yendo/fcqs/internal/value"
	"github.com/yendo/fcqs/test"
)

func TestWriteTitles(t *testing.T) {
	t.Parallel()

	t.Run("success to output titles", func(t *testing.T) {
		t.Parallel()

		file := test.OpenTestNotesFile(t, test.NotesFile)

		var buf bytes.Buffer
		err := fcqs.WriteTitles(&buf, file)

		assert.NoError(t, err)
		assert.Equal(t, test.ExpectedTitles(), buf.String())
	})

	t.Run("fail with scan error", func(t *testing.T) {
		t.Parallel()

		errStr := "scan error"
		file := iotest.ErrReader(errors.New(errStr))

		var buf bytes.Buffer
		err := fcqs.WriteTitles(&buf, file)

		assert.EqualError(t, err, fmt.Sprintf("seek titles: %s", errStr))
		assert.Empty(t, buf.String())
	})
}

func TestWriteContents(t *testing.T) {
	t.Parallel()

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
		{"# Titles without a space after the # are not recognized\n", "#no_space_title\n\n" + "contents\n\n" +
			"  # Titles with spaces before the # are not recognized\n\n" + "contents\n"},
		{"# URL\n", "fcqs: http://github.com/yendo/fcqs/\n" + "github: http://github.com/\n"},
		{"# command-line\n", "```sh\n" + "ls -l | nl\n" + "```\n"},
		{"# command-line with $\n", "```console\n" + "$ date\n" + "```\n"},
	}

	t.Run("contents with title", func(t *testing.T) {
		t.Parallel()

		for _, tc := range tests {
			t.Run(tc.title, func(t *testing.T) {
				t.Parallel()

				file := test.OpenTestNotesFile(t, test.NotesFile)
				titleStr := strings.TrimRight(strings.TrimLeft(tc.title, "# "), "\n")
				title, err := value.NewTitle(titleStr)
				require.NoError(t, err)

				var buf bytes.Buffer
				err = fcqs.WriteContents(&buf, file, title, false)

				assert.NoError(t, err)
				assert.Equal(t, tc.title+"\n"+tc.contents, buf.String())
			})
		}
	})

	t.Run("contents without title", func(t *testing.T) {
		t.Parallel()

		for _, tc := range tests {
			t.Run(tc.title, func(t *testing.T) {
				t.Parallel()

				file := test.OpenTestNotesFile(t, test.NotesFile)
				titleStr := strings.TrimRight(strings.TrimLeft(tc.title, "# "), "\n")
				title, err := value.NewTitle(titleStr)
				require.NoError(t, err)

				var buf bytes.Buffer
				err = fcqs.WriteContents(&buf, file, title, true)

				assert.NoError(t, err)
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
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			file := test.OpenTestNotesFile(t, test.NotesFile)
			titleStr := strings.TrimLeft(tc.title, "#")
			title, err := value.NewTitle(titleStr)
			require.NoError(t, err)

			var buf bytes.Buffer
			err = fcqs.WriteContents(&buf, file, title, false)

			assert.NoError(t, err)
			assert.Empty(t, buf.String())
		})
	}

	t.Run("scan failed", func(t *testing.T) {
		t.Parallel()

		errStr := "scan error"
		file := iotest.ErrReader(errors.New(errStr))
		title, err := value.NewTitle("title")
		require.NoError(t, err)

		var buf bytes.Buffer
		err = fcqs.WriteContents(&buf, file, title, false)

		assert.EqualError(t, err, fmt.Sprintf("seek contents: %s", errStr))
		assert.Empty(t, buf.String())
	})
}

func TestWriteFirstURL(t *testing.T) {
	t.Parallel()

	title, err := value.NewTitle("URL")
	require.NoError(t, err)

	t.Run("scan succeeded", func(t *testing.T) {
		t.Parallel()

		file := test.OpenTestNotesFile(t, test.NotesFile)

		var buf bytes.Buffer
		err = fcqs.WriteFirstURL(&buf, file, title)

		assert.NoError(t, err)
		assert.Equal(t, "http://github.com/yendo/fcqs/\n", buf.String())
	})

	t.Run("scan failed", func(t *testing.T) {
		t.Parallel()

		errStr := "scan error"
		file := iotest.ErrReader(errors.New(errStr))

		var buf bytes.Buffer
		err = fcqs.WriteFirstURL(&buf, file, title)

		assert.EqualError(t, err, fmt.Sprintf("seek contents: %s", errStr))
		assert.Empty(t, buf.String())
	})
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
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			file := test.OpenTestNotesFile(t, test.ShellBlockFile)
			title, err := value.NewTitle(tc.title)
			require.NoError(t, err)

			var buf bytes.Buffer
			err = fcqs.WriteFirstCmdLineBlock(&buf, file, title)

			assert.NoError(t, err)
			expected := map[bool]string{true: "ls -l | nl\n", false: ""}
			assert.Equal(t, expected[tc.output], buf.String())
		})
	}

	t.Run("scan failed", func(t *testing.T) {
		t.Parallel()

		errStr := "scan error"
		r := iotest.ErrReader(errors.New(errStr))
		title, err := value.NewTitle("title")
		require.NoError(t, err)

		var buf bytes.Buffer
		err = fcqs.WriteFirstCmdLineBlock(&buf, r, title)

		assert.EqualError(t, err, fmt.Sprintf("seek contents: %s", errStr))
		assert.Empty(t, buf.String())
	})
}

func TestWriteNoteLocation(t *testing.T) {
	t.Parallel()

	t.Run("single file", func(t *testing.T) {
		t.Parallel()

		var testFiles []*os.File
		testFile := test.OpenTestNotesFile(t, test.LocationFile)
		testFiles = append(testFiles, testFile)

		title, err := value.NewTitle("5th Line")
		require.NoError(t, err)

		var buf bytes.Buffer
		err = fcqs.WriteNoteLocation(&buf, testFiles, title)

		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%q 5\n", testFile.Name()), buf.String())
	})

	t.Run("multi files", func(t *testing.T) {
		t.Parallel()

		var testFiles []*os.File
		testFile1 := test.OpenTestNotesFile(t, test.LocationFile)
		testFile2 := test.OpenTestNotesFile(t, test.LocationExtraFile)
		testFiles = append(testFiles, testFile1, testFile2)

		title, err := value.NewTitle("9th Line")
		require.NoError(t, err)

		var buf bytes.Buffer
		err = fcqs.WriteNoteLocation(&buf, testFiles, title)

		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%q 9\n", testFile2.Name()), buf.String())
	})
}

func TestNotesFileName(t *testing.T) {
	t.Run("failed to access user home directory", func(t *testing.T) {
		t.Setenv("FCQS_NOTES_FILE", "")
		t.Setenv("HOME", "")

		fileName, err := fcqs.NotesFileName()

		assert.Empty(t, fileName)
		assert.Error(t, err)
		assert.EqualError(t, err, "user home directory: $HOME is not defined")
	})

	t.Run("set a file from environment variable", func(t *testing.T) {
		expectedFileName := "test_file_from_env.md"
		t.Setenv("FCQS_NOTES_FILE", expectedFileName)

		fileName, err := fcqs.NotesFileName()

		assert.NoError(t, err)
		assert.Equal(t, expectedFileName, fileName[0])
	})

	t.Run("set files from environment variable", func(t *testing.T) {
		expectedFileNames := []string{"test_file_1.md", "test_file_2.md", "test_file_3.md"}
		t.Setenv("FCQS_NOTES_FILE", strings.Join(expectedFileNames, test.FileSeparator()))

		fileName, err := fcqs.NotesFileName()

		assert.NoError(t, err)
		assert.ElementsMatch(t, expectedFileNames, fileName)
	})

	t.Run("default filename", func(t *testing.T) {
		t.Setenv("FCQS_NOTES_FILE", "")
		home, err := os.UserHomeDir()
		require.NoError(t, err)

		filename, err := fcqs.NotesFileName()

		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(home, fcqs.DefaultNotesFile), filename[0])
	})
}
