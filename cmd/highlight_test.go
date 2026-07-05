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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeDupTree(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range []string{"a.txt", "b.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("dup"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestFindColorAlwaysHighlightsButFileStaysRaw(t *testing.T) {
	dir := makeDupTree(t)
	outFile := filepath.Join(dir, "out.json")
	out, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--color", "always", "-o", outFile, "**/**")
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if !strings.Contains(out, "\x1b[") {
		t.Errorf("--color always should highlight stdout:\n%q", out)
	}
	fileRaw, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(fileRaw), "\x1b[") {
		t.Errorf("--output file must be raw (no ANSI):\n%s", fileRaw)
	}
	if !strings.Contains(string(fileRaw), `"duplicates"`) {
		t.Errorf("--output file should be valid raw JSON:\n%s", fileRaw)
	}
}

func TestFindDefaultPipedIsRaw(t *testing.T) {
	dir := makeDupTree(t)
	out, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "**/**")
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if strings.Contains(out, "\x1b[") {
		t.Errorf("piped (non-TTY) stdout must be raw JSON:\n%q", out)
	}
}

func TestScriptColorAlwaysHighlights(t *testing.T) {
	rf := writeScriptReport(t)
	out, err := captureScript(t, "script", "-f", rf, "--shell", "bash", "--color", "always")
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if !strings.Contains(out, "\x1b[") {
		t.Errorf("script --color always should highlight:\n%q", out)
	}
}

func TestScriptDefaultPipedIsRaw(t *testing.T) {
	rf := writeScriptReport(t)
	out, err := captureScript(t, "script", "-f", rf, "--shell", "bash")
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if strings.Contains(out, "\x1b[") {
		t.Errorf("piped script must be raw:\n%q", out)
	}
	if !strings.Contains(out, "rm -f -- '/a/second'") {
		t.Errorf("raw script content missing:\n%s", out)
	}
}

func TestInvalidColorMode(t *testing.T) {
	rf := writeScriptReport(t)
	if _, err := captureScript(t, "script", "-f", rf, "--color", "rainbow"); err == nil {
		t.Fatal("expected error for invalid --color value")
	}
}

func TestScriptColorNeverIsRaw(t *testing.T) {
	rf := writeScriptReport(t)
	out, err := captureScript(t, "script", "-f", rf, "--shell", "bash", "--color", "never")
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if strings.Contains(out, "\x1b[") {
		t.Errorf("--color never must never emit ANSI:\n%q", out)
	}
}

func TestScriptNoColorEnvDisablesUnderAlwaysOverride(t *testing.T) {
	// NO_COLOR set: under auto (piped, non-TTY) output stays raw; and --color
	// never stays raw regardless. Guards that NO_COLOR never forces color on.
	t.Setenv("NO_COLOR", "1")
	rf := writeScriptReport(t)
	out, err := captureScript(t, "script", "-f", rf, "--shell", "bash")
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if strings.Contains(out, "\x1b[") {
		t.Errorf("NO_COLOR piped output must be raw:\n%q", out)
	}
}
