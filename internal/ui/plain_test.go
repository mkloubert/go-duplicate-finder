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

package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mkloubert/go-duplicate-finder/internal/ui"
)

func TestPlainWritesToWriter(t *testing.T) {
	var buf bytes.Buffer
	var r ui.Reporter = ui.NewPlain(&buf)

	r.ScanStarted()
	r.HashStarted(3)
	r.Errorf("broken: %s", "file.txt")
	r.Done()

	out := buf.String()
	for _, want := range []string{"Scanning", "3", "broken: file.txt", "Done"} {
		if !strings.Contains(out, want) {
			t.Errorf("output does not contain %q:\n%s", want, out)
		}
	}
}

func TestNoopSatisfiesReporter(t *testing.T) {
	var r ui.Reporter = ui.Noop{}
	r.ScanStarted()
	r.FileFound()
	r.HashStarted(1)
	r.HashProgress(1)
	r.Errorf("x")
	r.Done()
}
