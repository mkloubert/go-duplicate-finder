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

package hasher

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"

	"github.com/zeebo/xxh3"
	"lukechampine.com/blake3"
)

// Algorithm selects the content hash used to compare files.
type Algorithm int

const (
	BLAKE3 Algorithm = iota
	SHA256
	XXH3
)

// ParseAlgorithm maps a name to an Algorithm (case-insensitive). Empty means the
// default, BLAKE3.
func ParseAlgorithm(s string) (Algorithm, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "blake3":
		return BLAKE3, nil
	case "sha256":
		return SHA256, nil
	case "xxh3":
		return XXH3, nil
	default:
		return BLAKE3, fmt.Errorf("unknown hash algorithm %q (use blake3, sha256, or xxh3)", s)
	}
}

// String returns the algorithm name.
func (a Algorithm) String() string {
	switch a {
	case SHA256:
		return "sha256"
	case XXH3:
		return "xxh3"
	default:
		return "blake3"
	}
}

// newHash creates a fresh hash.Hash for the algorithm.
func newHash(algo Algorithm) hash.Hash {
	switch algo {
	case SHA256:
		return sha256.New()
	case XXH3:
		return xxh3.New()
	default:
		return blake3.New(32, nil)
	}
}

// HashFile hashes the whole file with the given algorithm and returns the digest
// as lowercase hex.
func HashFile(path string, algo Algorithm) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := newHash(algo)
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// HashSample hashes only the first and last sampleBytes of the file (or the whole
// file when it is at most 2*sampleBytes). It is faster than HashFile but
// approximate: two files of the same size with identical ends but different
// middles produce the same digest.
func HashSample(path string, algo Algorithm, sampleBytes int64) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	size := info.Size()

	h := newHash(algo)
	if sampleBytes <= 0 || size <= 2*sampleBytes {
		if _, err := io.Copy(h, f); err != nil {
			return "", err
		}
		return hex.EncodeToString(h.Sum(nil)), nil
	}

	if _, err := io.CopyN(h, f, sampleBytes); err != nil {
		return "", err
	}
	if _, err := f.Seek(size-sampleBytes, io.SeekStart); err != nil {
		return "", err
	}
	if _, err := io.CopyN(h, f, sampleBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
