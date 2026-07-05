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

// fileKey identifies a physical file by device and inode.
type fileKey struct{ dev, ino uint64 }

// Options configures a scan.
type Options struct {
	Patterns       []string // include globs (relative to the base dir)
	FollowSymlinks bool     // resolve symlinks and include their targets
	MinSize        int64    // skip files smaller than this (0 = no minimum)
	MaxSize        int64    // skip files larger than this (0 = no maximum)
	Exclude        []string // globs (relative, slash paths) whose matches are skipped
}

// Scan expands opts.Patterns relative to baseDir and returns regular, non-empty,
// deduplicated and sorted files by absolute path. Symlinks are skipped unless
// FollowSymlinks is set. Files that share the same physical storage (hardlinks,
// or a symlink and its target when following) collapse to a single entry, so
// they are not reported as reclaimable duplicates. Exclude globs and the
// Min/MaxSize bounds remove files before hashing.
func Scan(baseDir string, opts Options, rep ui.Reporter) ([]FileEntry, error) {
	rep.ScanStarted()

	fsys := os.DirFS(baseDir)
	seen := make(map[string]struct{})
	seenID := make(map[fileKey]struct{})
	var entries []FileEntry

	for _, pat := range opts.Patterns {
		matches, err := doublestar.Glob(fsys, pat)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %q: %w", pat, err)
		}
		for _, rel := range matches {
			excl, err := isExcluded(rel, opts.Exclude)
			if err != nil {
				return nil, err
			}
			if excl {
				continue
			}

			abs := filepath.Join(baseDir, filepath.FromSlash(rel))
			if _, ok := seen[abs]; ok {
				continue
			}

			var info os.FileInfo
			if opts.FollowSymlinks {
				info, err = os.Stat(abs) // resolves symlinks to their target
			} else {
				info, err = os.Lstat(abs) // a symlink stays non-regular and is skipped
			}
			if err != nil {
				rep.Errorf("%s: %v", abs, err)
				continue
			}
			if !info.Mode().IsRegular() {
				continue
			}
			size := info.Size()
			if size == 0 {
				continue
			}
			if !sizeInRange(size, opts.MinSize, opts.MaxSize) {
				continue
			}
			if dev, ino, ok := fileID(info); ok {
				key := fileKey{dev: dev, ino: ino}
				if _, dup := seenID[key]; dup {
					continue
				}
				seenID[key] = struct{}{}
			}
			seen[abs] = struct{}{}
			entries = append(entries, FileEntry{AbsPath: abs, Size: size})
			rep.FileFound()
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].AbsPath < entries[j].AbsPath
	})
	return entries, nil
}

// sizeInRange reports whether size is within [min, max]; a bound of 0 disables it.
func sizeInRange(size, min, max int64) bool {
	if min > 0 && size < min {
		return false
	}
	if max > 0 && size > max {
		return false
	}
	return true
}

// isExcluded reports whether the relative path matches any exclude glob.
func isExcluded(rel string, patterns []string) (bool, error) {
	for _, p := range patterns {
		ok, err := doublestar.Match(p, rel)
		if err != nil {
			return false, fmt.Errorf("invalid exclude pattern %q: %w", p, err)
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}
