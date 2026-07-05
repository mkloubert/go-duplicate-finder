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

package dedup

import (
	"bytes"
	"io"
	"os"
	"sort"
	"sync"

	"github.com/mkloubert/go-duplicate-finder/internal/hasher"
	"github.com/mkloubert/go-duplicate-finder/internal/model"
	"github.com/mkloubert/go-duplicate-finder/internal/scanner"
	"github.com/mkloubert/go-duplicate-finder/internal/ui"
)

type hashed struct {
	entry scanner.FileEntry
	hash  string
}

// Find runs the complete duplicate pipeline.
func Find(files []scanner.FileEntry, jobs int, rep ui.Reporter) (*model.Output, error) {
	out := model.New()

	// 1. Group by size; only groups with >1 file are candidates.
	bySize := make(map[int64][]scanner.FileEntry)
	for _, f := range files {
		bySize[f.Size] = append(bySize[f.Size], f)
	}
	var candidates []scanner.FileEntry
	for _, g := range bySize {
		if len(g) > 1 {
			candidates = append(candidates, g...)
		}
	}
	if len(candidates) == 0 {
		return out, nil
	}

	// 2. Hash candidates in parallel.
	rep.HashStarted(len(candidates))
	hashes := hashCandidates(candidates, jobs, rep)

	// 3. Group by (size, hash).
	type key struct {
		size int64
		hash string
	}
	byKey := make(map[key][]scanner.FileEntry)
	for _, h := range hashes {
		k := key{h.entry.Size, h.hash}
		byKey[k] = append(byKey[k], h.entry)
	}

	// 4+5. Cluster within each hash group by byte comparison and build output.
	for k, group := range byKey {
		if len(group) < 2 {
			continue
		}
		clusters, err := clusterByContent(group)
		if err != nil {
			rep.Errorf("%v", err)
			continue
		}
		for _, cl := range clusters {
			if len(cl) < 2 {
				continue
			}
			sort.Slice(cl, func(i, j int) bool { return cl[i].AbsPath < cl[j].AbsPath })
			dups := make([]string, 0, len(cl)-1)
			for _, e := range cl[1:] {
				dups = append(dups, e.AbsPath)
			}
			out.Result[cl[0].AbsPath] = &model.FileResult{
				Hash:       k.hash,
				Size:       k.size,
				Duplicates: dups,
			}
		}
	}

	return out, nil
}

// hashCandidates hashes with jobs workers and reports progress.
func hashCandidates(candidates []scanner.FileEntry, jobs int, rep ui.Reporter) []hashed {
	if jobs < 1 {
		jobs = 1
	}
	if jobs > len(candidates) {
		jobs = len(candidates)
	}
	in := make(chan scanner.FileEntry)
	results := make(chan hashed)

	var wg sync.WaitGroup
	for i := 0; i < jobs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for e := range in {
				h, err := hasher.HashFile(e.AbsPath)
				if err != nil {
					rep.Errorf("%s: %v", e.AbsPath, err)
					continue
				}
				results <- hashed{entry: e, hash: h}
			}
		}()
	}

	go func() {
		for _, e := range candidates {
			in <- e
		}
		close(in)
	}()
	go func() {
		wg.Wait()
		close(results)
	}()

	var out []hashed
	done := 0
	for h := range results {
		out = append(out, h)
		done++
		rep.HashProgress(done)
	}
	return out
}

// clusterByContent splits a group of equally-hashed files into clusters of
// identical content using real byte-by-byte comparison.
func clusterByContent(group []scanner.FileEntry) ([][]scanner.FileEntry, error) {
	var clusters [][]scanner.FileEntry
	for _, e := range group {
		placed := false
		for i := range clusters {
			same, err := sameContent(clusters[i][0].AbsPath, e.AbsPath)
			if err != nil {
				return nil, err
			}
			if same {
				clusters[i] = append(clusters[i], e)
				placed = true
				break
			}
		}
		if !placed {
			clusters = append(clusters, []scanner.FileEntry{e})
		}
	}
	return clusters, nil
}

// sameContent compares two (equally sized) files byte by byte.
func sameContent(a, b string) (bool, error) {
	fa, err := os.Open(a)
	if err != nil {
		return false, err
	}
	defer fa.Close()
	fb, err := os.Open(b)
	if err != nil {
		return false, err
	}
	defer fb.Close()

	const chunk = 64 * 1024
	ba := make([]byte, chunk)
	bb := make([]byte, chunk)
	for {
		na, ea := io.ReadFull(fa, ba)
		nb, eb := io.ReadFull(fb, bb)
		if na != nb || !bytes.Equal(ba[:na], bb[:nb]) {
			return false, nil
		}
		if eb != nil && eb != io.EOF && eb != io.ErrUnexpectedEOF {
			return false, eb
		}
		if ea == io.EOF || ea == io.ErrUnexpectedEOF {
			return eb == io.EOF || eb == io.ErrUnexpectedEOF, nil
		}
		if ea != nil {
			return false, ea
		}
	}
}
