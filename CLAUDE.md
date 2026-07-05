# About this project

`dupfind` is a Go CLI tool that finds duplicate files. It has three commands:

- **`find`** — scan glob patterns, detect duplicates, emit a JSON (or HTML) report.
- **`summary`** — read a JSON report and show it in a rich terminal UI.
- **`script`** — read a JSON report and generate a bash/zsh/PowerShell delete script.

Binary name: `dupfind`. Module path: `github.com/mkloubert/go-duplicate-finder`.

---

## Hard rules (never break these)

### 1. Copyright header

Every source file we own (including `_test.go` files) MUST begin with this exact
header, before `package`:

```
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
```

### 2. Language: English only

All code, comments, identifiers, user-facing strings, error messages, tests, and
documentation are in English (Simple English where possible). No German in the
codebase.

### 3. No git

This repository is not managed with git. Do NOT run any `git` command and do NOT
commit, branch, or push — unless the user explicitly asks. Verify work by running
tests, not by inspecting diffs/history.

### 4. Stream separation is sacred

STDOUT carries ONLY the machine-consumable payload (the JSON/HTML from `find`,
the script from `script`). Every status line, progress indicator, TUI frame,
warning, and error goes to STDERR. Never write anything else to STDOUT — a stray
byte breaks piping. `summary` is a UI, so its rendered view is its payload on
STDOUT.

---

## Tech stack

Go (toolchain 1.26; `min`/`max` builtins are available). Direct dependencies:

- `spf13/cobra` — CLI framework.
- `bmatcuk/doublestar/v4` — glob patterns supporting `**` (stdlib `filepath.Glob` does not).
- `lukechampine.com/blake3` — default content hash (fast, secure).
- `zeebo/xxh3` — optional faster hash (`--hash xxh3`); `crypto/sha256` (stdlib) is the third option.
- `charmbracelet/bubbletea` (v1), `bubbles`, `lipgloss` (v1) — terminal UI.
- `mattn/go-isatty` — TTY detection.
- `alecthomas/chroma/v2` — terminal syntax highlighting.
- `atotto/clipboard` — OS clipboard (only the `system` clipboard backend).
- `html/template` (stdlib) — the HTML report (`--format html`), auto-escaped.

Run `go mod tidy` after adding a dependency so it becomes a direct requirement.

---

## Architecture

`cmd/` is thin Cobra wiring; `internal/*` packages hold pure, isolated,
independently testable logic. The command layer resolves inputs/flags and calls
into `internal`. Keep files focused and small.

```
main.go                      -> cmd.Execute()
cmd/root.go                  root command; persistent --color/--theme/--si; Execute() maps errDuplicatesFound -> exit 2, other errors -> STDERR + exit 1
cmd/find.go                  find command; filterGroups(), chooseIndent(); errDuplicatesFound sentinel
cmd/summary.go               summary command (interactive vs static, /dev/tty)
cmd/script.go                script command + shellLang()
cmd/cwd.go                   resolveBaseDir()   (--cwd)
cmd/reportinput.go           resolveReportInput(), parseReport()   (-f / stdin)
cmd/color.go                 resolveHighlight(), resolveTheme()
cmd/env.go                   env-var helpers (applyString/Bool/Int/StringSliceEnv, resolveSI)
internal/model               Output/FileResult; Marshal() (indented), MarshalCompact() (single line)
internal/scanner             Scan(baseDir, Options, rep): glob -> filtered []FileEntry (size bounds, exclude globs, hardlink/symlink inode dedup)
internal/scanner/inode_*.go  build-tagged (dev,inode) extraction: inode_unix.go (//go:build unix), inode_other.go (fallback, no dedup)
internal/hasher              Algorithm (blake3/sha256/xxh3); HashFile(path, algo), HashSample(path, algo, n)
internal/dedup               Find(files, jobs, algo, quick, rep): size-group -> concurrent hash -> byte-compare (or sample in quick) -> clusters
internal/report              report JSON -> Summary/Group/Totals; SortBy, Filter, Humanize/HumanizeBytes, ParseSize, SetSIUnits
internal/htmlreport          Write(w, Summary): self-contained HTML page (html/template, auto-escaped)
internal/ui                  Reporter interface (Noop/Plain/TUI) for find progress
internal/summaryui           static + interactive (bubbletea) renderers, clipboard backends
internal/script              delete-script generation + shell-safe quoting + commentSafe()
internal/highlight           chroma-based ColorMode/Enabled/Write
```

---

