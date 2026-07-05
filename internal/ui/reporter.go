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

// Reporter receives progress and error events from the pipeline.
// All implementations write exclusively to STDERR.
type Reporter interface {
	ScanStarted()
	FileFound()
	HashStarted(total int)
	HashProgress(done int)
	Done()
	Errorf(format string, args ...any)
}

// Noop discards all events (useful in tests).
type Noop struct{}

func (Noop) ScanStarted()                   {}
func (Noop) FileFound()                     {}
func (Noop) HashStarted(total int)          {}
func (Noop) HashProgress(done int)          {}
func (Noop) Done()                          {}
func (Noop) Errorf(format string, a ...any) {}
