package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/yendo/fcqs"
)

var (
	version = "unknown"

	showVersion = flag.Bool("v", false, "output the version")
	showURL     = flag.Bool("u", false, "output the first URL from the note")
	showCmd     = flag.Bool("c", false, "output the first command from the note")
	showLoc     = flag.Bool("l", false, "output the note location")

	ErrInvalidNumberOfArgs = errors.New("invalid number of arguments")
)

func run(w io.Writer) error {
	flag.Parse()
	args := flag.Args()

	if *showVersion {
		fmt.Fprintln(w, version)
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
			fcqs.WriteFirstCmdLine(w, file, args[0])
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
