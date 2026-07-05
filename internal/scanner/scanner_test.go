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

package scanner_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mkloubert/go-duplicate-finder/internal/scanner"
	"github.com/mkloubert/go-duplicate-finder/internal/ui"
)

func mk(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestScanFiltersAndDedups(t *testing.T) {
	dir := t.TempDir()
	mk(t, filepath.Join(dir, "a.txt"), "aaa")
	mk(t, filepath.Join(dir, "sub", "b.txt"), "bbbb")
	mk(t, filepath.Join(dir, "empty.txt"), "") // 0 bytes ⇒ skip

	// Two overlapping patterns ⇒ dedup must apply.
	got, err := scanner.Scan(dir, []string{"**/**", "a.txt"}, ui.Noop{})
	if err != nil {
		t.Fatal(err)
	}

	var paths []string
	for _, e := range got {
		paths = append(paths, e.AbsPath)
	}

	wantA := filepath.Join(dir, "a.txt")
	wantB := filepath.Join(dir, "sub", "b.txt")
	if len(got) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(got), paths)
	}
	// Sorted by path: a.txt comes before sub/b.txt.
	if got[0].AbsPath != wantA || got[1].AbsPath != wantB {
		t.Fatalf("wrong/unsorted paths: %v", paths)
	}
	if !filepath.IsAbs(got[0].AbsPath) {
		t.Errorf("expected absolute path, got %q", got[0].AbsPath)
	}
	if got[0].Size != 3 {
		t.Errorf("expected size 3 for a.txt, got %d", got[0].Size)
	}
}

func TestScanInvalidPattern(t *testing.T) {
	if _, err := scanner.Scan(t.TempDir(), []string{"[bad"}, ui.Noop{}); err == nil {
		t.Fatal("expected error for invalid pattern")
	}
}
