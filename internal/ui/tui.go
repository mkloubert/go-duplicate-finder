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

package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// --- Bubble Tea messages ---

type scanMsg struct{}
type fileMsg struct{}
type hashStartMsg struct{ total int }
type hashProgMsg struct{ done int }
type doneMsg struct{}

// --- Model ---

type tuiModel struct {
	spinner  spinner.Model
	progress progress.Model
	phase    int // 0=scan, 1=hash, 2=done
	files    int
	total    int
	done     int
}

func newTUIModel() tuiModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return tuiModel{
		spinner:  s,
		progress: progress.New(progress.WithDefaultGradient()),
	}
}

func (m tuiModel) Init() tea.Cmd { return m.spinner.Tick }

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case scanMsg:
		m.phase = 0
		return m, nil
	case fileMsg:
		m.files++
		return m, nil
	case hashStartMsg:
		m.phase = 1
		m.total = msg.total
		return m, nil
	case hashProgMsg:
		m.done = msg.done
		return m, nil
	case doneMsg:
		m.phase = 2
		return m, tea.Quit
	}
	return m, nil
}

func (m tuiModel) View() string {
	switch m.phase {
	case 0:
		return fmt.Sprintf("%s Scanning files … (%d found)\n", m.spinner.View(), m.files)
	case 1:
		pct := 0.0
		if m.total > 0 {
			pct = float64(m.done) / float64(m.total)
		}
		return fmt.Sprintf("%s Hashing %d/%d\n%s\n",
			m.spinner.View(), m.done, m.total, m.progress.ViewAs(pct))
	default:
		return "Done.\n"
	}
}

// --- Reporter implementation ---

// TUI drives a Bubble Tea program in a goroutine and renders to STDERR.
type TUI struct {
	prog *tea.Program
}

// NewTUI starts the TUI. Rendering goes explicitly to STDERR so that STDOUT
// remains reserved for JSON output.
func NewTUI() *TUI {
	p := tea.NewProgram(newTUIModel(), tea.WithOutput(os.Stderr))
	go func() { _, _ = p.Run() }()
	return &TUI{prog: p}
}

func (t *TUI) ScanStarted()          { t.prog.Send(scanMsg{}) }
func (t *TUI) FileFound()            { t.prog.Send(fileMsg{}) }
func (t *TUI) HashStarted(total int) { t.prog.Send(hashStartMsg{total: total}) }
func (t *TUI) HashProgress(done int) { t.prog.Send(hashProgMsg{done: done}) }

// Done stops the program and waits until the terminal is cleaned up.
func (t *TUI) Done() {
	t.prog.Send(doneMsg{})
	t.prog.Wait()
}

func (t *TUI) Errorf(format string, a ...any) {
	t.prog.Println(fmt.Sprintf("Error: "+format, a...))
}
