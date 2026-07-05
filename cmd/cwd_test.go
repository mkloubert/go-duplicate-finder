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
	"testing"
)

func TestResolveBaseDirFlagWinsOverEnvAndCwd(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(envCwd, "/this/env/value/must/be/ignored")

	got, err := resolveBaseDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != dir {
		t.Fatalf("flag should win: got %q, want %q", got, dir)
	}
}

func TestResolveBaseDirUsesEnvWhenNoFlag(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(envCwd, dir)

	got, err := resolveBaseDir("")
	if err != nil {
		t.Fatal(err)
	}
	if got != dir {
		t.Fatalf("env should be used: got %q, want %q", got, dir)
	}
}

func TestResolveBaseDirFallsBackToWorkingDir(t *testing.T) {
	t.Setenv(envCwd, "")

	got, err := resolveBaseDir("")
	if err != nil {
		t.Fatal(err)
	}
	wd, _ := os.Getwd()
	if got != wd {
		t.Fatalf("should fall back to Getwd: got %q, want %q", got, wd)
	}
}

func TestResolveBaseDirMakesRelativeAbsolute(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	defer os.Chdir(old)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	got, err := resolveBaseDir("sub")
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(got) {
		t.Fatalf("expected absolute path, got %q", got)
	}
	if got != filepath.Join(dir, "sub") {
		t.Fatalf("got %q, want %q", got, filepath.Join(dir, "sub"))
	}
}

func TestResolveBaseDirErrorsOnMissingDir(t *testing.T) {
	if _, err := resolveBaseDir(filepath.Join(t.TempDir(), "does-not-exist")); err == nil {
		t.Fatal("expected error for missing directory")
	}
}

func TestResolveBaseDirErrorsOnFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "a-file.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := resolveBaseDir(f); err == nil {
		t.Fatal("expected error when path is a file, not a directory")
	}
}
