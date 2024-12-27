package fcqs

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
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
func WriteTitles(w io.Writer, r io.Reader) error {
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
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("seek titles: %w", err)
	}

	return nil
}

// WriteContents writes the contents of the note.
func WriteContents(w io.Writer, r io.Reader, title string, isNoTitle bool) error {
	title = strings.Trim(title, " ")
	if title == "" {
		return nil
	}

	state := Normal

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
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("seek contents: %w", err)
	}

	f.Write("")
	return nil
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
func WriteFirstURL(w io.Writer, r io.Reader, title string) error {
	if isEmptyTrimmedTitle(title) {
		return nil
	}

	var buf bytes.Buffer
	if err := WriteContents(&buf, r, title, false); err != nil {
		return err
	}

	rxStrict := xurls.Strict()
	url := rxStrict.FindString(buf.String())

	if url != "" {
		fmt.Fprintln(w, url)
	}

	return nil
}

// WriteFirstCmdLineBlock writes the first command-line block in the contents of the note.
func WriteFirstCmdLineBlock(w io.Writer, r io.Reader, title string) error {
	if isEmptyTrimmedTitle(title) {
		return nil
	}

	state := Normal

	var buf bytes.Buffer
	if err := WriteContents(&buf, r, title, false); err != nil {
		return err
	}
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
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("seek command line block: %w", err)
	}

	return nil
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
func WriteNoteLocation(w io.Writer, files []*os.File, title string) error {
	if isEmptyTrimmedTitle(title) {
		return nil
	}

	for _, file := range files {
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
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("seek note location: %w", err)
		}
	}
	return nil
}

// NotesFileName returns the filename of the notes.
func NotesFileName() ([]string, error) {
	var fileNames []string

	f := os.Getenv("FCQS_NOTES_FILE")
	if f != "" {
		sep := ":"
		if runtime.GOOS == "windows" {
			sep = ";"
		}

		fileNames = strings.Split(f, sep)
		return fileNames, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("user home directory: %w", err)
	}

	fileNames = append(fileNames, filepath.Join(home, DefaultNotesFile))
	return fileNames, nil
}

// isEmptyTrimmedTitle determines if trimmed tile is empty.
func isEmptyTrimmedTitle(title string) bool {
	return strings.Trim(title, " ") == ""
}
