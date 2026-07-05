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

package dedup_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mkloubert/go-duplicate-finder/internal/dedup"
	"github.com/mkloubert/go-duplicate-finder/internal/hasher"
	"github.com/mkloubert/go-duplicate-finder/internal/scanner"
	"github.com/mkloubert/go-duplicate-finder/internal/ui"
)

func write(t *testing.T, path, content string) scanner.FileEntry {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return scanner.FileEntry{AbsPath: path, Size: int64(len(content))}
}

func TestFindGroupsDuplicates(t *testing.T) {
	dir := t.TempDir()
	// dup1/dup2/dup3: identical. uniq: same length as "hello" but different content.
	e1 := write(t, filepath.Join(dir, "1_dup.txt"), "hello")
	e2 := write(t, filepath.Join(dir, "2_dup.txt"), "hello")
	e3 := write(t, filepath.Join(dir, "3_dup.txt"), "hello")
	eU := write(t, filepath.Join(dir, "uniq.txt"), "world") // 5 bytes, but different content
	eS := write(t, filepath.Join(dir, "solo.txt"), "unique-and-longer")

	out, err := dedup.Find([]scanner.FileEntry{e3, e1, eU, e2, eS}, 4, hasher.BLAKE3, false, ui.Noop{})
	if err != nil {
		t.Fatal(err)
	}

	if len(out.Result) != 1 {
		t.Fatalf("expected exactly 1 duplicate cluster, got %d: %+v", len(out.Result), out.Result)
	}
	key := e1.AbsPath // lexicographically smallest
	res, ok := out.Result[key]
	if !ok {
		t.Fatalf("key %q missing in %+v", key, out.Result)
	}
	if res.Size != 5 {
		t.Errorf("expected size 5, got %d", res.Size)
	}
	if len(res.Hash) != 64 {
		t.Errorf("expected 64-character hash, got %q", res.Hash)
	}
	wantDups := []string{e2.AbsPath, e3.AbsPath}
	if len(res.Duplicates) != 2 || res.Duplicates[0] != wantDups[0] || res.Duplicates[1] != wantDups[1] {
		t.Errorf("expected duplicates %v, got %v", wantDups, res.Duplicates)
	}
	// uniq.txt (same size, different content) must NOT appear.
	if _, bad := out.Result[eU.AbsPath]; bad {
		t.Error("uniq.txt incorrectly reported as duplicate")
	}
}

func TestFindNoDuplicates(t *testing.T) {
	dir := t.TempDir()
	e1 := write(t, filepath.Join(dir, "a.txt"), "aaa")
	e2 := write(t, filepath.Join(dir, "b.txt"), "bbbb")
	out, err := dedup.Find([]scanner.FileEntry{e1, e2}, 2, hasher.BLAKE3, false, ui.Noop{})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Result) != 0 {
		t.Fatalf("expected no duplicates, got %+v", out.Result)
	}
}

func TestFindQuickModeIsApproximate(t *testing.T) {
	dir := t.TempDir()
	// Two same-size files with identical first/last 64 KB but different middles.
	prefix := strings.Repeat("A", 70*1024)
	suffix := strings.Repeat("Z", 70*1024)
	e1 := write(t, filepath.Join(dir, "1.bin"), prefix+strings.Repeat("B", 30*1024)+suffix)
	e2 := write(t, filepath.Join(dir, "2.bin"), prefix+strings.Repeat("C", 30*1024)+suffix)
	files := []scanner.FileEntry{e1, e2}

	// Exact mode reads the whole file and finds them different.
	exact, err := dedup.Find(files, 2, hasher.BLAKE3, false, ui.Noop{})
	if err != nil {
		t.Fatal(err)
	}
	if len(exact.Result) != 0 {
		t.Errorf("exact mode should not match different middles, got %+v", exact.Result)
	}

	// Quick mode samples only the (identical) ends and reports them as duplicates.
	quick, err := dedup.Find(files, 2, hasher.BLAKE3, true, ui.Noop{})
	if err != nil {
		t.Fatal(err)
	}
	if len(quick.Result) != 1 {
		t.Errorf("quick mode should match on sampled ends, got %d groups", len(quick.Result))
	}
}
