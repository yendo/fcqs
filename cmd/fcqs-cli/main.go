package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	flag "github.com/spf13/pflag"
	"github.com/yendo/fcqs"
)

var (
	version = "unknown"

	showVersion = flag.BoolP("version", "v", false, "output the version")
	showURL     = flag.BoolP("url", "u", false, "output the first URL from the note")
	showCmd     = flag.BoolP("command", "c", false, "output the first command from the note")
	showLoc     = flag.BoolP("location", "l", false, "output the note location")
	showBash    = flag.BoolP("bash", "", false, "output bash integration script")

	ErrInvalidNumberOfArgs = errors.New("invalid number of arguments")
)

func run(w io.Writer) error {
	flag.Parse()
	args := flag.Args()

	if *showVersion {
		fmt.Fprintln(w, version)
		return nil
	}

	if *showBash {
		fcqs.WriteBashScript(w)
		return nil
	}

	fileName, err := fcqs.GetNotesFileName()
	if err != nil {
		return fmt.Errorf("cannot get notes file name: %w", err)
	}

	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("cannot access notes file: %w", err)
	}
	defer file.Close()

	if *showURL || *showCmd || *showLoc {
		if len(args) != 1 {
			return ErrInvalidNumberOfArgs
		}

		switch {
		case *showURL:
			fcqs.WriteFirstURL(w, file, args[0])
		case *showCmd:
			fcqs.WriteFirstCmdLineBlock(w, file, args[0])
		case *showLoc:
			fcqs.WriteNoteLocation(w, file, args[0])
		}

		return nil
	}

	switch len(args) {
	case 0:
		fcqs.WriteTitles(w, file)
	case 1:
		fcqs.WriteContents(w, file, args[0])
	default:
		return ErrInvalidNumberOfArgs
	}

	return nil
}

func main() {
	exitCode := 0

	if err := run(os.Stdout); err != nil {
		exitCode = 1
		fmt.Fprintln(os.Stderr, err)
	}

	os.Exit(exitCode)
}
