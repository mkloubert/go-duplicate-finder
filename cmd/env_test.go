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
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func firstHash(t *testing.T, out string) string {
	t.Helper()
	var parsed struct {
		Result map[string]struct {
			Hash string `json:"hash"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("stdout not JSON: %v\n%s", err, out)
	}
	for _, v := range parsed.Result {
		return v.Hash
	}
	t.Fatalf("no result in %s", out)
	return ""
}

func TestEnvHash(t *testing.T) {
	dir := makeSizedDupTree(t)
	t.Setenv("DUPFIND_HASH", "xxh3")

	out, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--compact", "**/**")
	if err != nil {
		t.Fatal(err)
	}
	if h := firstHash(t, out); len(h) != 16 {
		t.Errorf("DUPFIND_HASH=xxh3 should give a 16-char hash, got %d (%q)", len(h), h)
	}

	// The flag overrides the env.
	out, err = captureScript(t, "find", "--cwd", dir, "--no-tui", "--compact", "--hash", "blake3", "**/**")
	if err != nil {
		t.Fatal(err)
	}
	if h := firstHash(t, out); len(h) != 64 {
		t.Errorf("--hash blake3 should override env (64-char hash), got %d", len(h))
	}
}

func TestEnvFormat(t *testing.T) {
	dir := makeSizedDupTree(t)
	t.Setenv("DUPFIND_FORMAT", "html")

	out, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "**/**")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "<!DOCTYPE html>") {
		t.Errorf("DUPFIND_FORMAT=html should render HTML:\n%s", out)
	}

	// The flag overrides the env.
	out, err = captureScript(t, "find", "--cwd", dir, "--no-tui", "--compact", "--format", "json", "**/**")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "<!DOCTYPE") {
		t.Errorf("--format json should override env:\n%s", out)
	}
}

func TestEnvExclude(t *testing.T) {
	dir := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("keep1.txt", "X")
	write("keep2.txt", "X")
	write(".git/g1", "Y")
	write(".git/g2", "Y")

	t.Setenv("DUPFIND_EXCLUDE", ".git/**")
	out, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--compact", "**/**")
	if err != nil {
		t.Fatal(err)
	}
	keys := findResultKeys(t, out)
	if len(keys) != 1 {
		t.Fatalf("DUPFIND_EXCLUDE should leave 1 group, got %d: %v", len(keys), keys)
	}
	for k := range keys {
		if strings.Contains(k, ".git") {
			t.Errorf("excluded path present: %s", k)
		}
	}
}

func TestEnvColorAlways(t *testing.T) {
	dir := makeSizedDupTree(t)
	t.Setenv("DUPFIND_COLOR", "always")

	out, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "**/**")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "\x1b[") {
		t.Errorf("DUPFIND_COLOR=always should colorize piped output:\n%q", out)
	}

	// The flag overrides the env.
	out, err = captureScript(t, "find", "--cwd", dir, "--no-tui", "--color", "never", "**/**")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "\x1b[") {
		t.Errorf("--color never should override env:\n%q", out)
	}
}

func TestEnvSISummary(t *testing.T) {
	rf := writeSizedReport(t) // 1,000,000-byte file
	t.Setenv("DUPFIND_SI", "true")
	out, err := captureScript(t, "summary", "--no-tui", "-f", rf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "1.0 MB") {
		t.Errorf("DUPFIND_SI=true should use SI units (1.0 MB):\n%s", out)
	}
}

func TestEnvInvalidJobs(t *testing.T) {
	dir := makeSizedDupTree(t)
	t.Setenv("DUPFIND_JOBS", "notanumber")
	if _, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "**/**"); err == nil {
		t.Fatal("invalid DUPFIND_JOBS should error")
	}
}

func TestEnvInvalidSI(t *testing.T) {
	rf := writeSizedReport(t)
	t.Setenv("DUPFIND_SI", "maybe")
	if _, err := captureScript(t, "summary", "--no-tui", "-f", rf); err == nil {
		t.Fatal("invalid DUPFIND_SI should error")
	}
}
