package fcqs

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"mvdan.cc/xurls/v2"
)

const DefaultNotesFile = "fcnotes.md"

// State of text line.
const (
	Normal = iota
	Fenced
	Scoped
	ScopedFenced
)

// WriteTitles writes the titles of all notes.
func WriteTitles(w io.Writer, r io.Reader) {
	var allTitles []string
	var title string

	scanner := bufio.NewScanner(r)
	state := Normal

	for scanner.Scan() {
		line := scanner.Text()

		switch state {
		case Normal:
			if isTitleLine(line) {
				title = strings.Trim(line, "# ")
			} else if line != "" {
				if title != "" && !slices.Contains(allTitles, title) {
					fmt.Fprintln(w, title)
					allTitles = append(allTitles, title)
				}

				if strings.HasPrefix(line, "```") {
					state = Fenced
				}
			}

		case Fenced:
			if strings.HasPrefix(line, "```") {
				state = Normal
			}
		}
	}
}

// WriteContents writes the contents of the note.
func WriteContents(w io.Writer, r io.Reader, title string) {
	title = strings.Trim(title, " ")
	if title == "" {
		return
	}

	state := Normal

	_, isNoTitle := os.LookupEnv("FCQS_CONTENTS_NO_TITLE")
	f := NewFilter(w, isNoTitle)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		switch state {
		case Normal:
			if strings.HasPrefix(line, "```") {
				state = Fenced
			}

			if isSearchedTitleLine(line, title) {
				state = Scoped
				f.Write(line)
			}

		case Fenced:
			if strings.HasPrefix(line, "```") {
				state = Normal
			}

		case Scoped:
			if isTitleLine(line) && !isSearchedTitleLine(line, title) {
				state = Normal
				break
			}

			if strings.HasPrefix(line, "```") {
				state = ScopedFenced
			}

			f.Write(line)

		case ScopedFenced:
			if strings.HasPrefix(line, "```") {
				state = Scoped
			}
			f.Write(line)
		}
	}

	f.Write("")
}

// isTitleLine returns if the line is title line.
func isTitleLine(line string) bool {
	// Title line must start with #.
	if !strings.HasPrefix(line, "#") {
		return false
	}

	// Blank title is valid.
	if strings.TrimLeft(line, "# ") == "" {
		return true
	}

	// The title line must have a space after #.
	return strings.HasPrefix(strings.TrimLeft(line, "#"), " ")
}

// isSearchedTitleLine returns if the line is the searched title line.
func isSearchedTitleLine(line string, title string) bool {
	// Title must start with #.
	if !strings.HasPrefix(line, "#") {
		return false
	}

	// When the title is not blank, the title must have a space after #.
	if title != "" && !strings.HasPrefix(strings.TrimLeft(line, "#"), " ") {
		return false
	}

	// When the trimmed line and title match, the content starts.
	return strings.Trim(line, "# ") == strings.Trim(title, "# ")
}

// WriteFirstURL writes the first URL in the contents of the note.
func WriteFirstURL(w io.Writer, r io.Reader, title string) {
	if isEmptyTrimmedTitle(title) {
		return
	}

	var buf bytes.Buffer
	WriteContents(&buf, r, title)

	rxStrict := xurls.Strict()
	url := rxStrict.FindString(buf.String())

	if url != "" {
		fmt.Fprintln(w, url)
	}
}

// WriteFirstCmdLineBlock writes the first command-line block in the contents of the note.
func WriteFirstCmdLineBlock(w io.Writer, r io.Reader, title string) {
	if isEmptyTrimmedTitle(title) {
		return
	}

	state := Normal

	var buf bytes.Buffer
	WriteContents(&buf, r, title)
	scanner := bufio.NewScanner(&buf)

	for scanner.Scan() {
		line := scanner.Text()

		switch state {
		case Normal:
			if isShellCodeBlockBegin(line) {
				state = Fenced
			}

		case Fenced:
			if strings.HasPrefix(line, "```") {
				break
			}
			fmt.Fprintln(w, strings.TrimLeft(line, "$ "))
		}
	}
}

var reShellCodeBlock = regexp.MustCompile("^```+\\s*(\\S+).*$")

// isShellCodeBlockBegin determines if the line is the beginning of a shell code block.
func isShellCodeBlockBegin(line string) bool {
	shellList := []string{
		"shell", "sh", "shell-script", "bash", "zsh",
		"powershell", "posh", "pwsh",
		"shellsession", "console",
	}

	match := reShellCodeBlock.FindStringSubmatch(line)
	if len(match) == 0 {
		return false
	}

	return slices.Contains(shellList, match[1])
}

// WriteNoteLocation writes the file name and line number of the note.
func WriteNoteLocation(w io.Writer, file *os.File, title string) {
	if isEmptyTrimmedTitle(title) {
		return
	}

	c := 0
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		c++
		line := scanner.Text()

		if isSearchedTitleLine(line, title) {
			fmt.Fprintf(w, "%q %d\n", file.Name(), c)
			break
		}
	}
}

// GetNotesFileName returns the filename of the notes.
func GetNotesFileName() (string, error) {
	fileName := os.Getenv("FCQS_NOTES_FILE")
	if fileName != "" {
		return fileName, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot access user home directory: %w", err)
	}

	fileName = filepath.Join(home, DefaultNotesFile)
	return fileName, nil
}

// isEmptyTrimmedTitle determines if trimmed tile is empty.
func isEmptyTrimmedTitle(title string) bool {
	return strings.Trim(title, " ") == ""
}
