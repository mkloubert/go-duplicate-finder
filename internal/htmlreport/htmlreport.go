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

// Package htmlreport renders a duplicate report as a self-contained HTML page.
package htmlreport

import (
	"html/template"
	"io"

	"github.com/mkloubert/go-duplicate-finder/internal/report"
)

// page is parsed once. html/template escapes every interpolated value, so paths
// that contain HTML characters cannot inject markup.
var page = template.Must(template.New("report").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>dupfind report</title>
<style>
 body { font-family: system-ui, -apple-system, sans-serif; margin: 2rem; color: #222; background: #fafafa; }
 h1 { font-size: 1.4rem; margin: 0 0 .25rem; }
 .totals { color: #555; margin: 0 0 1.5rem; }
 .group { background: #fff; border: 1px solid #ddd; border-radius: 6px; padding: .75rem 1rem; margin-bottom: 1rem; }
 .original { font-weight: 600; word-break: break-all; }
 .meta { color: #777; font-size: .85rem; margin: .25rem 0 .5rem; }
 ul { margin: 0; padding-left: 1.25rem; }
 li { word-break: break-all; color: #444; }
 .empty { color: #777; }
</style>
</head>
<body>
<h1>dupfind report</h1>
<p class="totals">{{.NumGroups}} groups &middot; {{.NumFiles}} files &middot; {{.TotalSize}} total &middot; {{.Reclaimable}} reclaimable</p>
{{if not .Groups}}<p class="empty">No duplicates found.</p>{{end}}
{{range .Groups}}<div class="group">
<div class="original">{{.Original}}</div>
<div class="meta">{{.Size}} &middot; {{len .Duplicates}} duplicate(s) &middot; {{.Reclaimable}} reclaimable</div>
<ul>
{{range .Duplicates}}<li>{{.}}</li>
{{end}}</ul>
</div>
{{end}}</body>
</html>
`))

type groupView struct {
	Original    string
	Size        string
	Reclaimable string
	Duplicates  []string
}

type pageData struct {
	NumGroups   int
	NumFiles    int
	TotalSize   string
	Reclaimable string
	Groups      []groupView
}

// Write renders the summary as an HTML page to w. Groups keep the summary's
// order (reclaimable space, descending, by default).
func Write(w io.Writer, s report.Summary) error {
	d := pageData{
		NumGroups:   s.Totals.Groups,
		NumFiles:    s.Totals.Files,
		TotalSize:   report.Humanize(s.Totals.TotalSize),
		Reclaimable: report.Humanize(s.Totals.Reclaimable),
	}
	for _, g := range s.Groups {
		d.Groups = append(d.Groups, groupView{
			Original:    g.Original,
			Size:        report.Humanize(g.Size),
			Reclaimable: report.Humanize(g.Reclaimable()),
			Duplicates:  g.Duplicates,
		})
	}
	return page.Execute(w, d)
}
