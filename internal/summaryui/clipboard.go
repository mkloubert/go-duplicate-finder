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

package summaryui

import (
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/atotto/clipboard"
)

// Clipboard copies text to the user's clipboard.
type Clipboard interface {
	Copy(text string) error
}

// OSC52Clipboard sets the terminal clipboard with an OSC 52 escape sequence.
// It needs no external tools and works over SSH.
type OSC52Clipboard struct {
	W io.Writer
}

// Copy writes the OSC 52 sequence for text to the underlying writer.
func (c OSC52Clipboard) Copy(text string) error {
	enc := base64.StdEncoding.EncodeToString([]byte(text))
	_, err := io.WriteString(c.W, "\x1b]52;c;"+enc+"\a")
	return err
}

// SystemClipboard uses the operating system clipboard (atotto/clipboard).
type SystemClipboard struct{}

// Copy writes text to the OS clipboard.
func (SystemClipboard) Copy(text string) error {
	return clipboard.WriteAll(text)
}

// SelectClipboard returns the clipboard for the given mode. Empty or "osc52"
// selects OSC 52 (writing to out); "system" selects the OS clipboard. Any
// other value is an error.
func SelectClipboard(mode string, out io.Writer) (Clipboard, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "osc52":
		return OSC52Clipboard{W: out}, nil
	case "system":
		return SystemClipboard{}, nil
	default:
		return nil, fmt.Errorf("unknown clipboard mode %q (use \"osc52\" or \"system\")", mode)
	}
}
