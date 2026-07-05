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
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestFindEndToEnd(t *testing.T) {
	dir := t.TempDir()
	for name, content := range map[string]string{
		"a.txt": "hello",
		"b.txt": "hello",
		"c.txt": "world",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	// Capture STDOUT.
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	outFile := filepath.Join(dir, "out.json")
	root := newRootCmd()
	root.SetArgs([]string{"find", "--no-tui", "-o", outFile, "**/**"})
	runErr := root.Execute()

	w.Close()
	os.Stdout = oldStdout
	raw, _ := io.ReadAll(r)

	if runErr != nil {
		t.Fatalf("Execute error: %v", runErr)
	}

	var parsed struct {
		Result map[string]struct {
			Hash       string   `json:"hash"`
			Size       int64    `json:"size"`
			Duplicates []string `json:"duplicates"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("STDOUT is not valid JSON: %v\n%s", err, raw)
	}
	if len(parsed.Result) != 1 {
		t.Fatalf("expected 1 duplicate cluster, got %d: %s", len(parsed.Result), raw)
	}
	// a.txt/b.txt are duplicates; c.txt is not.
	key := filepath.Join(dir, "a.txt")
	res, ok := parsed.Result[key]
	if !ok {
		t.Fatalf("key %q missing: %s", key, raw)
	}
	if len(res.Duplicates) != 1 || res.Duplicates[0] != filepath.Join(dir, "b.txt") {
		t.Errorf("expected b.txt as duplicate, got %v", res.Duplicates)
	}

	// --output must contain the same structure.
	fileRaw, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("output file not readable: %v", err)
	}
	if err := json.Unmarshal(fileRaw, &parsed); err != nil {
		t.Fatalf("output file is not valid JSON: %v", err)
	}
}
