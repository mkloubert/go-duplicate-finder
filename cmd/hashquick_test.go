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
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindHashAlgorithms(t *testing.T) {
	dir := makeSizedDupTree(t) // two duplicate groups
	for _, algo := range []string{"blake3", "sha256", "xxh3"} {
		out, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--compact", "--hash", algo, "**/**")
		if err != nil {
			t.Fatalf("%s: %v", algo, err)
		}
		if got := len(findResultKeys(t, out)); got != 2 {
			t.Errorf("%s: expected 2 groups, got %d", algo, got)
		}
	}
}

func TestFindInvalidHash(t *testing.T) {
	dir := makeSizedDupTree(t)
	if _, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--hash", "md5", "**/**"); err == nil {
		t.Fatal("expected error for invalid --hash")
	}
}

func TestFindQuickApproximate(t *testing.T) {
	dir := t.TempDir()
	prefix := strings.Repeat("A", 70*1024)
	suffix := strings.Repeat("Z", 70*1024)
	if err := os.WriteFile(filepath.Join(dir, "1.bin"), []byte(prefix+strings.Repeat("B", 30*1024)+suffix), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "2.bin"), []byte(prefix+strings.Repeat("C", 30*1024)+suffix), 0o644); err != nil {
		t.Fatal(err)
	}

	exact, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--compact", "**/**")
	if err != nil {
		t.Fatal(err)
	}
	if got := len(findResultKeys(t, exact)); got != 0 {
		t.Errorf("exact mode should find no duplicates, got %d:\n%s", got, exact)
	}

	quick, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--compact", "--quick", "**/**")
	if err != nil {
		t.Fatal(err)
	}
	if got := len(findResultKeys(t, quick)); got != 1 {
		t.Errorf("quick mode should find 1 group, got %d:\n%s", got, quick)
	}
}

func TestFindFollowSymlinksNoFalseDuplicate(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "real.txt")
	if err := os.WriteFile(real, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "dup.txt"), []byte("payload"), 0o644); err != nil { // genuine duplicate
		t.Fatal(err)
	}
	if err := os.Symlink(real, filepath.Join(dir, "link.txt")); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}

	out, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--compact", "--follow-symlinks", "**/**")
	if err != nil {
		t.Fatal(err)
	}

	var parsed struct {
		Result map[string]struct {
			Duplicates []string `json:"duplicates"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("stdout not JSON: %v\n%s", err, out)
	}
	if len(parsed.Result) != 1 {
		t.Fatalf("expected 1 group, got %d: %s", len(parsed.Result), out)
	}
	// The symlink must collapse with its target (same inode), so the group has
	// exactly one duplicate — not a spurious third entry.
	for _, g := range parsed.Result {
		if len(g.Duplicates) != 1 {
			t.Errorf("follow + inode-dedup should yield 1 duplicate, got %d: %v", len(g.Duplicates), g.Duplicates)
		}
	}
}
