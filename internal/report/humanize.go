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

package report

import "fmt"

// siUnits is the package-wide default for Humanize. It is set once at startup
// via SetSIUnits and read single-threaded, so it needs no synchronization.
var siUnits bool

// SetSIUnits selects 1000-based (SI) units for Humanize. Call once during
// startup; do not modify concurrently.
func SetSIUnits(v bool) { siUnits = v }

// Humanize formats a byte count using the package unit setting (1024-based by
// default; 1000-based when SetSIUnits(true) was called).
func Humanize(bytes int64) string { return HumanizeBytes(bytes, siUnits) }

// HumanizeBytes formats a byte count. With si=false it uses 1024-based units
// (KB, MB, …); with si=true it uses 1000-based SI units (kB, MB, …).
func HumanizeBytes(bytes int64, si bool) string {
	base := int64(1024)
	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	if si {
		base = 1000
		units = []string{"kB", "MB", "GB", "TB", "PB", "EB"}
	}
	if bytes < base {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := base, 0
	for n := bytes / base; n >= base; n /= base {
		div *= base
		exp++
	}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}
