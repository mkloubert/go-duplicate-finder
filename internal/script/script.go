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

package script

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/mkloubert/go-duplicate-finder/internal/report"
)

// Shell is a target shell dialect for the generated script.
type Shell int

const (
	Bash Shell = iota
	Zsh
	PowerShell
)

// Detect chooses a shell from the current environment.
func Detect() Shell {
	return DetectFrom(runtime.GOOS, os.Getenv("SHELL"))
}

// DetectFrom chooses a shell from an OS name and a $SHELL value.
func DetectFrom(goos, shellEnv string) Shell {
	if goos == "windows" {
		return PowerShell
	}
	if strings.Contains(shellEnv, "zsh") {
		return Zsh
	}
	return Bash
}

// ParseShell maps an explicit shell name to a Shell (case-insensitive).
func ParseShell(s string) (Shell, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "bash":
		return Bash, nil
	case "zsh":
		return Zsh, nil
	case "powershell", "pwsh":
		return PowerShell, nil
	default:
		return Bash, fmt.Errorf("unknown shell %q (use bash, zsh, or powershell)", s)
	}
}

// shebang returns the interpreter line for the shell.
func (sh Shell) shebang() string {
	switch sh {
	case Zsh:
		return "#!/usr/bin/env zsh"
	case PowerShell:
		return "#!/usr/bin/env pwsh"
	default:
		return "#!/usr/bin/env bash"
	}
}

// deleteCommand returns the delete command for one path in the shell's dialect.
func (sh Shell) deleteCommand(path string) string {
	if sh == PowerShell {
		return "Remove-Item -LiteralPath " + quotePowerShell(path) + " -Force"
	}
	return "rm -f -- " + quotePOSIX(path)
}

// quotePOSIX single-quotes a string for bash/zsh, escaping embedded quotes.
func quotePOSIX(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// quotePowerShell single-quotes a string for PowerShell, doubling embedded quotes.
func quotePowerShell(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// commentSafe makes s safe to place on a single comment line by replacing the
// line-break characters that would otherwise end the comment and turn the rest
// of the value into an active command line.
func commentSafe(s string) string {
	return strings.NewReplacer("\r", " ", "\n", " ").Replace(s)
}

// Generate writes a delete script to w. Groups are ordered by original path.
// The first occurrence of each group is commented out; the duplicates are
// active delete commands.
func Generate(w io.Writer, s report.Summary, shell Shell) {
	s.SortBy(report.SortPath)

	fmt.Fprintln(w, shell.shebang())
	fmt.Fprintln(w, "# dupfind delete script. Review before running!")
	fmt.Fprintln(w, "# Removes duplicate files, keeping the first occurrence of each group.")

	commands := 0
	for _, g := range s.Groups {
		commands += len(g.Duplicates)
	}
	fmt.Fprintf(w, "# %d delete commands across %d groups (%s reclaimable).\n",
		commands, len(s.Groups), report.Humanize(s.Totals.Reclaimable))

	if len(s.Groups) == 0 {
		fmt.Fprintln(w, "# No duplicates found.")
		return
	}

	for i, g := range s.Groups {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "# Group %d  %s  (%s, %d duplicates)\n",
			i+1, commentSafe(g.Hash), report.Humanize(g.Size), len(g.Duplicates))
		fmt.Fprintf(w, "# keep: %s\n", commentSafe(g.Original))
		fmt.Fprintf(w, "# %s\n", commentSafe(shell.deleteCommand(g.Original)))
		for _, d := range g.Duplicates {
			fmt.Fprintln(w, shell.deleteCommand(d))
		}
	}
}
