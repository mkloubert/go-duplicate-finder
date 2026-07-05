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

package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestChooseIndent(t *testing.T) {
	cases := []struct {
		compact, pretty, isTTY, want bool
	}{
		{false, false, true, true},   // TTY default -> indent
		{false, false, false, false}, // pipe default -> compact
		{true, false, true, false},   // --compact wins even on TTY
		{true, false, false, false},  // --compact when piped -> compact
		{false, true, false, true},   // --pretty wins even when piped
		{false, true, true, true},    // --pretty on TTY -> indent
	}
	for _, c := range cases {
		if got := chooseIndent(c.compact, c.pretty, c.isTTY); got != c.want {
			t.Errorf("chooseIndent(%v,%v,%v) = %v, want %v", c.compact, c.pretty, c.isTTY, got, c.want)
		}
	}
}

func TestFindPipedIsCompact(t *testing.T) {
	dir := makeDupTree(t)
	out, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "**/**")
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if strings.Contains(strings.TrimRight(out, "\n"), "\n") {
		t.Errorf("piped JSON should be compact (single line):\n%q", out)
	}
	if !strings.Contains(out, `"duplicates"`) {
		t.Errorf("output should be JSON:\n%s", out)
	}
}

func TestFindColorAlwaysPipedStaysCompact(t *testing.T) {
	// Layout is independent of --color: forcing color on a pipe must not make
	// the JSON indented.
	dir := makeDupTree(t)
	out, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--color", "always", "**/**")
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if strings.Contains(strings.TrimRight(out, "\n"), "\n") {
		t.Errorf("--color always piped must still be compact (single line):\n%q", out)
	}
}

func TestFindPrettyForcesIndent(t *testing.T) {
	dir := makeDupTree(t)
	out, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--pretty", "**/**")
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if !strings.Contains(out, "\n  ") {
		t.Errorf("--pretty should produce indented JSON:\n%q", out)
	}
}

func TestFindCompactFlag(t *testing.T) {
	dir := makeDupTree(t)
	out, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--compact", "**/**")
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if strings.Contains(strings.TrimRight(out, "\n"), "\n") {
		t.Errorf("--compact should be single line:\n%q", out)
	}
}

func TestFindOutputFileAlwaysCompact(t *testing.T) {
	dir := makeDupTree(t)
	outFile := filepath.Join(dir, "out.json")
	// Even with --pretty, the file must stay compact.
	if _, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--pretty", "-o", outFile, "**/**"); err != nil {
		t.Fatalf("execute error: %v", err)
	}
	raw, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.TrimRight(string(raw), "\n"), "\n") {
		t.Errorf("--output file must be compact even with --pretty:\n%s", raw)
	}
}

func TestFindCompactPrettyMutuallyExclusive(t *testing.T) {
	dir := makeDupTree(t)
	if _, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--compact", "--pretty", "**/**"); err == nil {
		t.Fatal("expected error when both --compact and --pretty are set")
	}
}
