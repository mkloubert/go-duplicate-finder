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

package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/mkloubert/go-duplicate-finder/internal/ui"
)

// FileEntry is a found regular file with its size.
type FileEntry struct {
	AbsPath string
	Size    int64
}

// Scan expands the glob patterns relative to baseDir and returns regular,
// non-empty, deduplicated and sorted files by absolute path.
func Scan(baseDir string, patterns []string, rep ui.Reporter) ([]FileEntry, error) {
	rep.ScanStarted()

	fsys := os.DirFS(baseDir)
	seen := make(map[string]struct{})
	var entries []FileEntry

	for _, pat := range patterns {
		matches, err := doublestar.Glob(fsys, pat)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %q: %w", pat, err)
		}
		for _, rel := range matches {
			abs := filepath.Join(baseDir, filepath.FromSlash(rel))
			if _, ok := seen[abs]; ok {
				continue
			}
			info, err := os.Lstat(abs)
			if err != nil {
				rep.Errorf("%s: %v", abs, err)
				continue
			}
			if !info.Mode().IsRegular() {
				continue
			}
			if info.Size() == 0 {
				continue
			}
			seen[abs] = struct{}{}
			entries = append(entries, FileEntry{AbsPath: abs, Size: info.Size()})
			rep.FileFound()
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].AbsPath < entries[j].AbsPath
	})
	return entries, nil
}
