# yup-exec

```
NAME:
   exec - execute external commands

USAGE:
   exec [OPTIONS] COMMAND [ARG...]

   Execute COMMAND with the given arguments, piping standard input
   to its stdin and its stdout to standard output.

VERSION:
   dev

GLOBAL OPTIONS:
   --directory string, -C string                        run command in DIRECTORY
   --env string, -e string [ --env string, -e string ]  set environment variable (NAME=VALUE)
   --shell string                                       shell to use for execution
   --use-shell, -s                                      execute command through a shell
   --ignore-errors                                      succeed even if the command exits non-zero
   --quiet, -q                                          discard the command's stderr
   --help, -h                                           show help
   --version                                            print version information and exit
```
