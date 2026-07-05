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

import "strings"

// Filter returns the groups whose original path or any duplicate path contains
// the query (case-insensitive). An empty query returns all groups.
func (s Summary) Filter(query string) []Group {
	if query == "" {
		return s.Groups
	}
	q := strings.ToLower(query)
	var out []Group
	for _, g := range s.Groups {
		if strings.Contains(strings.ToLower(g.Original), q) {
			out = append(out, g)
			continue
		}
		for _, d := range g.Duplicates {
			if strings.Contains(strings.ToLower(d), q) {
				out = append(out, g)
				break
			}
		}
	}
	return out
}
