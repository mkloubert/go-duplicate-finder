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

import "sort"

// SortMode selects the ordering of duplicate groups.
type SortMode int

const (
	SortReclaimable SortMode = iota // default
	SortSize
	SortDupCount
	SortPath
)

// sortModeCount is the number of modes, used to cycle.
const sortModeCount = 4

// String returns a short label for the header.
func (m SortMode) String() string {
	switch m {
	case SortReclaimable:
		return "reclaimable"
	case SortSize:
		return "size"
	case SortDupCount:
		return "duplicates"
	case SortPath:
		return "path"
	default:
		return "unknown"
	}
}

// Next returns the following mode, wrapping around.
func (m SortMode) Next() SortMode { return (m + 1) % sortModeCount }

// SortBy orders the groups in place. Every mode breaks ties by Original path
// ascending, so the result is deterministic.
func (s *Summary) SortBy(mode SortMode) {
	sort.SliceStable(s.Groups, func(i, j int) bool {
		a, b := s.Groups[i], s.Groups[j]
		switch mode {
		case SortReclaimable:
			if a.Reclaimable() != b.Reclaimable() {
				return a.Reclaimable() > b.Reclaimable()
			}
		case SortSize:
			if a.Size != b.Size {
				return a.Size > b.Size
			}
		case SortDupCount:
			if len(a.Duplicates) != len(b.Duplicates) {
				return len(a.Duplicates) > len(b.Duplicates)
			}
		case SortPath:
			// fall through to the path tie-break
		}
		return a.Original < b.Original
	})
}
