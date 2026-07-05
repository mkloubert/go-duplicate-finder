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

package summaryui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/mkloubert/go-duplicate-finder/internal/report"
)

// RenderStatic writes a one-shot summary to w. The lipgloss renderer is bound
// to w, so color is used only when w is a color-capable terminal and NO_COLOR
// is unset.
func RenderStatic(w io.Writer, s report.Summary, mode report.SortMode) {
	s.SortBy(mode)

	r := lipgloss.NewRenderer(w)
	bold := r.NewStyle().Bold(true)
	dim := r.NewStyle().Faint(true)

	fmt.Fprintln(w, bold.Render(fmt.Sprintf(
		"%d groups · %d files · %s total · %s reclaimable",
		s.Totals.Groups, s.Totals.Files,
		report.Humanize(s.Totals.TotalSize), report.Humanize(s.Totals.Reclaimable))))
	fmt.Fprintln(w)

	if len(s.Groups) == 0 {
		fmt.Fprintln(w, "No duplicates found.")
		return
	}

	for _, g := range s.Groups {
		fmt.Fprintf(w, "%s  %s  %s\n",
			bold.Render(g.Original),
			report.Humanize(g.Size),
			dim.Render(fmt.Sprintf("(%d dupes, %s reclaimable)",
				len(g.Duplicates), report.Humanize(g.Reclaimable()))))
		for _, d := range g.Duplicates {
			fmt.Fprintf(w, "  %s %s\n", dim.Render("↳"), d)
		}
	}
}
