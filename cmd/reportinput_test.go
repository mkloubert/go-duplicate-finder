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

func TestResolveReportInputFlagWins(t *testing.T) {
	dir := t.TempDir()
	flagFile := filepath.Join(dir, "flag.json")
	envFile := filepath.Join(dir, "env.json")
	os.WriteFile(flagFile, []byte("FLAG"), 0o644)
	os.WriteFile(envFile, []byte("ENV"), 0o644)
	t.Setenv(envReportFile, envFile)

	got, err := resolveReportInput(flagFile, strings.NewReader("STDIN"), false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "FLAG" {
		t.Fatalf("got %q, want FLAG", got)
	}
}

func TestResolveReportInputEnvWhenNoFlag(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, "env.json")
	os.WriteFile(envFile, []byte("ENV"), 0o644)
	t.Setenv(envReportFile, envFile)

	got, err := resolveReportInput("", strings.NewReader("STDIN"), false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "ENV" {
		t.Fatalf("got %q, want ENV", got)
	}
}

func TestResolveReportInputStdinWhenPiped(t *testing.T) {
	t.Setenv(envReportFile, "")
	got, err := resolveReportInput("", strings.NewReader("STDIN"), false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "STDIN" {
		t.Fatalf("got %q, want STDIN", got)
	}
}

func TestResolveReportInputErrorWhenNoInput(t *testing.T) {
	t.Setenv(envReportFile, "")
	if _, err := resolveReportInput("", strings.NewReader(""), true); err == nil {
		t.Fatal("expected error when no input and stdin is a TTY")
	}
}

func TestParseReportValidAndEmpty(t *testing.T) {
	out, err := parseReport([]byte(`{"result":{"/a":{"hash":"h","size":5,"duplicates":["/b"]}}}`))
	if err != nil {
		t.Fatal(err)
	}
	if out.Result["/a"].Size != 5 {
		t.Errorf("size = %d, want 5", out.Result["/a"].Size)
	}

	empty, err := parseReport([]byte(`{"result":{}}`))
	if err != nil {
		t.Fatalf("empty report should not error: %v", err)
	}
	if empty.Result == nil || len(empty.Result) != 0 {
		t.Errorf("empty result = %+v", empty.Result)
	}
}

func TestParseReportInvalidJSON(t *testing.T) {
	if _, err := parseReport([]byte("{not json")); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
