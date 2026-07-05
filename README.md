# dupfind

`dupfind` is a command-line tool that finds duplicate files.

It works in three steps, one command each:

1. **`find`** — scan folders and write a JSON report of the duplicates.
2. **`summary`** — read that report and show it in a nice terminal view.
3. **`script`** — read that report and create a shell script that deletes the
   duplicates.

Everything is piped together, so you can mix and match the steps.

## How it finds duplicates

Files are duplicates only when they are truly identical. `dupfind` is both fast
and safe:

1. Group files by size (different sizes cannot be duplicates).
2. Hash the same-size files with BLAKE3 (a fast, secure hash). You can choose
   `sha256` or `xxh3` instead with `--hash`.
3. Compare matching files byte by byte to be 100% sure.

Empty (0-byte) files are skipped. Hardlinks (and, with `--follow-symlinks`,
symlink targets) are treated as the same file, so they are never counted as
duplicates. For very large trees, `--quick` samples only the file ends and skips
the byte comparison — faster, but approximate.

## Install

You need Go installed. Build the binary:

```sh
go build -o dupfind .
```

Then run `./dupfind`, or move it into a folder on your `PATH`.

## Quick start

```sh
# Find duplicates in the current folder and show them:
dupfind find | dupfind summary

# Save a report, then make a delete script from it:
dupfind find > report.json
dupfind script -f report.json > clean.sh
# read clean.sh, then run it when you are happy:
bash clean.sh

# A browsable HTML report:
dupfind find '**/**' --format html > report.html

# In CI: fail when duplicates are found (exit code 2):
dupfind find '**/**' --fail-if-duplicates
```

## Commands

### `find [patterns...]`

Scans one or more glob patterns and prints a JSON report to STDOUT. The default
pattern is `**/**` (everything, recursively).

```sh
dupfind find '**/*.jpg' '**/*.png'
dupfind find --cwd /photos '**/**' -o report.json
dupfind find '**/**' --min-size 10M --exclude '**/node_modules/**' --top 20
dupfind find '**/**' --format html > report.html
```

Output:

| Flag | Meaning |
|------|---------|
| `-o, --output <file>` | Also write the report to a file (JSON is always compact there). |
| `--format json\|html` | Output format (default `json`); `html` writes a browsable page. |
| `--compact` | Force single-line JSON, even in a terminal. |
| `--pretty` | Force indented JSON, even when piped. |

Which files to scan:

| Flag | Meaning |
|------|---------|
| `--cwd <dir>` | Scan from this folder instead of the current one. |
| `--exclude <glob>` | Skip paths matching this glob (repeatable). |
| `--min-size <size>` | Skip files smaller than this (e.g. `1M`). |
| `--max-size <size>` | Skip files larger than this (e.g. `1G`). |
| `--follow-symlinks` | Follow symlinks and include their targets. |

Which duplicates to report:

| Flag | Meaning |
|------|---------|
| `--min-count <n>` | Only groups with at least `n` copies. |
| `--min-reclaimable <size>` | Only groups reclaiming at least this much space. |
| `--top <n>` | Keep only the `n` groups with the most reclaimable space. |

Matching and speed:

| Flag | Meaning |
|------|---------|
| `--hash blake3\|sha256\|xxh3` | Content hash (default `blake3`). |
| `--quick` | Approximate: sample-hash file ends, skip the byte comparison. |
| `-j, --jobs <n>` | Number of parallel hash workers (default: CPU count). |

Other:

| Flag | Meaning |
|------|---------|
| `--no-tui` | Turn off the rich progress UI, use plain log lines. |
| `--fail-if-duplicates` | Exit with code 2 if any duplicate group remains (for CI). |

### `summary`

Reads a report and shows it. On a terminal you get an interactive view; when the
output is piped you get a static table.

```sh
dupfind find | dupfind summary
dupfind summary -f report.json
```

Interactive keys: `↑/↓` move, `Enter` expand a group, `s` change sort, `/`
search, `c` copy the selected path, `q` quit.

Flags:

| Flag | Meaning |
|------|---------|
| `-f, --report-file <file>` | Read the report from a file. |
| `--no-tui` | Always use the static view. |
| `--clipboard osc52\|system` | How `c` copies paths (default: `osc52`). |

### `script`

Reads a report and prints a delete script for your shell. The first file of each
group stays (its delete command is commented out); the duplicates are deleted.

```sh
dupfind script -f report.json --shell bash > clean.sh
dupfind find | dupfind script --shell powershell > clean.ps1
```

Flags:

| Flag | Meaning |
|------|---------|
| `-f, --report-file <file>` | Read the report from a file. |
| `--shell auto\|bash\|zsh\|powershell` | Target shell (default: auto-detect). |

**Please review the script before you run it.** It deletes files.

## Input: file, environment, or STDIN

`summary` and `script` read the report in this order:

1. the `-f/--report-file` flag,
2. the `DUPFIND_REPORT_FILE` environment variable,
3. STDIN (when data is piped in).

## Global options

These work on every command:

- `--color auto|always|never` (default `auto`) — colorize output. `auto` only
  colors a real terminal; `NO_COLOR` is respected; `--color always` forces it.
- `--theme <name>` — syntax highlight theme (default `monokai`).
- `--si` — use 1000-based size units (kB, MB) instead of 1024-based (KB, MB).

## Colors and layout

- On a terminal, JSON and scripts are syntax-highlighted, and JSON is indented.
- When piped or redirected, output is plain and JSON is compact (one line), so it
  stays easy for other programs to read. `--compact` / `--pretty` force either
  way regardless of where the output goes.

## Environment variables

Every environment variable is a default: a command-line flag always wins over
it.

| Variable | Flag | Used by |
|----------|------|---------|
| `DUPFIND_CWD` | `--cwd` | `find` working directory |
| `DUPFIND_REPORT_FILE` | `-f/--report-file` | report input for `summary` / `script` |
| `DUPFIND_SHELL` | `--shell` | `script` target shell |
| `DUPFIND_THEME` | `--theme` | syntax highlight theme |
| `DUPFIND_CLIPBOARD` | `--clipboard` | `summary` clipboard backend |
| `DUPFIND_COLOR` | `--color` | color mode (auto/always/never) |
| `DUPFIND_SI` | `--si` | 1000-based (SI) size units |
| `DUPFIND_JOBS` | `-j/--jobs` | `find` parallel hash workers |
| `DUPFIND_NO_TUI` | `--no-tui` | disable the rich UI |
| `DUPFIND_HASH` | `--hash` | `find` hash algorithm |
| `DUPFIND_FORMAT` | `--format` | `find` output format |
| `DUPFIND_EXCLUDE` | `--exclude` | `find` skip globs (comma-separated) |
| `DUPFIND_FOLLOW_SYMLINKS` | `--follow-symlinks` | `find` follow symlinks |
| `NO_COLOR` | — | disables colors (standard) |

## Report format

```json
{
  "result": {
    "/full/path/to/first": {
      "hash": "<content-hash-hex>",
      "size": 123456789,
      "duplicates": [
        "/full/path/to/second",
        "/full/path/to/third"
      ]
    }
  }
}
```

Only files that have at least one duplicate appear. The key is the first
occurrence; the `duplicates` list holds the other copies.

## Development

```sh
go build ./...
go test ./...
```

## License

MIT — © 2026 Marcel Joachim Kloubert <marcel@kloubert.dev>
