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

package report_test

import (
	"testing"

	"github.com/mkloubert/go-duplicate-finder/internal/report"
)

func TestParseSizeValid(t *testing.T) {
	cases := map[string]int64{
		"0":      0,
		"1024":   1024,
		"10B":    10,
		"1K":     1024,
		"1KB":    1024,
		"1KiB":   1024,
		"1.5M":   1572864,
		"2G":     2147483648,
		"1T":     1099511627776,
		"  4M  ": 4194304,
		"1.5mib": 1572864,
	}
	for in, want := range cases {
		got, err := report.ParseSize(in)
		if err != nil {
			t.Errorf("ParseSize(%q) unexpected error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("ParseSize(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestParseSizeInvalid(t *testing.T) {
	for _, bad := range []string{"", "  ", "abc", "M", "KB", "-5M", "1.2.3"} {
		if _, err := report.ParseSize(bad); err == nil {
			t.Errorf("ParseSize(%q) should return an error", bad)
		}
	}
}

func TestHumanizeBytes(t *testing.T) {
	// IEC (1024-based) — the default.
	iec := map[int64]string{
		0:          "0 B",
		512:        "512 B",
		1024:       "1.0 KB",
		1000000:    "976.6 KB",
		2147483648: "2.0 GB",
	}
	for in, want := range iec {
		if got := report.HumanizeBytes(in, false); got != want {
			t.Errorf("HumanizeBytes(%d, false) = %q, want %q", in, got, want)
		}
	}
	// SI (1000-based).
	si := map[int64]string{
		500:     "500 B",
		1000:    "1.0 kB",
		1000000: "1.0 MB",
		1500000: "1.5 MB",
	}
	for in, want := range si {
		if got := report.HumanizeBytes(in, true); got != want {
			t.Errorf("HumanizeBytes(%d, true) = %q, want %q", in, got, want)
		}
	}
}

func TestSetSIUnitsAffectsHumanize(t *testing.T) {
	defer report.SetSIUnits(false) // restore default for other tests
	report.SetSIUnits(true)
	if got := report.Humanize(1000000); got != "1.0 MB" {
		t.Errorf("Humanize with SI = %q, want 1.0 MB", got)
	}
	report.SetSIUnits(false)
	if got := report.Humanize(1000000); got != "976.6 KB" {
		t.Errorf("Humanize without SI = %q, want 976.6 KB", got)
	}
}
