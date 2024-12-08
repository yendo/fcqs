# fcqs

# You can customize settings:
#
# FCQS_EDITOR="default" or "vscode"
# FCQS_COPY_KEY="ctrl-y"
# FCQS_OPEN_KEY="ctrl-o"
# FCQS_EDIT_KEY="ctrl-e"
# FCQS_BASH_BIND_KEY="\C-o"

EDITOR=${FCQS_EDITOR:-default}
COPY_KEY=${FCQS_COPY_KEY:-ctrl-y}
OPEN_KEY=${FCQS_OPEN_KEY:-ctrl-o}
EDIT_KEY=${FCQS_EDIT_KEY:-ctrl-e}
BASH_BIND_KEY=${FCQS_BASH_BIND_KEY:-"\C-o"}

EDIT_COMMAND_DEFAULT="awk '{printf \"+%s %s\n\",\$2,\$1}' | xargs -o ${VISUAL} > /dev/tty"
EDIT_COMMAND_VSCODE="awk '{printf \"%s:%s\n\",\$1,\$2}' | xargs -o code -g"
[ "${EDITOR}" = "vscode" ] && EDIT_COMMAND=${EDIT_COMMAND_VSCODE} || EDIT_COMMAND=${EDIT_COMMAND_DEFAULT}

fcqs() {
  local title
  title=$(fcqs-cli |
    fzf --preview "fcqs-cli {}" \
      --bind "${COPY_KEY}:execute-silent(fcqs-cli {} | xclip -selection c),${OPEN_KEY}:execute-silent(fcqs-cli -u {} | xargs xdg-open),${EDIT_KEY}:execute-silent(fcqs-cli -l {} | ${EDIT_COMMAND})+abort")
  fcqs-cli "$title"
  local command
  command=$(fcqs-cli -c "$title")
  READLINE_LINE="${READLINE_LINE:0:$READLINE_POINT}${command}${READLINE_LINE:$READLINE_POINT}"
  READLINE_POINT=$((READLINE_POINT + ${#command}))
}

eval "bind -x '\"${BASH_BIND_KEY}\":fcqs'"
