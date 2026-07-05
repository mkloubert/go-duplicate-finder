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
	"fmt"
	"os"
	"runtime"

	"github.com/mkloubert/go-duplicate-finder/internal/dedup"
	"github.com/mkloubert/go-duplicate-finder/internal/scanner"
	"github.com/mkloubert/go-duplicate-finder/internal/ui"
	"github.com/spf13/cobra"
)

func newFindCmd() *cobra.Command {
	var (
		output string
		jobs   int
		noTUI  bool
		cwd    string
	)

	cmd := &cobra.Command{
		Use:   "find [patterns...]",
		Short: "Find duplicate files by glob patterns",
		RunE: func(cmd *cobra.Command, args []string) error {
			patterns := args
			if len(patterns) == 0 {
				patterns = []string{"**/**"}
			}

			baseDir, err := resolveBaseDir(cwd)
			if err != nil {
				return err
			}

			rep := ui.New(noTUI)

			files, err := scanner.Scan(baseDir, patterns, rep)
			if err != nil {
				rep.Done()
				return err
			}

			out, err := dedup.Find(files, jobs, rep)
			rep.Done()
			if err != nil {
				return err
			}

			data, err := out.Marshal()
			if err != nil {
				return err
			}

			fmt.Fprintln(os.Stdout, string(data))

			if output != "" {
				if err := os.WriteFile(output, append(data, '\n'), 0o644); err != nil {
					return fmt.Errorf("cannot write output file %q: %w", output, err)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Also write the JSON to this file")
	cmd.Flags().IntVarP(&jobs, "jobs", "j", runtime.NumCPU(), "Number of parallel hash workers")
	cmd.Flags().BoolVar(&noTUI, "no-tui", false, "Disable the rich UI, plain logs only")
	cmd.Flags().StringVar(&cwd, "cwd", "", "Override the working directory (env: DUPFIND_CWD)")
	return cmd
}
