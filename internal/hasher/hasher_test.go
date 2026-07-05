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

package hasher_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mkloubert/go-duplicate-finder/internal/hasher"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestHashFileStableAndDistinct(t *testing.T) {
	dir := t.TempDir()
	a := writeFile(t, dir, "a.txt", "hello world")
	b := writeFile(t, dir, "b.txt", "hello world")
	c := writeFile(t, dir, "c.txt", "different")

	ha, err := hasher.HashFile(a)
	if err != nil {
		t.Fatal(err)
	}
	hb, _ := hasher.HashFile(b)
	hc, _ := hasher.HashFile(c)

	if len(ha) != 64 {
		t.Fatalf("expected 64 hex characters, got %d (%q)", len(ha), ha)
	}
	if ha != hb {
		t.Errorf("same content ⇒ expected same hash: %s != %s", ha, hb)
	}
	if ha == hc {
		t.Errorf("different content ⇒ expected different hash, both are %s", ha)
	}
}

func TestHashFileMissing(t *testing.T) {
	if _, err := hasher.HashFile("/does/not/exist"); err == nil {
		t.Fatal("expected error for missing file")
	}
}
