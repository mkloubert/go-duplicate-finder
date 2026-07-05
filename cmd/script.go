// Copyright © 2026 Marcel Joachim Kloubert <marcel@kloubert.dev>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package cmd

import (
	"bytes"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/mkloubert/go-duplicate-finder/internal/highlight"
	"github.com/mkloubert/go-duplicate-finder/internal/report"
	"github.com/mkloubert/go-duplicate-finder/internal/script"
	"github.com/spf13/cobra"
)

// envShell names the target shell when --shell is unset.
const envShell = "DUPFIND_SHELL"

// resolveShell picks the target shell: flag over env over auto-detect.
func resolveShell(flag, env string) (script.Shell, error) {
	value := flag
	if value == "" {
		value = env
	}
	if value == "" || strings.EqualFold(value, "auto") {
		return script.Detect(), nil
	}
	return script.ParseShell(value)
}

func newScriptCmd() *cobra.Command {
	var (
		reportFile string
		shellName  string
	)

	cmd := &cobra.Command{
		Use:   "script",
		Short: "Generate a shell script that deletes the duplicates",
		RunE: func(cmd *cobra.Command, args []string) error {
			si, _ := cmd.Flags().GetBool("si")
			report.SetSIUnits(si)

			stdinIsTTY := isatty.IsTerminal(os.Stdin.Fd())

			data, err := resolveReportInput(reportFile, os.Stdin, stdinIsTTY)
			if err != nil {
				return err
			}
			out, err := parseReport(data)
			if err != nil {
				return err
			}

			sh, err := resolveShell(shellName, os.Getenv(envShell))
			if err != nil {
				return err
			}

			enabled, theme, err := resolveHighlight(cmd)
			if err != nil {
				return err
			}

			var buf bytes.Buffer
			script.Generate(&buf, report.FromOutput(out), sh)
			return highlight.Write(os.Stdout, buf.String(), shellLang(sh), enabled, theme)
		},
	}

	cmd.Flags().StringVarP(&reportFile, "report-file", "f", "", "Read the report from this file (env: DUPFIND_REPORT_FILE)")
	cmd.Flags().StringVar(&shellName, "shell", "", "Target shell: auto (default), bash, zsh, or powershell (env: DUPFIND_SHELL)")
	return cmd
}

// shellLang maps a shell to its Chroma lexer name.
func shellLang(sh script.Shell) string {
	if sh == script.PowerShell {
		return "powershell"
	}
	return "bash"
}
