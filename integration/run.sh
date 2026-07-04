#!/bin/sh
# Integration checks for yup-exec, run inside a Debian (GNU coreutils) container.
#
# exec is a gloo-specific / structural command: it runs an external program as a
# pipeline stage (stdin -> program stdin, program stdout -> stdout). There is NO
# GNU coreutils equivalent to compare against, so this harness is ASSERT-ONLY: it
# exercises exec's own documented contract (its flags, inputs, outputs) using
# deterministic target programs (echo, true, false, tr, pwd, printenv, sh).
#
# Note on `--`: yup-exec parses leading dash-prefixed tokens as ITS OWN flags
# (urfave/cli). To pass dash-flags through to the target program, separate them
# with `--` (e.g. `yup-exec -s -- sh -c '...'`). The cases below use `--`
# wherever the target itself takes dash arguments.
#
# stdin holds the data piped to the target on its stdin; the want value is
# compared against yup-exec's stdout. assert_code checks the exit status only.
set -eu

fails=0
stdin=''

# assert WANT ARGS... : run `yup-exec ARGS` (with $stdin on stdin) and require
#                       stdout to equal WANT exactly.
assert() {
  want=$1
  shift
  got=$(printf '%s' "$stdin" | yup-exec "$@" 2>/dev/null || true)
  if [ "$got" = "$want" ]; then
    printf 'ok    assert  exec %s\n' "$*"
  else
    printf 'FAIL  assert  exec %s\n        want: %s\n        got:  %s\n' "$*" "$want" "$got"
    fails=$((fails + 1))
  fi
}

# assert_code WANT ARGS... : require the exit status of `yup-exec ARGS` (with
#                            $stdin on stdin) to equal WANT.
assert_code() {
  want=$1
  shift
  # `|| got=$?` keeps `set -e` from aborting on the intentional non-zero exits.
  got=0
  printf '%s' "$stdin" | yup-exec "$@" >/dev/null 2>&1 || got=$?
  if [ "$got" = "$want" ]; then
    printf 'ok    code    exec %s (exit %s)\n' "$*" "$got"
  else
    printf 'FAIL  code    exec %s\n        want exit: %s\n        got exit:  %s\n' "$*" "$want" "$got"
    fails=$((fails + 1))
  fi
}

# Direct execution: the program's stdout becomes exec's stdout. `--` lets the
# target's own -n flag through instead of yup-exec consuming it.
assert 'hello' -- echo -n hello

# stdin is piped to the program's stdin; tr's stdout is exec's stdout.
stdin='hello'
assert 'HELLO' tr a-z A-Z
stdin=''

# --directory (-C): run the program in DIRECTORY.
assert '/tmp' -C /tmp pwd

# --env (-e) with --use-shell (-s): the variable is exported to the program.
assert 'hi' -s -e GREETING=hi printenv GREETING

# --env (-e) repeats: multiple variables are all exported.
assert 'a-b' -s -e A=a -e B=b -- sh -c 'printf "%s-%s" "$A" "$B"'

# --shell selects the interpreter used by --use-shell.
assert 'shelled' --shell sh -s echo shelled

# Exit-status contract: a successful program yields exit 0.
assert_code 0 true

# A failing program propagates a non-zero exit status (exec maps it to 1).
assert_code 1 false

# --ignore-errors makes a failing program succeed.
assert_code 0 --ignore-errors false

# --quiet (-q) discards the program's stderr; stdout still flows.
assert 'out' -s -q -- sh -c 'echo discarded >&2; printf out'

# Missing operand: no command specified -> exit 1.
assert_code 1

# A nonexistent program fails -> exit 1.
assert_code 1 definitely-not-a-real-cmd-xyz

if [ "$fails" -ne 0 ]; then
  printf '\n%s check(s) failed\n' "$fails"
  exit 1
fi
printf '\nall checks passed\n'
