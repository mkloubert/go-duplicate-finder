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

func writeScriptReport(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	rf := filepath.Join(dir, "report.json")
	if err := os.WriteFile(rf, []byte(`{"result":{"/a/first":{"hash":"h","size":100,"duplicates":["/a/second","/a/third"]}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	return rf
}

func captureScript(t *testing.T, args ...string) (string, error) {
	t.Helper()
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	root := newRootCmd()
	root.SetArgs(args)
	err := root.Execute()
	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	return string(out), err
}

func TestScriptBashEndToEnd(t *testing.T) {
	rf := writeScriptReport(t)
	out, err := captureScript(t, "script", "--shell", "bash", "-f", rf)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if !strings.HasPrefix(out, "#!/usr/bin/env bash") {
		t.Errorf("expected bash shebang, got:\n%s", out)
	}
	if !strings.Contains(out, "# rm -f -- '/a/first'") {
		t.Errorf("original should be commented:\n%s", out)
	}
	for _, want := range []string{"rm -f -- '/a/second'", "rm -f -- '/a/third'"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing active delete %q:\n%s", want, out)
		}
	}
}

func TestScriptPowerShellEndToEnd(t *testing.T) {
	rf := writeScriptReport(t)
	out, err := captureScript(t, "script", "--shell", "powershell", "-f", rf)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if !strings.Contains(out, "Remove-Item -LiteralPath '/a/second' -Force") {
		t.Errorf("missing powershell delete:\n%s", out)
	}
}

func TestScriptUnknownShell(t *testing.T) {
	rf := writeScriptReport(t)
	if _, err := captureScript(t, "script", "--shell", "fish", "-f", rf); err == nil {
		t.Fatal("expected error for unknown shell")
	}
}
