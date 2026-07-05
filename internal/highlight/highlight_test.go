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

package highlight_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mkloubert/go-duplicate-finder/internal/highlight"
)

func TestParseColorMode(t *testing.T) {
	valid := map[string]highlight.ColorMode{
		"":       highlight.Auto,
		"auto":   highlight.Auto,
		"Always": highlight.Always,
		"NEVER":  highlight.Never,
	}
	for in, want := range valid {
		got, err := highlight.ParseColorMode(in)
		if err != nil || got != want {
			t.Errorf("ParseColorMode(%q) = %v,%v want %v", in, got, err, want)
		}
	}
	if _, err := highlight.ParseColorMode("rainbow"); err == nil {
		t.Error("expected error for invalid mode")
	}
}

func TestEnabled(t *testing.T) {
	cases := []struct {
		mode           highlight.ColorMode
		isTTY, noColor bool
		want           bool
	}{
		{highlight.Never, true, false, false},
		{highlight.Always, false, true, true}, // always overrides NO_COLOR and non-TTY
		{highlight.Auto, true, false, true},
		{highlight.Auto, false, false, false}, // pipe -> no color (pipe safety)
		{highlight.Auto, true, true, false},   // NO_COLOR -> no color
	}
	for _, c := range cases {
		if got := highlight.Enabled(c.mode, c.isTTY, c.noColor); got != c.want {
			t.Errorf("Enabled(%v,%v,%v) = %v, want %v", c.mode, c.isTTY, c.noColor, got, c.want)
		}
	}
}

func TestWriteRawWhenDisabled(t *testing.T) {
	var buf bytes.Buffer
	code := `{"a":1}`
	if err := highlight.Write(&buf, code, "json", false, "monokai"); err != nil {
		t.Fatal(err)
	}
	if buf.String() != code {
		t.Errorf("disabled output = %q, want raw %q", buf.String(), code)
	}
	if strings.Contains(buf.String(), "\x1b[") {
		t.Error("disabled output must not contain ANSI escapes")
	}
}

func TestWriteHighlightsWhenEnabled(t *testing.T) {
	var buf bytes.Buffer
	code := `{"hello":"world"}`
	if err := highlight.Write(&buf, code, "json", true, "monokai"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "\x1b[") {
		t.Errorf("enabled output should contain ANSI escapes:\n%q", out)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("highlighted output should still contain the source text:\n%q", out)
	}
}

func TestWriteUnknownLanguageStillWorks(t *testing.T) {
	var buf bytes.Buffer
	if err := highlight.Write(&buf, "some text", "no-such-lang", true, "monokai"); err != nil {
		t.Fatalf("unknown language should not error: %v", err)
	}
	if !strings.Contains(buf.String(), "some text") {
		t.Errorf("output should contain the text:\n%q", buf.String())
	}
}
