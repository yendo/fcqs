# fcqs

fcqs is a quick searcher for flashcards-like notes with fzf.

## Usage

Press `Ctrl+o` (customizable) to launch fcqs on command-line.

You can search for the title of the note with fzf.
The preview screen shows the contents of the note.
The following key bindings are available.

- Enter key: Output the note to standard output.
  If the notes has shell fenced code blocks, the first block is pasted to the command-line.
- Ctrl+y: Copy the note to clip board.
- Ctrl+o: Open the first URL in the note with a browser.
- Ctrl+e: Edit the note

## Installation

Install [fzf](https://github.com/junegunn/fzf) which is required to use fcqs.

Download the fcqs archive from [GitHub Releases](https://github.com/yendo/fcqs/releases) and extract it.

Copy the command `fcqs-cli` to a directory with a path.

``` sh
install fcqs-cli ~/.local/bin/
```

Add the following function for fcqs to `~/.bashrc` for Bash & Unix users.

For Unix standard editor (Vim, Emacs, nano, gedit, etc.):

``` bash
export VISUAL="vim"
eval "$(fcqs-cli --bash)"
```

For Visual Studio Code:

``` bash
export FCQS_EDITOR="vscode"
eval "$(fcqs-cli --bash)"
```

You can customize settings.

``` bash
export FCQS_COPY_KEY="ctrl-y"
export FCQS_OPEN_KEY="ctrl-o"
export FCQS_EDIT_KEY="ctrl-e"
export FCQS_BASH_BIND_KEY="\C-o"
export FCQS_COPY_COMMAND="xclip -selection c"
export FCQS_COPY_WITH_TITLE=true
export FCQS_OPEN_COMMAND="open"
export FCQS_NOTES_FILES="~/fcnotes.md"
```

> [!NOTE]
> `--bash` option is only available in fcqs 0.3.0 or later.
> If you have an older version of fcqs, or want more control,
> you can use [shell.bash](shell.bash).

## Notes specification

### File

The default notes file is `~/fcnotes.md`.

The file can be changed by the environment variable `FCQS_NOTES_FILES`.
The variable can specify multiple files using path separators.
The path separator is `;` on Windows and `:` on other operating systems.

``` bash
export FCQS_NOTES_FILES=~/note.md:/usr/local/doc/note.md
```

### Format

The format of notes is similar to Markdown.
However, all you really need are the titles of each note in the heading
and the content below it.

``` markdown
# title1

contents1

# title2

contents2
```

## Develop

Build the command `fcqs-cli`:

``` sh
make
```

Unit test:

``` sh
make unit-test
```

Integration test:

``` sh
make integration-test
```
