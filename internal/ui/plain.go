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

package ui

import (
	"fmt"
	"io"
	"os"
)

// Plain writes simple status lines (no TTY / --no-tui).
type Plain struct {
	w io.Writer
}

// NewPlain creates a Plain reporter. w=nil ⇒ STDERR.
func NewPlain(w io.Writer) *Plain {
	if w == nil {
		w = os.Stderr
	}
	return &Plain{w: w}
}

func (p *Plain) ScanStarted() { fmt.Fprintln(p.w, "Scanning files …") }
func (p *Plain) FileFound()   {}
func (p *Plain) HashStarted(total int) {
	fmt.Fprintf(p.w, "Hashing %d candidates …\n", total)
}
func (p *Plain) HashProgress(done int) {}
func (p *Plain) Done()                 { fmt.Fprintln(p.w, "Done.") }
func (p *Plain) Errorf(format string, a ...any) {
	fmt.Fprintf(p.w, "Error: "+format+"\n", a...)
}
