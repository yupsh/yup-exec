package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestRun(t *testing.T) {
	dir := t.TempDir()

	cases := []struct {
		name       string
		version    string
		args       []string
		stdin      string
		wantOut    string
		wantCode   int
		wantErrSub string
	}{
		{
			name:    "tr uppercases stdin",
			args:    []string{"exec", "tr", "a-z", "A-Z"},
			stdin:   "hello\nworld\n",
			wantOut: "HELLO\nWORLD\n",
		},
		{
			name:    "working directory runs in temp dir",
			args:    []string{"exec", "-C", dir, "pwd"},
			wantOut: dir + "\n",
		},
		{
			name:    "shell flag selects interpreter",
			args:    []string{"exec", "--shell", "sh", "-s", "echo", "shelled"},
			wantOut: "shelled\n",
		},
		{
			name:    "env var is exported",
			args:    []string{"exec", "-s", "-e", "GREETING=hi", "printenv", "GREETING"},
			wantOut: "hi\n",
		},
		{
			name:    "ignore errors suppresses failure",
			args:    []string{"exec", "--ignore-errors", "-q", "false"},
			wantOut: "",
		},
		{
			name:    "version flag reports injected version",
			version: "1.2.3",
			args:    []string{"exec", "--version"},
			wantOut: "exec version 1.2.3\n",
		},
		{
			name:       "no command errors",
			args:       []string{"exec"},
			wantCode:   1,
			wantErrSub: "exec: no command specified",
		},
		{
			name:       "nonexistent command errors",
			args:       []string{"exec", "definitely-not-a-real-cmd-xyz"},
			wantCode:   1,
			wantErrSub: "exec:",
		},
		{
			name:       "unknown flag errors",
			args:       []string{"exec", "--nope"},
			wantCode:   1,
			wantErrSub: "exec:",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			code := run(tc.version, tc.args, strings.NewReader(tc.stdin), &out, &errOut, afero.NewOsFs())

			if code != tc.wantCode {
				t.Fatalf("exit code = %d, want %d (stderr=%q)", code, tc.wantCode, errOut.String())
			}
			if tc.wantErrSub == "" && out.String() != tc.wantOut {
				t.Fatalf("stdout = %q, want %q", out.String(), tc.wantOut)
			}
			if tc.wantErrSub != "" && !strings.Contains(errOut.String(), tc.wantErrSub) {
				t.Fatalf("stderr = %q, want substring %q", errOut.String(), tc.wantErrSub)
			}
		})
	}
}

func Test_main(t *testing.T) {
	origExit, origRun := osExit, runCLI
	t.Cleanup(func() { osExit, runCLI = origExit, origRun })

	gotCode := -1
	osExit = func(code int) { gotCode = code }
	runCLI = func(string, []string, io.Reader, io.Writer, io.Writer, afero.Fs) int { return 7 }

	main()

	if gotCode != 7 {
		t.Fatalf("main propagated exit code %d, want 7", gotCode)
	}
}
