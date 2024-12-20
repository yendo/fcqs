# title

contents

# Long title and contents have lines

line 1

line 2

# Regular expression meta chars in the title are ignored $

contents

# Consecutive blank lines are combined into a single line


line 1


line 2


# same title

Contents with the same title are combined into one.

# same title

2nd

## Heading levels and structures are ignored

contents

# Trailing spaces in the title are ignored  

The contents have trailing spaces.  

# same title

3rd

# Notes without content output the title only

#   Spaces before the title are ignored

contents

# Headings in fenced code blocks are ignored

```
# fenced heading
```

# There can be no blank line
contents
#

no title contents are not output.

#  

title is only spaces

# Titles without a space after the # are not recognized

#no_space_title

contents

  # Titles with spaces before the # are not recognized

contents

# URL

fcqs: http://github.com/yendo/fcqs/
github: http://github.com/

# command-line

```sh
ls -l | nl
```

# command-line with $

```console
$ date
```

