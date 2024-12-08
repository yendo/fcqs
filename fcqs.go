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

// WriteTitles writes the titles of all notes.
func WriteTitles(w io.Writer, r io.Reader) {
	var allTitles []string
	var title string

	scanner := bufio.NewScanner(r)
	isFenced := false

	for scanner.Scan() {
		line := scanner.Text()

		if !isFenced && strings.HasPrefix(line, "#") {
			if !strings.HasPrefix(strings.TrimLeft(line, "#"), " ") {
				// skip titles without a space after the `#`
				continue
			}

			title = strings.Trim(line, "# ")
			if title == "" {
				continue
			}

		} else if line != "" && title != "" && !slices.Contains(allTitles, title) {
			// skip if content or title is blank
			fmt.Fprintln(w, title)
			allTitles = append(allTitles, title)
		}

		if strings.HasPrefix(line, "```") {
			isFenced = !isFenced
		}
	}
}

// WriteContents writes the contents of the note.
//
//gocyclo:ignore
func WriteContents(w io.Writer, r io.Reader, title string) {
	title = strings.Trim(title, " ")
	_, isNoTitle := os.LookupEnv("FCQS_CONTENTS_NO_TITLE")

	var isScope, isFenced, isBlank bool
	var outputLines int

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		if !isFenced && isContentsBegin(line, title) {
			isScope = true
		} else if isScope {
			switch {
			case !isFenced && isContentsEnd(line):
				isScope = false
			case strings.HasPrefix(line, "```"):
				isFenced = !isFenced
			case line == "":
				isBlank = true
			}
		}

		if isScope && line != "" {
			if isBlank {
				isBlank = false

				if !isNoTitle || outputLines > 1 {
					fmt.Fprintln(w, "")
				}
			}

			if !isNoTitle || outputLines > 0 {
				fmt.Fprintln(w, line)
			}
			outputLines++
		}
	}
}

// isContentsBegin returns if the line is the beginning of the contents.
func isContentsBegin(line string, title string) bool {
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

// isContentsEnd returns if the line is the end of the contents.
func isContentsEnd(line string) bool {
	// Title must start with #.
	if !strings.HasPrefix(line, "#") {
		return false
	}

	// Title may be blank.
	if strings.Trim(line, "# ") == "" {
		return true
	}

	// Title must have a space after #
	return strings.HasPrefix(strings.TrimLeft(line, "#"), " ")
}

// WriteFirstURL writes the first URL in the contents of the note.
func WriteFirstURL(w io.Writer, r io.Reader, title string) {
	var buf bytes.Buffer

	WriteContents(&buf, r, title)

	rxStrict := xurls.Strict()
	url := rxStrict.FindString(buf.String())

	if url != "" {
		fmt.Fprintln(w, url)
	}
}

// WriteFirstCmdLine writes the first command-line in the contents of the note.
func WriteFirstCmdLine(w io.Writer, r io.Reader, title string) {
	var buf bytes.Buffer

	isFenced := false

	WriteContents(&buf, r, title)
	scanner := bufio.NewScanner(&buf)

	for scanner.Scan() {
		line := scanner.Text()

		if isShellCodeBlockBegin(line) {
			isFenced = true
			continue
		} else if strings.HasPrefix(line, "```") && isFenced {
			break
		}

		if isFenced {
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
	c := 0
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		c++
		line := scanner.Text()

		if isContentsBegin(line, title) {
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
