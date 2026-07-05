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

package script_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mkloubert/go-duplicate-finder/internal/model"
	"github.com/mkloubert/go-duplicate-finder/internal/report"
	"github.com/mkloubert/go-duplicate-finder/internal/script"
)

func TestDetectFrom(t *testing.T) {
	cases := []struct {
		goos, shell string
		want        script.Shell
	}{
		{"windows", "", script.PowerShell},
		{"windows", "/bin/bash", script.PowerShell},
		{"linux", "/bin/zsh", script.Zsh},
		{"linux", "/usr/bin/bash", script.Bash},
		{"darwin", "", script.Bash},
	}
	for _, c := range cases {
		if got := script.DetectFrom(c.goos, c.shell); got != c.want {
			t.Errorf("DetectFrom(%q,%q) = %v, want %v", c.goos, c.shell, got, c.want)
		}
	}
}

func TestParseShell(t *testing.T) {
	valid := map[string]script.Shell{
		"bash": script.Bash, "ZSH": script.Zsh,
		"PowerShell": script.PowerShell, "pwsh": script.PowerShell,
	}
	for in, want := range valid {
		got, err := script.ParseShell(in)
		if err != nil || got != want {
			t.Errorf("ParseShell(%q) = %v,%v want %v", in, got, err, want)
		}
	}
	if _, err := script.ParseShell("fish"); err == nil {
		t.Error("expected error for unknown shell")
	}
}

func sampleSummary() report.Summary {
	o := model.New()
	o.Result["/photos/a.jpg"] = &model.FileResult{
		Hash: "abc123", Size: 2097152, Duplicates: []string{"/backup/a.jpg", "/tmp/a.jpg"},
	}
	return report.FromOutput(o)
}

func TestGenerateBash(t *testing.T) {
	var buf bytes.Buffer
	script.Generate(&buf, sampleSummary(), script.Bash)
	out := buf.String()

	for _, want := range []string{
		"#!/usr/bin/env bash",
		"# Removes duplicate files",
		"# 2 delete commands across 1 groups",
		"# Group 1  abc123  (2.0 MB, 2 duplicates)",
		"# keep: /photos/a.jpg",
		"# rm -f -- '/photos/a.jpg'",
		"rm -f -- '/backup/a.jpg'",
		"rm -f -- '/tmp/a.jpg'",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("bash output missing %q:\n%s", want, out)
		}
	}
	// the duplicate lines must be active, not commented
	if strings.Contains(out, "# rm -f -- '/backup/a.jpg'") {
		t.Errorf("duplicate should not be commented out:\n%s", out)
	}
}

func TestGeneratePowerShell(t *testing.T) {
	var buf bytes.Buffer
	script.Generate(&buf, sampleSummary(), script.PowerShell)
	out := buf.String()

	for _, want := range []string{
		"#!/usr/bin/env pwsh",
		"# Remove-Item -LiteralPath '/photos/a.jpg' -Force",
		"Remove-Item -LiteralPath '/backup/a.jpg' -Force",
		"Remove-Item -LiteralPath '/tmp/a.jpg' -Force",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("powershell output missing %q:\n%s", want, out)
		}
	}
}

func TestGenerateQuoting(t *testing.T) {
	o := model.New()
	o.Result["/weird/it's.txt"] = &model.FileResult{
		Hash: "h", Size: 5, Duplicates: []string{"/copy/it's.txt"},
	}
	s := report.FromOutput(o)

	var bash bytes.Buffer
	script.Generate(&bash, s, script.Bash)
	if !strings.Contains(bash.String(), `rm -f -- '/copy/it'\''s.txt'`) {
		t.Errorf("bash quoting wrong:\n%s", bash.String())
	}

	var ps bytes.Buffer
	script.Generate(&ps, s, script.PowerShell)
	if !strings.Contains(ps.String(), `Remove-Item -LiteralPath '/copy/it''s.txt' -Force`) {
		t.Errorf("powershell quoting wrong:\n%s", ps.String())
	}
}

func TestGenerateEmpty(t *testing.T) {
	var buf bytes.Buffer
	script.Generate(&buf, report.FromOutput(model.New()), script.Bash)
	out := buf.String()
	if !strings.Contains(out, "# No duplicates found.") {
		t.Errorf("empty output = %q", out)
	}
	if strings.Contains(out, "rm -f") {
		t.Errorf("empty report should have no delete commands:\n%s", out)
	}
}

// A path or hash containing a newline must not break out of its comment line
// and become an active command. Only the quoted duplicate delete commands may
// appear on uncommented lines.
func TestGenerateCommentInjectionNeutralized(t *testing.T) {
	o := model.New()
	o.Result["/x\nrm -rf INJECT\n#"] = &model.FileResult{
		Hash: "h\necho HASHINJECT", Size: 5, Duplicates: []string{"/y"},
	}
	s := report.FromOutput(o)

	for _, shell := range []struct {
		sh     script.Shell
		active string
	}{
		{script.Bash, "rm -f -- '"},
		{script.PowerShell, "Remove-Item -LiteralPath '"},
	} {
		var buf bytes.Buffer
		script.Generate(&buf, s, shell.sh)
		for _, line := range strings.Split(buf.String(), "\n") {
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if !strings.HasPrefix(line, shell.active) {
				t.Errorf("shell %v: uncommented active line from injected content: %q\nfull:\n%s",
					shell.sh, line, buf.String())
			}
			for _, bad := range []string{"rm -rf INJECT", "echo HASHINJECT"} {
				if strings.Contains(line, bad) {
					t.Errorf("shell %v: injected token %q leaked onto active line: %q", shell.sh, bad, line)
				}
			}
		}
	}
}
