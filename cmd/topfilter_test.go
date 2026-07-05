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

	"github.com/mkloubert/go-duplicate-finder/internal/model"
)

func TestFilterGroups(t *testing.T) {
	out := model.New()
	out.Result["/a"] = &model.FileResult{Size: 100, Duplicates: []string{"/a2", "/a3"}} // recl 200
	out.Result["/b"] = &model.FileResult{Size: 500, Duplicates: []string{"/b2"}}        // recl 500
	out.Result["/c"] = &model.FileResult{Size: 10, Duplicates: []string{"/c2"}}         // recl 10

	if got := filterGroups(out, 0, 0); len(got.Result) != 3 {
		t.Errorf("no filter kept %d groups, want 3", len(got.Result))
	}

	got := filterGroups(out, 0, 100) // min reclaimable drops /c (recl 10)
	if len(got.Result) != 2 || got.Result["/c"] != nil {
		t.Errorf("min-reclaimable filter kept %v", mapKeys(got))
	}

	got = filterGroups(out, 1, 0) // top 1 -> only /b (largest reclaimable)
	if len(got.Result) != 1 || got.Result["/b"] == nil {
		t.Errorf("top 1 kept %v, want [/b]", mapKeys(got))
	}

	got = filterGroups(out, 2, 0) // top 2 -> /b and /a
	if len(got.Result) != 2 || got.Result["/b"] == nil || got.Result["/a"] == nil {
		t.Errorf("top 2 kept %v, want [/b /a]", mapKeys(got))
	}
}

func mapKeys(o *model.Output) []string {
	var ks []string
	for k := range o.Result {
		ks = append(ks, k)
	}
	return ks
}

func makeSizedDupTree(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	big := strings.Repeat("A", 100)  // group A, reclaimable 100
	small := strings.Repeat("B", 10) // group B, reclaimable 10
	files := map[string]string{
		"a1.bin": big, "a2.bin": big,
		"b1.bin": small, "b2.bin": small,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func findResultKeys(t *testing.T, out string) map[string]struct{} {
	t.Helper()
	var parsed struct {
		Result map[string]json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, out)
	}
	keys := map[string]struct{}{}
	for k := range parsed.Result {
		keys[k] = struct{}{}
	}
	return keys
}

func TestFindTopKeepsLargest(t *testing.T) {
	dir := makeSizedDupTree(t)
	out, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--compact", "--top", "1", "**/**")
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	keys := findResultKeys(t, out)
	if len(keys) != 1 {
		t.Fatalf("--top 1 kept %d groups, want 1: %v", len(keys), keys)
	}
	if _, ok := keys[filepath.Join(dir, "a1.bin")]; !ok {
		t.Errorf("--top 1 should keep the bigger group (a1.bin): %v", keys)
	}
}

func TestFindMinReclaimableFilters(t *testing.T) {
	dir := makeSizedDupTree(t)
	out, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--compact", "--min-reclaimable", "50", "**/**")
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	keys := findResultKeys(t, out)
	if len(keys) != 1 {
		t.Fatalf("--min-reclaimable 50 kept %d groups, want 1: %v", len(keys), keys)
	}
}

func TestFindInvalidMinReclaimable(t *testing.T) {
	dir := makeSizedDupTree(t)
	if _, err := captureScript(t, "find", "--cwd", dir, "--no-tui", "--min-reclaimable", "bogus", "**/**"); err == nil {
		t.Fatal("expected error for invalid --min-reclaimable")
	}
}

func writeSizedReport(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	rf := filepath.Join(dir, "report.json")
	if err := os.WriteFile(rf, []byte(`{"result":{"/x/first":{"hash":"h","size":1000000,"duplicates":["/x/second"]}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	return rf
}

func TestSummaryDefaultUsesIECUnits(t *testing.T) {
	rf := writeSizedReport(t)
	out, err := captureScript(t, "summary", "--no-tui", "-f", rf)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if !strings.Contains(out, "976.6 KB") {
		t.Errorf("default should use 1024-based units (976.6 KB):\n%s", out)
	}
}

func TestSummarySIUsesSIUnits(t *testing.T) {
	rf := writeSizedReport(t)
	out, err := captureScript(t, "summary", "--no-tui", "-f", rf, "--si")
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if !strings.Contains(out, "1.0 MB") {
		t.Errorf("--si should use 1000-based units (1.0 MB):\n%s", out)
	}
}
