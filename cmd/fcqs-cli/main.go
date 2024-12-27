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
	noTitle     = flag.BoolP("notitle", "t", false, "no title on output content")

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

	fileName, err := fcqs.NotesFileName()
	if err != nil {
		return fmt.Errorf("notes file name: %w", err)
	}

	var files []*os.File
	for _, v := range fileName {
		file, err := os.Open(v)
		if err != nil {
			return fmt.Errorf("notes file: %w", err)
		}
		files = append(files, file)
		defer file.Close()
	}

	file := io.MultiReader(files[0])

	if *showURL || *showCmd || *showLoc {
		if len(args) != 1 {
			return ErrInvalidNumberOfArgs
		}

		switch {
		case *showURL:
			err = fcqs.WriteFirstURL(w, file, args[0])
		case *showCmd:
			err = fcqs.WriteFirstCmdLineBlock(w, file, args[0])
		case *showLoc:
			err = fcqs.WriteNoteLocation(w, files, args[0])
		}

		return err
	}

	switch len(args) {
	case 0:
		err = fcqs.WriteTitles(w, file)
	case 1:
		err = fcqs.WriteContents(w, file, args[0], *noTitle)
	default:
		return ErrInvalidNumberOfArgs
	}

	return err
}

func main() {
	exitCode := 0

	if err := run(os.Stdout); err != nil {
		exitCode = 1
		fmt.Fprintln(os.Stderr, err)
	}

	os.Exit(exitCode)
}
