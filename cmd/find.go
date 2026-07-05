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
	"bytes"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sort"

	"github.com/mattn/go-isatty"
	"github.com/mkloubert/go-duplicate-finder/internal/dedup"
	"github.com/mkloubert/go-duplicate-finder/internal/hasher"
	"github.com/mkloubert/go-duplicate-finder/internal/highlight"
	"github.com/mkloubert/go-duplicate-finder/internal/htmlreport"
	"github.com/mkloubert/go-duplicate-finder/internal/model"
	"github.com/mkloubert/go-duplicate-finder/internal/report"
	"github.com/mkloubert/go-duplicate-finder/internal/scanner"
	"github.com/mkloubert/go-duplicate-finder/internal/ui"
	"github.com/spf13/cobra"
)

// errDuplicatesFound is returned by find when --fail-if-duplicates is set and at
// least one duplicate group remains. Execute maps it to exit code 2.
var errDuplicatesFound = errors.New("duplicates found")

func newFindCmd() *cobra.Command {
	var (
		output           string
		jobs             int
		noTUI            bool
		cwd              string
		compact          bool
		pretty           bool
		top              int
		minReclaimable   string
		format           string
		minCount         int
		failIfDuplicates bool
		hashName         string
		followSymlinks   bool
		quick            bool
	)

	cmd := &cobra.Command{
		Use:   "find [patterns...]",
		Short: "Find duplicate files by glob patterns",
		RunE: func(cmd *cobra.Command, args []string) error {
			patterns := args
			if len(patterns) == 0 {
				patterns = []string{"**/**"}
			}

			si, _ := cmd.Flags().GetBool("si")
			report.SetSIUnits(si)

			switch format {
			case "", "json", "html":
			default:
				return fmt.Errorf("invalid --format %q (use json or html)", format)
			}

			enabled, theme, err := resolveHighlight(cmd)
			if err != nil {
				return err
			}

			var minRecl int64
			if minReclaimable != "" {
				minRecl, err = report.ParseSize(minReclaimable)
				if err != nil {
					return fmt.Errorf("invalid --min-reclaimable: %w", err)
				}
			}

			algo, err := hasher.ParseAlgorithm(hashName)
			if err != nil {
				return err
			}

			if quick {
				fmt.Fprintln(os.Stderr, "warning: --quick uses approximate matching (sampled hashing, no byte comparison)")
			}

			baseDir, err := resolveBaseDir(cwd)
			if err != nil {
				return err
			}

			rep := ui.New(noTUI)

			files, err := scanner.Scan(baseDir, patterns, followSymlinks, rep)
			if err != nil {
				rep.Done()
				return err
			}

			out, err := dedup.Find(files, jobs, algo, quick, rep)
			rep.Done()
			if err != nil {
				return err
			}

			out = filterGroups(out, top, minRecl, minCount)

			if format == "html" {
				var buf bytes.Buffer
				if err := htmlreport.Write(&buf, report.FromOutput(out)); err != nil {
					return err
				}
				if _, err := os.Stdout.Write(buf.Bytes()); err != nil {
					return err
				}
				if output != "" {
					if err := os.WriteFile(output, buf.Bytes(), 0o644); err != nil {
						return fmt.Errorf("cannot write output file %q: %w", output, err)
					}
				}
			} else {
				indent := chooseIndent(compact, pretty, isatty.IsTerminal(os.Stdout.Fd()))
				var data []byte
				if indent {
					data, err = out.Marshal()
				} else {
					data, err = out.MarshalCompact()
				}
				if err != nil {
					return err
				}

				if err := highlight.Write(os.Stdout, string(data), "json", enabled, theme); err != nil {
					return err
				}
				fmt.Fprintln(os.Stdout)

				if output != "" {
					fileData, ferr := out.MarshalCompact()
					if ferr != nil {
						return ferr
					}
					if err := os.WriteFile(output, append(fileData, '\n'), 0o644); err != nil {
						return fmt.Errorf("cannot write output file %q: %w", output, err)
					}
				}
			}

			if failIfDuplicates && len(out.Result) > 0 {
				return errDuplicatesFound
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Also write the JSON to this file")
	cmd.Flags().IntVarP(&jobs, "jobs", "j", runtime.NumCPU(), "Number of parallel hash workers")
	cmd.Flags().BoolVar(&noTUI, "no-tui", false, "Disable the rich UI, plain logs only")
	cmd.Flags().StringVar(&cwd, "cwd", "", "Override the working directory (env: DUPFIND_CWD)")
	cmd.Flags().BoolVar(&compact, "compact", false, "Force compact single-line JSON")
	cmd.Flags().BoolVar(&pretty, "pretty", false, "Force indented JSON")
	cmd.MarkFlagsMutuallyExclusive("compact", "pretty")
	cmd.Flags().IntVar(&top, "top", 0, "Keep only the N groups with the most reclaimable space (0 = all)")
	cmd.Flags().StringVar(&minReclaimable, "min-reclaimable", "", "Keep only groups reclaiming at least this size (e.g. 10M)")
	cmd.Flags().StringVar(&format, "format", "", "Output format: json (default) or html")
	cmd.Flags().IntVar(&minCount, "min-count", 0, "Keep only groups with at least this many files (original + duplicates)")
	cmd.Flags().BoolVar(&failIfDuplicates, "fail-if-duplicates", false, "Exit with code 2 if any duplicate group remains")
	cmd.Flags().StringVar(&hashName, "hash", "", "Hash algorithm: blake3 (default), sha256, or xxh3")
	cmd.Flags().BoolVar(&followSymlinks, "follow-symlinks", false, "Follow symlinks and include their targets")
	cmd.Flags().BoolVar(&quick, "quick", false, "Approximate: sample-hash file ends, skip the byte comparison")
	return cmd
}

// filterGroups returns a copy of out keeping only groups whose reclaimable space
// (size × number of duplicates) is at least minReclaimable and whose file count
// (original + duplicates) is at least minCount, then only the top groups by
// reclaimable space (top <= 0 means no limit). Ties break by path so the
// selection is deterministic.
func filterGroups(out *model.Output, top int, minReclaimable int64, minCount int) *model.Output {
	type entry struct {
		key  string
		res  *model.FileResult
		recl int64
	}

	var entries []entry
	for key, res := range out.Result {
		if res == nil {
			continue
		}
		recl := res.Size * int64(len(res.Duplicates))
		fileCount := 1 + len(res.Duplicates)
		if recl >= minReclaimable && fileCount >= minCount {
			entries = append(entries, entry{key: key, res: res, recl: recl})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].recl != entries[j].recl {
			return entries[i].recl > entries[j].recl
		}
		return entries[i].key < entries[j].key
	})

	if top > 0 && len(entries) > top {
		entries = entries[:top]
	}

	result := model.New()
	for _, e := range entries {
		result.Result[e.key] = e.res
	}
	return result
}

// chooseIndent decides whether to indent the JSON. --compact forces compact,
// --pretty forces indented; otherwise it follows whether STDOUT is a terminal.
func chooseIndent(compact, pretty, isTTY bool) bool {
	if compact {
		return false
	}
	if pretty {
		return true
	}
	return isTTY
}
