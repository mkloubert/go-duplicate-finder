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

package report_test

import (
	"testing"

	"github.com/mkloubert/go-duplicate-finder/internal/model"
	"github.com/mkloubert/go-duplicate-finder/internal/report"
)

func sample() *model.Output {
	o := model.New()
	// reclaimable: a=100*2=200, b=500*1=500, c=10*3=30
	o.Result["/a"] = &model.FileResult{Hash: "ha", Size: 100, Duplicates: []string{"/a2", "/a3"}}
	o.Result["/b"] = &model.FileResult{Hash: "hb", Size: 500, Duplicates: []string{"/b2"}}
	o.Result["/c"] = &model.FileResult{Hash: "hc", Size: 10, Duplicates: []string{"/c2", "/c3", "/c4"}}
	return o
}

func TestFromOutputTotalsAndDefaultSort(t *testing.T) {
	s := report.FromOutput(sample())
	if s.Totals.Groups != 3 {
		t.Errorf("groups = %d, want 3", s.Totals.Groups)
	}
	// files: a=3, b=2, c=4 => 9
	if s.Totals.Files != 9 {
		t.Errorf("files = %d, want 9", s.Totals.Files)
	}
	// total size: 100*3 + 500*2 + 10*4 = 300+1000+40 = 1340
	if s.Totals.TotalSize != 1340 {
		t.Errorf("total size = %d, want 1340", s.Totals.TotalSize)
	}
	// reclaimable: 200 + 500 + 30 = 730
	if s.Totals.Reclaimable != 730 {
		t.Errorf("reclaimable = %d, want 730", s.Totals.Reclaimable)
	}
	// default sort = reclaimable desc: b(500), a(200), c(30)
	if s.Groups[0].Original != "/b" || s.Groups[1].Original != "/a" || s.Groups[2].Original != "/c" {
		t.Errorf("default order = %s,%s,%s", s.Groups[0].Original, s.Groups[1].Original, s.Groups[2].Original)
	}
}

func TestGroupDerived(t *testing.T) {
	g := report.Group{Original: "/x", Size: 7, Duplicates: []string{"/y", "/z"}}
	if g.Reclaimable() != 14 {
		t.Errorf("reclaimable = %d, want 14", g.Reclaimable())
	}
	if g.FileCount() != 3 {
		t.Errorf("filecount = %d, want 3", g.FileCount())
	}
}

func TestSortBySizeAndDupCountAndPath(t *testing.T) {
	s := report.FromOutput(sample())

	s.SortBy(report.SortSize)
	if s.Groups[0].Original != "/b" { // 500
		t.Errorf("size sort first = %s, want /b", s.Groups[0].Original)
	}

	s.SortBy(report.SortDupCount)
	if s.Groups[0].Original != "/c" { // 3 dupes
		t.Errorf("dupcount sort first = %s, want /c", s.Groups[0].Original)
	}

	s.SortBy(report.SortPath)
	if s.Groups[0].Original != "/a" || s.Groups[2].Original != "/c" {
		t.Errorf("path sort = %s..%s, want /a../c", s.Groups[0].Original, s.Groups[2].Original)
	}
}

func TestSortTieBreakByPath(t *testing.T) {
	o := model.New()
	o.Result["/z"] = &model.FileResult{Size: 5, Duplicates: []string{"/z2"}}
	o.Result["/a"] = &model.FileResult{Size: 5, Duplicates: []string{"/a2"}}
	s := report.FromOutput(o)
	s.SortBy(report.SortReclaimable) // equal reclaimable => tie-break by path asc
	if s.Groups[0].Original != "/a" || s.Groups[1].Original != "/z" {
		t.Errorf("tie-break order = %s,%s, want /a,/z", s.Groups[0].Original, s.Groups[1].Original)
	}
}

func TestFilter(t *testing.T) {
	s := report.FromOutput(sample())
	if got := s.Filter(""); len(got) != 3 {
		t.Errorf("empty filter returned %d, want 3", len(got))
	}
	// match on original
	if got := s.Filter("/b"); len(got) != 1 || got[0].Original != "/b" {
		t.Errorf("filter /b = %+v", got)
	}
	// match on a duplicate path (case-insensitive)
	if got := s.Filter("A3"); len(got) != 1 || got[0].Original != "/a" {
		t.Errorf("filter A3 = %+v", got)
	}
	if got := s.Filter("nope"); len(got) != 0 {
		t.Errorf("filter nope = %+v", got)
	}
}

func TestSortModeStringAndNext(t *testing.T) {
	if report.SortReclaimable.String() != "reclaimable" {
		t.Errorf("string = %q", report.SortReclaimable.String())
	}
	if report.SortReclaimable.Next() != report.SortSize {
		t.Error("Next(reclaimable) should be size")
	}
	if report.SortPath.Next() != report.SortReclaimable {
		t.Error("Next(path) should wrap to reclaimable")
	}
}

func TestHumanize(t *testing.T) {
	cases := map[int64]string{
		0:          "0 B",
		512:        "512 B",
		1024:       "1.0 KB",
		1536:       "1.5 KB",
		4718592:    "4.5 MB",
		2147483648: "2.0 GB",
	}
	for in, want := range cases {
		if got := report.Humanize(in); got != want {
			t.Errorf("Humanize(%d) = %q, want %q", in, got, want)
		}
	}
}
