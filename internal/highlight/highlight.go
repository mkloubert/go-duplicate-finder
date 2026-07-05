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

package highlight

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// ColorMode controls when output is colorized.
type ColorMode int

const (
	Auto ColorMode = iota
	Always
	Never
)

// ParseColorMode maps a --color value to a ColorMode. Empty means Auto.
func ParseColorMode(s string) (ColorMode, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "auto":
		return Auto, nil
	case "always":
		return Always, nil
	case "never":
		return Never, nil
	default:
		return Auto, fmt.Errorf("invalid color mode %q (use auto, always, or never)", s)
	}
}

// Enabled decides whether to colorize output. Always overrides everything;
// Never disables; Auto colorizes only on a TTY without NO_COLOR set.
func Enabled(mode ColorMode, isTTY, noColorSet bool) bool {
	switch mode {
	case Always:
		return true
	case Never:
		return false
	default:
		return isTTY && !noColorSet
	}
}

// Write emits code to w. When enabled is false the raw code is written verbatim.
// When enabled is true it highlights with Chroma into a buffer first and writes
// the result only on success; on any tokenise/format error it falls back to the
// raw code, so highlighting can never corrupt or truncate the output.
func Write(w io.Writer, code, language string, enabled bool, theme string) error {
	if !enabled {
		_, err := io.WriteString(w, code)
		return err
	}

	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get(theme)
	if style == nil {
		style = styles.Fallback
	}

	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	var buf bytes.Buffer
	iterator, err := lexer.Tokenise(nil, code)
	if err == nil {
		err = formatter.Format(&buf, style, iterator)
	}
	if err != nil {
		_, werr := io.WriteString(w, code)
		return werr
	}
	_, werr := io.WriteString(w, buf.String())
	return werr
}
