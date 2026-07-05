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
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	mdl "github.com/mkloubert/go-duplicate-finder/internal/model"
	"github.com/mkloubert/go-duplicate-finder/internal/report"
)

type fakeClip struct {
	last string
	err  error
}

func (f *fakeClip) Copy(text string) error { f.last = text; return f.err }

func demoSummary() report.Summary {
	o := mdl.New()
	o.Result["/a"] = &mdl.FileResult{Size: 100, Duplicates: []string{"/a2", "/a3"}} // recl 200
	o.Result["/b"] = &mdl.FileResult{Size: 500, Duplicates: []string{"/b2"}}        // recl 500
	o.Result["/c"] = &mdl.FileResult{Size: 10, Duplicates: []string{"/c2"}}         // recl 10
	return report.FromOutput(o)
}

func runes(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

func send(m model, msg tea.Msg) model {
	next, _ := m.Update(msg)
	return next.(model)
}

func TestModelDefaultOrderAndRows(t *testing.T) {
	m := newModel(demoSummary(), &fakeClip{})
	rows := m.rows()
	// default sort reclaimable desc: /b, /a, /c ; nothing expanded -> 3 rows
	if len(rows) != 3 || rows[0].group.Original != "/b" {
		t.Fatalf("rows = %d, first = %v", len(rows), rows[0].group.Original)
	}
}

func TestModelExpandTogglesChildRows(t *testing.T) {
	m := newModel(demoSummary(), &fakeClip{}) // cursor 0 -> /b (1 dup)
	m = send(m, tea.KeyMsg{Type: tea.KeyEnter})
	rows := m.rows()
	if len(rows) != 4 { // /b + its 1 child + /a + /c
		t.Fatalf("after expand rows = %d, want 4", len(rows))
	}
	if rows[1].kind != rowDup || rows[1].dupPath != "/b2" {
		t.Fatalf("child row = %+v", rows[1])
	}
}

func TestModelSortCycle(t *testing.T) {
	m := newModel(demoSummary(), &fakeClip{})
	m = send(m, runes("s")) // reclaimable -> size
	if m.mode != report.SortSize {
		t.Fatalf("mode = %v, want size", m.mode)
	}
	if m.rows()[0].group.Original != "/b" { // 500 largest
		t.Fatalf("size-sorted first = %v", m.rows()[0].group.Original)
	}
}

func TestModelSearchFilters(t *testing.T) {
	m := newModel(demoSummary(), &fakeClip{})
	m = send(m, runes("/"))  // enter search mode
	m = send(m, runes("a3")) // type query
	m = send(m, tea.KeyMsg{Type: tea.KeyEnter})
	rows := m.rows()
	if len(rows) != 1 || rows[0].group.Original != "/a" {
		t.Fatalf("filtered rows = %+v", rows)
	}
}

func TestModelCopyUsesClipboard(t *testing.T) {
	fc := &fakeClip{}
	m := newModel(demoSummary(), fc) // cursor 0 -> /b
	m = send(m, runes("c"))
	if fc.last != "/b" {
		t.Fatalf("copied %q, want /b", fc.last)
	}
	if m.status == "" {
		t.Error("expected a status message after copy")
	}
}

func TestModelQuit(t *testing.T) {
	m := newModel(demoSummary(), &fakeClip{})
	_, cmd := m.Update(runes("q"))
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}

func TestModelCopyChildRow(t *testing.T) {
	fc := &fakeClip{}
	m := newModel(demoSummary(), fc)            // cursor 0 -> /b (1 dup /b2)
	m = send(m, tea.KeyMsg{Type: tea.KeyEnter}) // expand /b
	m = send(m, tea.KeyMsg{Type: tea.KeyDown})  // move onto the child row
	m = send(m, runes("c"))
	if fc.last != "/b2" {
		t.Fatalf("copied %q, want /b2 (duplicate path)", fc.last)
	}
}

func TestModelSearchModeConsumesKeys(t *testing.T) {
	m := newModel(demoSummary(), &fakeClip{})
	m = send(m, runes("/")) // enter search mode
	m = send(m, runes("s")) // 's' must be typed as filter text, not trigger sort
	if m.mode != report.SortReclaimable {
		t.Errorf("mode changed to %v; 's' leaked into normal handler", m.mode)
	}
	if !m.searching {
		t.Error("should still be searching")
	}
	if m.search.Value() != "s" {
		t.Errorf("search value = %q, want %q", m.search.Value(), "s")
	}
}

func TestModelCtrlCInSearchQuits(t *testing.T) {
	m := newModel(demoSummary(), &fakeClip{})
	m = send(m, runes("/"))
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected quit command from ctrl+c while searching")
	}
}

func TestModelWindowingKeepsCursorVisible(t *testing.T) {
	o := mdl.New()
	for i := 0; i < 20; i++ {
		p := fmt.Sprintf("/g%02d", i)
		o.Result[p] = &mdl.FileResult{Size: int64(i + 1), Duplicates: []string{p + "d"}}
	}
	m := newModel(report.FromOutput(o), &fakeClip{})
	m = send(m, tea.WindowSizeMsg{Width: 80, Height: 12}) // small terminal
	for i := 0; i < 19; i++ {
		m = send(m, tea.KeyMsg{Type: tea.KeyDown}) // scroll to the last row
	}
	if m.cursor != 19 {
		t.Fatalf("cursor = %d, want 19", m.cursor)
	}

	view := m.View()
	cursorPath := m.rows()[m.cursor].group.Original
	if !strings.Contains(view, cursorPath) {
		t.Errorf("cursor row %q not visible:\n%s", cursorPath, view)
	}
	if !strings.Contains(view, "more") {
		t.Errorf("expected a scroll indicator when list exceeds height:\n%s", view)
	}
	topPath := m.rows()[0].group.Original
	if strings.Contains(view, topPath+" ") {
		t.Errorf("top row %q should be scrolled off screen:\n%s", topPath, view)
	}
}

func TestModelViewRendersHeader(t *testing.T) {
	m := newModel(demoSummary(), &fakeClip{})
	if !strings.Contains(m.View(), "dupfind summary") {
		t.Error("View should render the summary header")
	}
}
