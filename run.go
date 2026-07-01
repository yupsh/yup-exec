package main

import (
	"context"
	"fmt"
	"io"

	command "github.com/gloo-foo/cmd-exec"
	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

const name = "exec"

const (
	flagWorkingDir   = "directory"
	flagEnvVar       = "env"
	flagShell        = "shell"
	flagUseShell     = "use-shell"
	flagIgnoreErrors = "ignore-errors"
	flagQuiet        = "quiet"
)

// usageText is the command's multi-line usage synopsis, shown in --help.
// cli/v3 indents the whole block by 3 spaces, so these lines are flush-left to
// stay aligned in the rendered output.
const usageText = `exec [OPTIONS] COMMAND [ARG...]

Execute COMMAND with the given arguments, piping standard input
to its stdin and its stdout to standard output.`

// init replaces urfave/cli's default --version/-v flag with a --version-only
// flag, freeing the single-letter -v for command flags while still exposing
// the injected build version.
func init() {
	cli.VersionFlag = &cli.BoolFlag{Name: "version", Usage: "print version information and exit"}
}

// Error is the sentinel error type for this package.
type Error string

func (e Error) Error() string { return string(e) }

// ErrNoCommand is emitted when no command operand is supplied.
const ErrNoCommand Error = "no command specified"

// run builds and executes the exec CLI against the injected version and I/O,
// returning the process exit code. The filesystem is accepted for parity with
// sibling wrappers; exec sources its input solely from stdin.
func run(version string, args []string, stdin io.Reader, stdout, stderr io.Writer, _ afero.Fs) int {
	cmd := newApp(version, stdin, stdout)
	cmd.Writer = stdout
	cmd.ErrWriter = stderr
	if err := cmd.Run(context.Background(), args); err != nil {
		_, _ = fmt.Fprintf(stderr, name+": %v\n", err)
		return 1
	}
	return 0
}

func newApp(version string, stdin io.Reader, stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:            name,
		Version:         version,
		Usage:           "execute external commands",
		UsageText:       usageText,
		HideHelpCommand: true,
		// Keep exit handling in run() rather than letting urfave/cli call
		// os.Exit, so the exit code stays testable.
		ExitErrHandler: func(context.Context, *cli.Command, error) {},
		Flags:          appFlags(),
		Action:         action(stdin, stdout),
	}
}

func appFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: flagWorkingDir, Aliases: []string{"C"}, Usage: "run command in DIRECTORY"},
		&cli.StringSliceFlag{Name: flagEnvVar, Aliases: []string{"e"}, Usage: "set environment variable (NAME=VALUE)"},
		&cli.StringFlag{Name: flagShell, Usage: "shell to use for execution"},
		&cli.BoolFlag{Name: flagUseShell, Aliases: []string{"s"}, Usage: "execute command through a shell"},
		&cli.BoolFlag{Name: flagIgnoreErrors, Usage: "succeed even if the command exits non-zero"},
		&cli.BoolFlag{Name: flagQuiet, Aliases: []string{"q"}, Usage: "discard the command's stderr"},
	}
}

func action(stdin io.Reader, stdout io.Writer) cli.ActionFunc {
	return func(_ context.Context, c *cli.Command) error {
		params, err := params(c)
		if err != nil {
			return err
		}
		_, err = gloo.Run(gloo.ByteReaderSource([]io.Reader{stdin}), gloo.ByteWriteTo(stdout), command.Exec(params...))
		return err
	}
}

func params(c *cli.Command) ([]any, error) {
	operand := positional(c)
	if len(operand) == 0 {
		return nil, ErrNoCommand
	}
	return append(operand, options(c)...), nil
}

func positional(c *cli.Command) []any {
	args := make([]any, c.NArg())
	for i := range args {
		args[i] = c.Args().Get(i)
	}
	return args
}

func options(c *cli.Command) []any {
	var opts []any
	opts = appendString(opts, c, flagWorkingDir, func(s string) any { return command.ExecWorkingDir(s) })
	opts = appendString(opts, c, flagShell, func(s string) any { return command.ExecShell(s) })
	for _, env := range c.StringSlice(flagEnvVar) {
		opts = append(opts, command.ExecEnvVar(env))
	}
	return append(opts, boolOptions(c)...)
}

func appendString(opts []any, c *cli.Command, name string, build func(string) any) []any {
	if c.IsSet(name) {
		return append(opts, build(c.String(name)))
	}
	return opts
}

func boolOptions(c *cli.Command) []any {
	flags := []struct {
		opt  any
		name string
	}{
		{name: flagUseShell, opt: command.ExecUseShell},
		{name: flagIgnoreErrors, opt: command.ExecIgnoreErrors},
		{name: flagQuiet, opt: command.ExecQuiet},
	}
	var opts []any
	for _, f := range flags {
		if c.Bool(f.name) {
			opts = append(opts, f.opt)
		}
	}
	return opts
}
