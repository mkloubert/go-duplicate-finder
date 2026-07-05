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
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindFormatHTML(t *testing.T) {
	dir := makeSizedDupTree(t)
	out, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--format", "html", "**/**")
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	for _, want := range []string{"<!DOCTYPE html>", "dupfind report", "a1.bin"} {
		if !strings.Contains(out, want) {
			t.Errorf("HTML output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "\x1b[") {
		t.Errorf("HTML output must not contain ANSI escapes:\n%s", out)
	}
}

func TestFindFormatHTMLToFile(t *testing.T) {
	dir := makeSizedDupTree(t)
	outFile := filepath.Join(dir, "report.html")
	if _, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--format", "html", "-o", outFile, "**/**"); err != nil {
		t.Fatalf("execute error: %v", err)
	}
	raw, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "<!DOCTYPE html>") {
		t.Errorf("HTML file missing doctype:\n%s", raw)
	}
}

func TestFindInvalidFormat(t *testing.T) {
	dir := makeSizedDupTree(t)
	if _, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--format", "xml", "**/**"); err == nil {
		t.Fatal("expected error for invalid --format")
	}
}

func TestFindFailIfDuplicates(t *testing.T) {
	dir := makeSizedDupTree(t) // has duplicates
	out, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--compact", "--fail-if-duplicates", "**/**")
	if !errors.Is(err, errDuplicatesFound) {
		t.Fatalf("expected errDuplicatesFound, got %v", err)
	}
	// the report is still written to STDOUT before the signal.
	if !strings.Contains(out, `"duplicates"`) {
		t.Errorf("report should still be printed:\n%s", out)
	}
}

func TestFindFailIfDuplicatesNone(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "only.txt"), []byte("unique"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--compact", "--fail-if-duplicates", "**/**"); err != nil {
		t.Fatalf("no duplicates should not fail, got %v", err)
	}
}

func TestFindFailIfDuplicatesRespectsThreshold(t *testing.T) {
	dir := makeSizedDupTree(t) // groups reclaim 100 and 10 bytes
	// min-reclaimable above both filters everything out -> no failure.
	if _, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--compact", "--fail-if-duplicates", "--min-reclaimable", "1G", "**/**"); err != nil {
		t.Fatalf("threshold filtered all groups, should not fail, got %v", err)
	}
}

func TestFindMinCount(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"t1", "t2", "t3"} { // group with 3 files
		if err := os.WriteFile(filepath.Join(dir, n), []byte("AAA"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, n := range []string{"d1", "d2"} { // group with 2 files
		if err := os.WriteFile(filepath.Join(dir, n), []byte("BB"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	out, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--compact", "--min-count", "3", "**/**")
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	keys := findResultKeys(t, out)
	if len(keys) != 1 {
		t.Fatalf("--min-count 3 kept %d groups, want 1: %v", len(keys), keys)
	}
}
