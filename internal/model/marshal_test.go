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

package model_test

import (
	"strings"
	"testing"

	"github.com/mkloubert/go-duplicate-finder/internal/model"
)

func TestOutputMarshal(t *testing.T) {
	o := model.New()
	o.Result["/a/first"] = &model.FileResult{
		Hash:       "abc123",
		Size:       42,
		Duplicates: []string{"/a/second", "/a/third"},
	}

	got, err := o.Marshal()
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	want := `{
  "result": {
    "/a/first": {
      "hash": "abc123",
      "size": 42,
      "duplicates": [
        "/a/second",
        "/a/third"
      ]
    }
  }
}`
	if string(got) != want {
		t.Fatalf("unexpected JSON:\n got:\n%s\nwant:\n%s", got, want)
	}
}

func TestEmptyOutputMarshal(t *testing.T) {
	got, err := model.New().Marshal()
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	want := `{
  "result": {}
}`
	if string(got) != want {
		t.Fatalf("unexpected empty JSON:\n got:\n%s\nwant:\n%s", got, want)
	}
}

func TestOutputMarshalCompact(t *testing.T) {
	o := model.New()
	o.Result["/a/first"] = &model.FileResult{
		Hash:       "abc123",
		Size:       42,
		Duplicates: []string{"/a/second"},
	}

	got, err := o.MarshalCompact()
	if err != nil {
		t.Fatalf("MarshalCompact returned error: %v", err)
	}

	want := `{"result":{"/a/first":{"hash":"abc123","size":42,"duplicates":["/a/second"]}}}`
	if string(got) != want {
		t.Fatalf("compact JSON:\n got: %s\nwant: %s", got, want)
	}
	if strings.Contains(string(got), "\n") {
		t.Error("compact JSON must not contain newlines")
	}
}
