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
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mkloubert/go-duplicate-finder/internal/report"
)

var (
	headerStyle = lipgloss.NewStyle().Bold(true)
	dimStyle    = lipgloss.NewStyle().Faint(true)
	selStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
)

type rowKind int

const (
	rowGroup rowKind = iota
	rowDup
)

// row is one visible line: a group header or one of its duplicate paths.
type row struct {
	kind    rowKind
	group   report.Group
	dupPath string
}

// copyText is the path this row copies to the clipboard.
func (r row) copyText() string {
	if r.kind == rowDup {
		return r.dupPath
	}
	return r.group.Original
}

type model struct {
	summary   report.Summary
	mode      report.SortMode
	visible   []report.Group
	cursor    int
	expanded  map[string]bool
	searching bool
	search    textinput.Model
	query     string
	clip      Clipboard
	status    string
	height    int
}

// newModel builds the interactive model, applying the default sort.
func newModel(s report.Summary, clip Clipboard) model {
	s.SortBy(report.SortReclaimable)
	ti := textinput.New()
	ti.Placeholder = "filter…"
	ti.Prompt = "/"
	m := model{
		summary:  s,
		mode:     report.SortReclaimable,
		expanded: map[string]bool{},
		search:   ti,
		clip:     clip,
	}
	m.applyFilter()
	return m
}

// applyFilter recomputes the visible groups and clamps the cursor.
func (m *model) applyFilter() {
	m.visible = m.summary.Filter(m.query)
	if n := len(m.rows()); m.cursor >= n {
		m.cursor = max(0, n-1)
	}
}

// rows flattens visible groups and the children of expanded groups.
func (m model) rows() []row {
	var rows []row
	for _, g := range m.visible {
		rows = append(rows, row{kind: rowGroup, group: g})
		if m.expanded[g.Original] {
			for _, d := range g.Duplicates {
				rows = append(rows, row{kind: rowDup, group: g, dupPath: d})
			}
		}
	}
	return rows
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if m.searching {
			return m.updateSearch(msg)
		}
		return m.updateNormal(msg)
	}
	return m, nil
}

func (m model) updateNormal(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	rows := m.rows()
	switch key.String() {
	case "q", "esc", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(rows)-1 {
			m.cursor++
		}
	case "enter":
		if m.cursor < len(rows) && rows[m.cursor].kind == rowGroup {
			orig := rows[m.cursor].group.Original
			m.expanded[orig] = !m.expanded[orig]
		}
	case "s":
		m.mode = m.mode.Next()
		m.summary.SortBy(m.mode)
		m.cursor = 0
		m.applyFilter()
		m.status = "Sorted by " + m.mode.String()
	case "/":
		m.searching = true
		m.search.SetValue(m.query)
		m.search.Focus()
		return m, textinput.Blink
	case "c":
		if m.cursor < len(rows) {
			text := rows[m.cursor].copyText()
			if err := m.clip.Copy(text); err != nil {
				m.status = "Copy failed: " + err.Error()
			} else {
				m.status = "Copied " + text
			}
		}
	}
	return m, nil
}

func (m model) updateSearch(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "enter":
		m.query = m.search.Value()
		m.searching = false
		m.search.Blur()
		m.cursor = 0
		m.applyFilter()
		return m, nil
	case "esc":
		m.searching = false
		m.search.Blur()
		return m, nil
	}
	var cmd tea.Cmd
	m.search, cmd = m.search.Update(key)
	return m, cmd
}

func (m model) View() string {
	var b strings.Builder
	t := m.summary.Totals
	b.WriteString(headerStyle.Render(fmt.Sprintf(
		"dupfind summary — %d groups · %d files · %s total · %s reclaimable",
		t.Groups, t.Files, report.Humanize(t.TotalSize), report.Humanize(t.Reclaimable))))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("sort: " + m.mode.String()))
	if m.query != "" {
		b.WriteString(dimStyle.Render("  filter: " + m.query))
	}
	b.WriteString("\n\n")

	rows := m.rows()
	if len(rows) == 0 {
		b.WriteString("No matching duplicates.\n")
	}

	start, end := m.window(len(rows))
	if start > 0 {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  ↑ %d more", start)) + "\n")
	}
	for gi := start; gi < end; gi++ {
		r := rows[gi]
		cursor := "  "
		if gi == m.cursor {
			cursor = selStyle.Render("▸ ")
		}
		switch r.kind {
		case rowGroup:
			marker := "▸"
			if m.expanded[r.group.Original] {
				marker = "▾"
			}
			b.WriteString(cursor + fmt.Sprintf("%s %s  %s  %s",
				marker, r.group.Original, report.Humanize(r.group.Size),
				dimStyle.Render(fmt.Sprintf("(%d dupes)", len(r.group.Duplicates)))) + "\n")
		case rowDup:
			b.WriteString(cursor + dimStyle.Render("    ↳ "+r.dupPath) + "\n")
		}
	}
	if end < len(rows) {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  ↓ %d more", len(rows)-end)) + "\n")
	}

	if m.searching {
		b.WriteString("\n" + m.search.View() + "\n")
	}
	b.WriteString("\n")
	if m.status != "" {
		b.WriteString(statusStyle.Render(m.status) + "\n")
	}
	b.WriteString(dimStyle.Render("↑/↓ move · enter expand · s sort · / search · c copy · q quit"))
	return b.String()
}

// window returns the [start,end) range of rows to show so the cursor stays
// visible within the available height. A height of 0 (size not yet known)
// shows all rows.
func (m model) window(total int) (int, int) {
	if m.height <= 0 {
		return 0, total
	}
	reserve := 6 // header (3 lines) + blank + status placeholder + help
	if m.searching {
		reserve += 2
	}
	if m.status != "" {
		reserve++
	}
	capacity := m.height - reserve
	if capacity < 1 {
		capacity = 1
	}
	if total <= capacity {
		return 0, total
	}
	start := 0
	if m.cursor >= capacity {
		start = m.cursor - capacity + 1
	}
	if start+capacity > total {
		start = total - capacity
	}
	if start < 0 {
		start = 0
	}
	return start, min(start+capacity, total)
}

// Run starts the interactive UI. When in is non-nil it is used for key input
// (e.g. /dev/tty when the report came from STDIN).
func Run(s report.Summary, clip Clipboard, in *os.File) error {
	opts := []tea.ProgramOption{}
	if in != nil {
		opts = append(opts, tea.WithInput(in))
	}
	p := tea.NewProgram(newModel(s, clip), opts...)
	_, err := p.Run()
	return err
}
