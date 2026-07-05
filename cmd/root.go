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
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "dupfind",
		Short:         "dupfind finds duplicate files",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().String("color", "auto", "Colorize output: auto, always, or never (env: DUPFIND_COLOR)")
	root.PersistentFlags().String("theme", "", "Syntax highlight theme (env: DUPFIND_THEME; default monokai)")
	root.PersistentFlags().Bool("si", false, "Use 1000-based SI size units (kB, MB) instead of 1024-based (env: DUPFIND_SI)")
	root.AddCommand(newFindCmd())
	root.AddCommand(newSummaryCmd())
	root.AddCommand(newScriptCmd())
	return root
}

// Execute is the entry point of the CLI.
func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		// --fail-if-duplicates is a result signal, not an error: exit 2 quietly.
		if errors.Is(err, errDuplicatesFound) {
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
