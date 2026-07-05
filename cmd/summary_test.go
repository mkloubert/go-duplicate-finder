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
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSummaryStaticEndToEnd(t *testing.T) {
	dir := t.TempDir()
	rf := filepath.Join(dir, "report.json")
	os.WriteFile(rf, []byte(`{"result":{"/a/first":{"hash":"h","size":100,"duplicates":["/a/second","/a/third"]}}}`), 0o644)

	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	root := newRootCmd()
	root.SetArgs([]string{"summary", "--no-tui", "-f", rf})
	runErr := root.Execute()

	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)

	if runErr != nil {
		t.Fatalf("execute error: %v", runErr)
	}
	for _, want := range []string{"1 groups", "/a/first", "/a/second", "/a/third", "reclaimable"} {
		if !strings.Contains(string(out), want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestSummaryUnknownClipboardMode(t *testing.T) {
	dir := t.TempDir()
	rf := filepath.Join(dir, "report.json")
	os.WriteFile(rf, []byte(`{"result":{}}`), 0o644)

	root := newRootCmd()
	root.SetArgs([]string{"summary", "--no-tui", "-f", rf, "--clipboard", "bogus"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error for unknown clipboard mode")
	}
}
