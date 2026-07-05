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
	"github.com/mkloubert/go-duplicate-finder/internal/highlight"
	"github.com/spf13/cobra"
)

// envTheme names the syntax-highlight theme when --theme is unset.
const envTheme = "DUPFIND_THEME"

// resolveTheme picks the theme: --theme flag, then DUPFIND_THEME, then monokai.
func resolveTheme(flagTheme, envValue string) string {
	if flagTheme != "" {
		return flagTheme
	}
	if envValue != "" {
		return envValue
	}
	return "monokai"
}

// resolveHighlight reads the persistent --color/--theme flags plus environment
// and the STDOUT TTY state to decide whether to colorize and which theme to use.
func resolveHighlight(cmd *cobra.Command) (bool, string, error) {
	colorStr, _ := cmd.Flags().GetString("color")
	applyStringEnv(cmd, "color", envColor, &colorStr)
	mode, err := highlight.ParseColorMode(colorStr)
	if err != nil {
		return false, "", err
	}
	_, noColor := os.LookupEnv("NO_COLOR")
	enabled := highlight.Enabled(mode, isatty.IsTerminal(os.Stdout.Fd()), noColor)

	flagTheme, _ := cmd.Flags().GetString("theme")
	return enabled, resolveTheme(flagTheme, os.Getenv(envTheme)), nil
}