## Commands, flags & environment

Persistent (all commands): `--color auto|always|never` (default `auto`),
`--theme <name>` (default `monokai`), `--si` (1000-based size units).

- **`find [patterns...]`** — default pattern `**/**`. Flags:
  - output: `-o/--output <file>`, `--format json|html`, `--compact`, `--pretty`
    (`--compact`/`--pretty` are mutually exclusive).
  - scan scope: `--cwd <dir>`, `--exclude <glob>` (repeatable), `--min-size`,
    `--max-size`, `--follow-symlinks`.
  - report filter: `--min-count <n>`, `--min-reclaimable <size>`, `--top <n>`.
  - matching/speed: `--hash blake3|sha256|xxh3`, `--quick`, `-j/--jobs <n>`.
  - other: `--no-tui`, `--fail-if-duplicates` (exit 2 if any group remains).
- **`summary`** — `-f/--report-file <file>`, `--no-tui`, `--clipboard osc52|system`.
- **`script`** — `-f/--report-file <file>`, `--shell auto|bash|zsh|powershell`.

Sizes (`--min-size`/`--max-size`/`--min-reclaimable`) accept human values like
`10M`, `1.5G` (1024-based) via `report.ParseSize`.

Environment variables (each is a default; the flag always wins):
`DUPFIND_CWD`, `DUPFIND_REPORT_FILE`, `DUPFIND_SHELL`, `DUPFIND_THEME`,
`DUPFIND_CLIPBOARD`, `DUPFIND_COLOR`, `DUPFIND_SI`, `DUPFIND_JOBS`,
`DUPFIND_NO_TUI`, `DUPFIND_HASH`, `DUPFIND_FORMAT`, `DUPFIND_EXCLUDE`
(comma-separated), `DUPFIND_FOLLOW_SYMLINKS`, and `NO_COLOR`. Per-invocation
flags (`-o`, `--top`, `--min-*`, `--max-size`, `--compact/--pretty`,
`--fail-if-duplicates`, `--quick`) are intentionally NOT env-backed.

### The JSON report shape

```json
{ "result": { "/abs/first": { "hash": "<content-hash-hex>", "size": 123, "duplicates": ["/abs/second", "..."] } } }
```

Only groups that actually have duplicates appear. The map key is the first
occurrence (lexicographically smallest path); `duplicates` are sorted.

---

## Cross-cutting patterns (use these; keep them consistent)

- **Config resolution precedence = flag > env var > default/auto.** Decisions
  with their own logic live in a small, pure, testable resolver (`resolveBaseDir`
  for `--cwd`, `resolveReportInput` for `-f`/stdin, `resolveShell`/`DetectFrom`
  for `--shell`, `resolveTheme` for `--theme`). Flags without special logic use
  the `cmd/env.go` helpers, which apply the env value only when
  `cmd.Flags().Changed(name)` is false. Make OS/TTY/env inputs parameters of a
  pure helper (`DetectFrom`, `chooseIndent`, `Enabled`, `ParseSize`) so the logic
  is unit-testable without a real terminal.

- **TTY-driven presentation.** Detect with `isatty.IsTerminal(os.Std{out,err}.Fd())`.
  - Color/highlight: `--color auto` colorizes only when STDOUT is a TTY and
    `NO_COLOR` is unset; `always` overrides `NO_COLOR`; `never` disables.
  - JSON layout: indented on a TTY, compact off-console; `--compact`/`--pretty`
    override. This is INDEPENDENT of `--color`.
  - find/summary UI: rich Bubble Tea UI on a TTY (rendered to STDERR for find,
    STDOUT for summary), plain fallback otherwise or with `--no-tui`.
  - When a report is read from STDIN, the interactive summary reads keys from
    `/dev/tty` (falls back to static if it cannot be opened).

- **Pipe safety (highest priority).** ANSI/coloring must never reach a pipe or
  file. `highlight.Write` formats into a buffer and writes only on success,
  falling back to the raw text on any error — so highlighting can never corrupt
  or truncate output. `find --output <file>` is ALWAYS raw and ALWAYS compact,
  independent of STDOUT.

- **Scan filters & hardlink correctness (`internal/scanner`).** `Scan` takes an
  `Options` struct (patterns, follow-symlinks, min/max size, exclude globs).
  Files that share a `(device, inode)` collapse to one entry — hardlinks (and,
  when following, a symlink and its target) are the same physical file and must
  NOT be reported as reclaimable duplicates. Inode extraction is build-tagged:
  `inode_unix.go` (`//go:build unix`) via `syscall.Stat_t`; `inode_other.go`
  returns `ok=false` (no inode dedup off Unix).

