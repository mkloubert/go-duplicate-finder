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

package report

import "github.com/mkloubert/go-duplicate-finder/internal/model"

// Group is one set of identical files: the first occurrence and its duplicates.
type Group struct {
	Original   string
	Size       int64
	Duplicates []string
}

// Reclaimable is the space freed by removing the duplicate copies.
func (g Group) Reclaimable() int64 { return g.Size * int64(len(g.Duplicates)) }

// FileCount is the number of files in the group (original + duplicates).
func (g Group) FileCount() int { return 1 + len(g.Duplicates) }

// Totals are the aggregate numbers shown in the header.
type Totals struct {
	Groups      int
	Files       int
	TotalSize   int64
	Reclaimable int64
}

// Summary is the whole report in a form ready to render.
type Summary struct {
	Groups []Group
	Totals Totals
}

// FromOutput builds a Summary from a parsed report and sorts it by the default
// mode (reclaimable space, descending).
func FromOutput(o *model.Output) Summary {
	var s Summary
	if o == nil {
		return s
	}
	for path, res := range o.Result {
		if res == nil {
			continue
		}
		g := Group{
			Original:   path,
			Size:       res.Size,
			Duplicates: append([]string(nil), res.Duplicates...),
		}
		s.Groups = append(s.Groups, g)
		s.Totals.Groups++
		s.Totals.Files += g.FileCount()
		s.Totals.TotalSize += g.Size * int64(g.FileCount())
		s.Totals.Reclaimable += g.Reclaimable()
	}
	s.SortBy(SortReclaimable)
	return s
}
