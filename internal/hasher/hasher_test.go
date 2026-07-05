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
	"strings"
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

	ha, err := hasher.HashFile(a, hasher.BLAKE3)
	if err != nil {
		t.Fatal(err)
	}
	hb, _ := hasher.HashFile(b, hasher.BLAKE3)
	hc, _ := hasher.HashFile(c, hasher.BLAKE3)

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
	if _, err := hasher.HashFile("/does/not/exist", hasher.BLAKE3); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestParseAlgorithm(t *testing.T) {
	valid := map[string]hasher.Algorithm{
		"":       hasher.BLAKE3,
		"blake3": hasher.BLAKE3,
		"SHA256": hasher.SHA256,
		"xxh3":   hasher.XXH3,
	}
	for in, want := range valid {
		got, err := hasher.ParseAlgorithm(in)
		if err != nil || got != want {
			t.Errorf("ParseAlgorithm(%q) = %v,%v want %v", in, got, err, want)
		}
	}
	if _, err := hasher.ParseAlgorithm("md5"); err == nil {
		t.Error("expected error for unknown algorithm")
	}
}

func TestHashFileAllAlgorithms(t *testing.T) {
	dir := t.TempDir()
	a := writeFile(t, dir, "a.txt", "hello world")
	b := writeFile(t, dir, "b.txt", "hello world")
	c := writeFile(t, dir, "c.txt", "different")

	for _, algo := range []hasher.Algorithm{hasher.BLAKE3, hasher.SHA256, hasher.XXH3} {
		ha, err := hasher.HashFile(a, algo)
		if err != nil {
			t.Fatalf("%s: %v", algo, err)
		}
		hb, _ := hasher.HashFile(b, algo)
		hc, _ := hasher.HashFile(c, algo)
		if ha == "" {
			t.Errorf("%s produced an empty digest", algo)
		}
		if ha != hb {
			t.Errorf("%s: same content should hash equally: %s != %s", algo, ha, hb)
		}
		if ha == hc {
			t.Errorf("%s: different content should differ, both %s", algo, ha)
		}
	}
}

func TestHashSample(t *testing.T) {
	dir := t.TempDir()
	const sample = 64 * 1024

	// Two files with identical first/last `sample` bytes but different middles.
	prefix := strings.Repeat("A", 70*1024)
	suffix := strings.Repeat("Z", 70*1024)
	c1 := prefix + strings.Repeat("B", 30*1024) + suffix
	c2 := prefix + strings.Repeat("C", 30*1024) + suffix
	p1 := writeFile(t, dir, "1.bin", c1)
	p2 := writeFile(t, dir, "2.bin", c2)

	s1, err := hasher.HashSample(p1, hasher.BLAKE3, sample)
	if err != nil {
		t.Fatal(err)
	}
	s2, _ := hasher.HashSample(p2, hasher.BLAKE3, sample)
	if s1 != s2 {
		t.Errorf("sample hashes should match on identical ends: %s != %s", s1, s2)
	}

	// The full hashes must differ (the middles differ).
	f1, _ := hasher.HashFile(p1, hasher.BLAKE3)
	f2, _ := hasher.HashFile(p2, hasher.BLAKE3)
	if f1 == f2 {
		t.Error("full hashes should differ on different middles")
	}
}