- **Selectable hash + quick mode (`internal/hasher`, `internal/dedup`).**
  `--hash` picks BLAKE3 (default), SHA256, or XXH3; downstream code treats the
  digest as an opaque hex string. `--quick` sample-hashes the file ends and skips
  the byte comparison — faster but approximate, so it warns on STDERR and stays
  opt-in (never env-backed, to keep the default exact).

- **Exit codes.** Real errors → STDERR + exit 1. `--fail-if-duplicates` returns
  the `errDuplicatesFound` sentinel, which `Execute` maps to a quiet exit 2 (a
  result signal, not an error); the report is still written first.

- **Shell-script generation is security-critical.** In `internal/script`:
  - Quote every path: POSIX `'...'` with `'`→`'\''` plus `--`; PowerShell
    `'...'` with `'`→`''` plus `-LiteralPath`.
  - `commentSafe()` every value interpolated into a `#` comment line (paths may
    contain newlines → an unsanitized newline escapes the comment and becomes an
    active command). This bug was found and fixed once; do not regress it.
  - The first occurrence is emitted commented out; the duplicates are active
    delete commands (so running the script deletes duplicates, keeps the first).

- **HTML report safety (`internal/htmlreport`).** Built with `html/template`, so
  interpolated paths are auto-escaped — a path containing markup cannot inject
  HTML. Same "untrusted paths are dangerous" mindset as the shell quoting.

- **Determinism.** Map iteration order is random — always sort (by path / the
  chosen key) before producing output. `encoding/json` already sorts map keys.
  Every `report` sort mode breaks ties by original path ascending; `filterGroups`
  (top-N) ranks by reclaimable space with a path tie-break.

- **Reporter decoupling.** Core packages (`scanner`, `dedup`) take a
  `ui.Reporter` interface, never a concrete TUI, so they run and test without a
  terminal. Use `ui.Noop{}` in tests.

- **lipgloss color correctness.** Bind the renderer to the actual output writer
  (`lipgloss.NewRenderer(w)`) so `NO_COLOR` and non-TTY targets are honored for
  that writer, not for the process stdout.

---

## Testing conventions

- TDD: write the failing test first, then the minimal implementation.
- Prefer table tests and black-box tests (`package foo_test`) where practical.
- Command end-to-end tests capture STDOUT by swapping `os.Stdout` for an
  `os.Pipe` (`captureScript` helper); use `t.TempDir()` for filesystem fixtures
  and `t.Setenv` for env vars (also confirms flag > env precedence).
- Run the race detector on concurrent code: `go test -race ./internal/dedup/`.
- Assert the invariants that matter here: stream separation, pipe-safety (0 ANSI
  when piped; raw `--output` file), compact-when-piped, shell-injection
  neutralization, hardlink collapse, and quick-vs-exact behavior.
- No hollow tests (a test must assert real behavior).
- Before considering any change done, ALL of these must be clean:
  `gofmt -l .` (empty output), `go vet ./...`, `go test ./...`.

---

## Gotchas / lessons learned

- Copyright header goes on EVERY `.go` file, tests included.
- Scanner inode extraction is build-tagged (`inode_unix.go` / `inode_other.go`);
  keep both in sync when changing `fileID`'s signature.
- `report.SetSIUnits` sets a package-global default read by `Humanize`; every
  command sets it once at RunE start (from `--si`/`DUPFIND_SI`) so it never leaks
  between runs. The pure `HumanizeBytes(bytes, si)` is the tested core.
- Hash digest length depends on the algorithm: BLAKE3-256 and SHA256 are 64 hex
  chars, XXH3 is 16 — tests must not hard-code 64.
- `internal/summaryui` defines a type named `model`; a white-box test there that
  also imports `internal/model` must alias the import (e.g. `mdl`).
- Bubble Tea v1 has no native `SetClipboard`; clipboard is OSC 52 by default or
  `atotto/clipboard` for the `system` backend.
- `internal/report.Group.Hash` exists so the `script` group header can show the
  hash; it is populated in `FromOutput`.
- New external dependencies show as `// indirect` until `go mod tidy` runs.

---

## Build & run

```
go build ./...
go run . find '**/**'
go run . find '**/**' | go run . summary
go run . find '**/**' | go run . script --shell bash
go run . find '**/**' --format html > report.html
go test ./...
```
