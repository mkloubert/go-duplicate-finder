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

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseSize parses a human-readable size like "1024", "10K", "1.5M", "2G" into
// a number of bytes. Units are 1024-based and case-insensitive; an optional
// trailing "B" or "iB" is allowed (e.g. "10MB", "10MiB"). A bare number is
// interpreted as bytes.
func ParseSize(s string) (int64, error) {
	in := strings.TrimSpace(s)
	if in == "" {
		return 0, fmt.Errorf("empty size")
	}

	// Normalize: uppercase, then drop an optional trailing "B" and binary "I"
	// so "MB"/"MIB" both reduce to the unit letter "M".
	up := strings.ToUpper(in)
	up = strings.TrimSuffix(up, "B")
	up = strings.TrimSuffix(up, "I")

	mult := int64(1)
	if len(up) > 0 {
		switch up[len(up)-1] {
		case 'K':
			mult = 1 << 10
		case 'M':
			mult = 1 << 20
		case 'G':
			mult = 1 << 30
		case 'T':
			mult = 1 << 40
		case 'P':
			mult = 1 << 50
		case 'E':
			mult = 1 << 60
		}
	}

	numStr := up
	if mult != 1 {
		numStr = up[:len(up)-1]
	}
	numStr = strings.TrimSpace(numStr)
	if numStr == "" {
		return 0, fmt.Errorf("invalid size %q", s)
	}

	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q", s)
	}
	if val < 0 {
		return 0, fmt.Errorf("negative size %q", s)
	}
	return int64(val * float64(mult)), nil
}
