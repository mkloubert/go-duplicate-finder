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

func TestFindMinSize(t *testing.T) {
	dir := makeSizedDupTree(t) // a-group 100 bytes, b-group 10 bytes
	out, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--compact", "--min-size", "50", "**/**")
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	keys := findResultKeys(t, out)
	if len(keys) != 1 {
		t.Fatalf("--min-size 50 kept %d groups, want 1: %v", len(keys), keys)
	}
	if _, ok := keys[filepath.Join(dir, "a1.bin")]; !ok {
		t.Errorf("--min-size 50 should keep the big group (a1.bin): %v", keys)
	}
}

func TestFindMaxSize(t *testing.T) {
	dir := makeSizedDupTree(t)
	out, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--compact", "--max-size", "50", "**/**")
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	keys := findResultKeys(t, out)
	if len(keys) != 1 {
		t.Fatalf("--max-size 50 kept %d groups, want 1: %v", len(keys), keys)
	}
	if _, ok := keys[filepath.Join(dir, "b1.bin")]; !ok {
		t.Errorf("--max-size 50 should keep the small group (b1.bin): %v", keys)
	}
}

func TestFindExclude(t *testing.T) {
	dir := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("keep1.txt", "X")
	write("keep2.txt", "X")
	write(".git/g1", "Y")
	write(".git/g2", "Y")

	// Without exclude both duplicate groups appear.
	out, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--compact", "**/**")
	if err != nil {
		t.Fatal(err)
	}
	if got := len(findResultKeys(t, out)); got != 2 {
		t.Fatalf("without exclude want 2 groups, got %d:\n%s", got, out)
	}

	// Excluding .git leaves only the keep group.
	out, err = captureScript(t, "find", "--cwd", dir, "--no-tui", "--compact", "--exclude", ".git/**", "**/**")
	if err != nil {
		t.Fatal(err)
	}
	keys := findResultKeys(t, out)
	if len(keys) != 1 {
		t.Fatalf("with exclude want 1 group, got %d: %v", len(keys), keys)
	}
	for k := range keys {
		if strings.Contains(k, ".git") {
			t.Errorf("excluded .git path present: %s", k)
		}
	}
}

func TestFindInvalidMinSize(t *testing.T) {
	dir := makeSizedDupTree(t)
	if _, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--min-size", "huge", "**/**"); err == nil {
		t.Fatal("expected error for invalid --min-size")
	}
}
