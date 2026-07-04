// Command yup-exec is the CLI wrapper around github.com/gloo-foo/cmd-exec.
package main

import (
	clix "github.com/gloo-foo/cli"
	command "github.com/gloo-foo/cmd-exec"
	urf "github.com/urfave/cli/v3"
)

// version is the build version. It defaults to "dev" for local builds and is
// overridden at release time via the linker: -ldflags "-X main.version=<v>".
var version = "dev"

const (
	name             = "exec"
	flagWorkingDir   = "directory"
	flagEnvVar       = "env"
	flagShell        = "shell"
	flagUseShell     = "use-shell"
	flagIgnoreErrors = "ignore-errors"
	flagQuiet        = "quiet"
)

// Error is the package sentinel type; every error the wrapper emits is a const
// of this type, making each path testable with errors.Is.
type Error string

func (e Error) Error() string { return string(e) }

// ErrNoCommand is emitted when no command operand is supplied.
const ErrNoCommand Error = "no command specified"

// synopsis is the multi-line --help usage block; urfave/cli indents it three
// spaces, so the lines stay flush-left.
const synopsis = `exec [OPTIONS] COMMAND [ARG...]

Execute COMMAND with the given arguments, piping standard input
to its stdin and its stdout to standard output.`

// spec declares the exec wrapper: a stdin filter whose operands are the program
// and its arguments, configured by the option flags.
var spec = clix.Spec{
	Name:     name,
	Summary:  "execute external commands",
	Synopsis: synopsis,
	Build:    build,
	Flags:    flags(),
}

// flags returns a fresh set of the wrapper's flags. Each call yields new flag
// values, so parsing one invocation never leaks urfave/cli's per-flag "was set"
// state into another (which IsSet reads).
func flags() []urf.Flag {
	return []urf.Flag{
		&urf.StringFlag{Name: flagWorkingDir, Aliases: []string{"C"}, Usage: "run command in DIRECTORY"},
		&urf.StringSliceFlag{Name: flagEnvVar, Aliases: []string{"e"}, Usage: "set environment variable (NAME=VALUE)"},
		&urf.StringFlag{Name: flagShell, Usage: "shell to use for execution"},
		&urf.BoolFlag{Name: flagUseShell, Aliases: []string{"s"}, Usage: "execute command through a shell"},
		&urf.BoolFlag{Name: flagIgnoreErrors, Usage: "succeed even if the command exits non-zero"},
		&urf.BoolFlag{Name: flagQuiet, Aliases: []string{"q"}, Usage: "discard the command's stderr"},
	}
}

// build maps the invocation to exec's pipeline: standard input feeds the
// command, whose program and arguments are the operands, configured by the
// flags. A bare invocation with no command is a usage error.
func build(inv clix.Invocation) (clix.Source, clix.Command, error) {
	operands := inv.Args.Args().Slice()
	if len(operands) == 0 {
		return nil, nil, ErrNoCommand
	}
	opts := options(inv.Args)
	params := make([]any, 0, len(operands)+len(opts))
	for _, o := range operands {
		params = append(params, o)
	}
	params = append(params, opts...)
	return clix.Stdin(inv.Stdin), command.Exec(params...), nil
}

// options folds the parsed string and slice flags into exec's option values.
func options(c *urf.Command) []any {
	var opts []any
	if c.IsSet(flagWorkingDir) {
		opts = append(opts, command.ExecWorkingDir(c.String(flagWorkingDir)))
	}
	if c.IsSet(flagShell) {
		opts = append(opts, command.ExecShell(c.String(flagShell)))
	}
	for _, env := range c.StringSlice(flagEnvVar) {
		opts = append(opts, command.ExecEnvVar(env))
	}
	return append(opts, boolOptions(c)...)
}

// boolOptions folds the parsed boolean flags into exec's option values.
func boolOptions(c *urf.Command) []any {
	var opts []any
	if c.Bool(flagUseShell) {
		opts = append(opts, command.ExecUseShell)
	}
	if c.Bool(flagIgnoreErrors) {
		opts = append(opts, command.ExecIgnoreErrors)
	}
	if c.Bool(flagQuiet) {
		opts = append(opts, command.ExecQuiet)
	}
	return opts
}

// runMain is an indirection seam so main's wiring is testable without spawning
// the process; a test swaps it and restores it.
var runMain = clix.Main

func main() { runMain(spec, version) }
