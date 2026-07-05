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
	"bytes"
	"encoding/base64"
	"testing"
)

func TestOSC52CopyWritesEscape(t *testing.T) {
	var buf bytes.Buffer
	c := OSC52Clipboard{W: &buf}
	if err := c.Copy("hello"); err != nil {
		t.Fatal(err)
	}
	want := "\x1b]52;c;" + base64.StdEncoding.EncodeToString([]byte("hello")) + "\a"
	if buf.String() != want {
		t.Fatalf("got %q, want %q", buf.String(), want)
	}
}

func TestSelectClipboard(t *testing.T) {
	var buf bytes.Buffer

	c, err := SelectClipboard("", &buf)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := c.(OSC52Clipboard); !ok {
		t.Errorf(`"" -> %T, want OSC52Clipboard`, c)
	}

	c, err = SelectClipboard("osc52", &buf)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := c.(OSC52Clipboard); !ok {
		t.Errorf(`"osc52" -> %T, want OSC52Clipboard`, c)
	}

	c, err = SelectClipboard("SYSTEM", &buf)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := c.(SystemClipboard); !ok {
		t.Errorf(`"system" -> %T, want SystemClipboard`, c)
	}

	if _, err := SelectClipboard("bogus", &buf); err == nil {
		t.Error("expected error for unknown mode")
	}
}
