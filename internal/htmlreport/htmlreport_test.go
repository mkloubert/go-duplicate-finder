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

package htmlreport_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mkloubert/go-duplicate-finder/internal/htmlreport"
	"github.com/mkloubert/go-duplicate-finder/internal/model"
	"github.com/mkloubert/go-duplicate-finder/internal/report"
)

func TestWriteContent(t *testing.T) {
	o := model.New()
	o.Result["/photos/a.jpg"] = &model.FileResult{Hash: "h", Size: 1048576, Duplicates: []string{"/backup/a.jpg"}}
	s := report.FromOutput(o)

	var buf bytes.Buffer
	if err := htmlreport.Write(&buf, s); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	for _, want := range []string{
		"<!DOCTYPE html>", "dupfind report",
		"/photos/a.jpg", "/backup/a.jpg", "1.0 MB",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("HTML missing %q:\n%s", want, out)
		}
	}
}

func TestWriteEscapesPaths(t *testing.T) {
	o := model.New()
	o.Result["/x/<script>alert(1)</script>.txt"] = &model.FileResult{
		Hash: "h", Size: 5, Duplicates: []string{"/y/a.txt"},
	}
	s := report.FromOutput(o)

	var buf bytes.Buffer
	if err := htmlreport.Write(&buf, s); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	if strings.Contains(out, "<script>alert(1)</script>") {
		t.Errorf("path was not escaped — XSS risk:\n%s", out)
	}
	if !strings.Contains(out, "&lt;script&gt;") {
		t.Errorf("expected an HTML-escaped path:\n%s", out)
	}
}

func TestWriteEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := htmlreport.Write(&buf, report.FromOutput(model.New())); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No duplicates found.") {
		t.Errorf("empty report should say no duplicates:\n%s", buf.String())
	}
}
