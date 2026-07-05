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
	"os"

	"github.com/mattn/go-isatty"
	"github.com/mkloubert/go-duplicate-finder/internal/report"
	"github.com/mkloubert/go-duplicate-finder/internal/summaryui"
	"github.com/spf13/cobra"
)

// envClipboard names the clipboard mode when --clipboard is unset.
const envClipboard = "DUPFIND_CLIPBOARD"

func newSummaryCmd() *cobra.Command {
	var (
		reportFile string
		noTUI      bool
		clipMode   string
	)

	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Show a duplicate report in a rich terminal UI",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := applyBoolEnv(cmd, "no-tui", envNoTUI, &noTUI); err != nil {
				return err
			}
			si, err := resolveSI(cmd)
			if err != nil {
				return err
			}
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
			summary := report.FromOutput(out)

			mode := clipMode
			if mode == "" {
				mode = os.Getenv(envClipboard)
			}
			clip, err := summaryui.SelectClipboard(mode, os.Stdout)
			if err != nil {
				return err
			}

			interactive := !noTUI && isatty.IsTerminal(os.Stdout.Fd())
			if interactive {
				in := os.Stdin
				if !stdinIsTTY {
					// STDIN carried the report; read keys from the controlling tty.
					tty, terr := os.Open("/dev/tty")
					if terr != nil {
						interactive = false
					} else {
						defer tty.Close()
						in = tty
					}
				}
				if interactive {
					return summaryui.Run(summary, clip, in)
				}
			}

			summaryui.RenderStatic(os.Stdout, summary, report.SortReclaimable)
			return nil
		},
	}

	cmd.Flags().StringVarP(&reportFile, "report-file", "f", "", "Read the report from this file (env: DUPFIND_REPORT_FILE)")
	cmd.Flags().BoolVar(&noTUI, "no-tui", false, "Force the static renderer (env: DUPFIND_NO_TUI)")
	cmd.Flags().StringVar(&clipMode, "clipboard", "", "Clipboard mode: osc52 (default) or system (env: DUPFIND_CLIPBOARD)")
	return cmd
}
