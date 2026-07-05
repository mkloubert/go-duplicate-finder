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
2. Hash the same-size files with BLAKE3 (a fast, secure hash).
3. Compare matching files byte by byte to be 100% sure.

Empty (0-byte) files are skipped.

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
```

## Commands

### `find [patterns...]`

Scans one or more glob patterns and prints a JSON report to STDOUT. The default
pattern is `**/**` (everything, recursively).

```sh
dupfind find '**/*.jpg' '**/*.png'
dupfind find --cwd /photos '**/**' -o report.json
```

Flags:

| Flag | Meaning |
|------|---------|
| `-o, --output <file>` | Also write the JSON to a file (always compact). |
| `-j, --jobs <n>` | Number of parallel hash workers (default: CPU count). |
| `--cwd <dir>` | Scan from this folder instead of the current one. |
| `--no-tui` | Turn off the rich progress UI, use plain log lines. |
| `--compact` | Force single-line JSON, even in a terminal. |
| `--pretty` | Force indented JSON, even when piped. |

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

## Colors and layout

- On a terminal, JSON and scripts are syntax-highlighted, and JSON is indented.
- When piped or redirected, output is plain and JSON is compact (one line), so it
  stays easy for other programs to read.
- `--color auto|always|never` (default `auto`) controls colors; `NO_COLOR` is
  respected. `--theme <name>` picks the highlight theme (default `monokai`).

## Environment variables

| Variable | Used by |
|----------|---------|
| `DUPFIND_CWD` | `find` working directory |
| `DUPFIND_REPORT_FILE` | report input for `summary` / `script` |
| `DUPFIND_SHELL` | `script` target shell |
| `DUPFIND_THEME` | syntax highlight theme |
| `DUPFIND_CLIPBOARD` | `summary` clipboard backend |
| `NO_COLOR` | disables colors |

## Report format

```json
{
  "result": {
    "/full/path/to/first": {
      "hash": "<blake3-hex>",
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
