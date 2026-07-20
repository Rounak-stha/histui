# histui

> Discover hidden patterns in your Git history

A lightning-fast CLI tool that reveals file coupling, contributor insights, and architectural patterns buried in your repository's commit history.

## Why histui?

Ever wondered which files always break together? Or which "unrelated" modules are secretly coupled? **histui** analyzes your Git history to surface the invisible dependencies that static analysis misses.

```bash
$ histui --coupling

Top 10 Strongly Coupled File Pairs:
──────────────────────────────────────────────────────────────────────────────────────────────────────────────
#    File A                               File B                               Score  Co-ch  Strength
──────────────────────────────────────────────────────────────────────────────────────────────────────────────
1    src/auth/login.go                    src/db/sessions.go                    0.95     38  Critical
2    models/user.go                       schemas/user.sql                      0.87     26  Strong
3    api/handlers.go                      api/middleware.go                     0.72     18  Strong
──────────────────────────────────────────────────────────────────────────────────────────────────────────────
```

## Features

### 🔗 File Coupling Detection

Identify files that change together across commits. Know what breaks when you refactor.

### 📊 Repository Statistics

Commit counts, contributor activity, line changes, date ranges — all in one glance.

### 🎯 Smart Filtering

Analyze specific branches, authors, date ranges, or commit limits. Ignore noise like docs and configs.

### ⚡ Blazing Fast

Built in Go. Analyzes thousands of commits in seconds.

## Installation

```bash
# From source
git clone https://github.com/rounak-stha/histui
cd histui
go build -o histui ./cmd/histui
```

## Quick Start

```bash
# Basic repository analysis
histui

# Analyze file coupling
histui --coupling

# Analyze specific branch
histui --coupling --branch main

# Limit to last 500 commits
histui --coupling -n 500

# Filter by author
histui --author "Jane Doe"

# Ignore documentation files
histui --coupling --ignore "*.md,*.txt,docs/*"
```

## Agent-friendly targeted queries

Build a bounded local index, then ask which files repeatedly changed with a target without generating every repository-wide pair:

```bash
histui index . --max-commits 5000
# Inspect freshness and policy without running a coupling query:
histui index status . --format json
# Later, process only fast-forward commits:
histui index . --update

histui query coupling . \
  --file internal/git/cli_repo.go \
  --max-commits 1000 \
  --limit 20 \
  --freshness cached \
  --format json
```

JSON stdout is stable, untruncated, and contains schema version, repository/ref, bounded history window, evidence counts, bulk-commit exclusions, and interpretation warnings. Errors go to stderr with a nonzero status. By default, commits touching more than 200 files are excluded and at least three co-changes are required. See [JSON schema v1](docs/json-schema-v1.md).

File and folder exclusions are supported by both global and targeted coupling queries:

```bash
histui query coupling . --file src/app.go \
  --ignore '*.md' \
  --ignore vendor \
  --ignore 'generated/**' \
  --format json
```

Patterns use repository-relative forward-slash paths. `*.md` matches that basename anywhere, `vendor` or `vendor/` excludes the whole directory recursively, `docs/*` matches direct children, and `docs/**` matches all descendants. Repeat `--ignore`, or pass comma-separated patterns in one occurrence. The explicitly requested target file is still analyzed even if a broad ignore pattern matches it.

Indexes are SQLite files under the platform user-cache directory, keyed by canonical repository path. Override the location with `--index-path`. They contain commit/file membership, subjects, and timestamps, remain local, and can be deleted safely. Queries never silently build an index: a missing index returns an actionable error. JSON reports `fresh` or `stale` by comparing indexed HEAD with current HEAD.

Freshness policies are `cached` (immediately use the index), `refresh` (incrementally append a bounded fast-forward range, otherwise return stale evidence with a warning), and `strict` (fail unless the index matches the selected HEAD). Diverged histories require a rebuild. Queries reject an index built for another ref; use separate `--index-path` values when maintaining multiple refs. `histui index --update` also performs only a fast-forward update and refuses option or history mismatches.

Use coupling to decide what current code and tests to inspect—not as proof of a dependency.

Useful for substantial refactors, cross-module bugs, and unfamiliar mature code. Prefer grep or language tooling for trivial edits and direct current-code questions.

### Pi installation

Install the `histui` binary separately and verify it is on `PATH`:

```bash
go install ./cmd/histui
histui --version
```

Then install this repository's Pi package, which bundles the typed `git_history` tool and the `histui` Agent Skill. CLI JSON schema v1 is compatible with `histui-pi` 0.x:

```bash
# Local development checkout
pi install /absolute/path/to/histui

# Or from Git after publication
pi install git:github.com/rounak-stha/histui
```

Build an index before the first agent query:

```bash
histui index . --max-commits 5000
```

You can explicitly load the skill in Pi with `/skill:histui`. The extension uses `ctx.cwd`, validates repository-relative paths, invokes the binary without a shell, applies a 15-second timeout and cancellation, and never builds a missing index automatically. Set `HISTUI_BIN` to an alternate binary path when developing. Pi extensions execute with user permissions; review package source and trust project-local packages before loading them.

## Usage

### Basic Analysis

```bash
histui [path]
```

Shows repository statistics:

- Current branch and total commits
- Date range of development
- Top contributors
- Files changed, lines added/deleted
- Recent commits

### File Coupling Analysis

```bash
histui --coupling
```

Reveals which files change together, helping you:

- **Plan refactoring**: Understand blast radius before making changes
- **Identify tech debt**: Find unintended dependencies between modules
- **Improve architecture**: Spot coupling that shouldn't exist
- **Organize teams**: Group related files for better ownership

### Flags

| Flag               | Short | Description                        | Default                          |
| ------------------ | ----- | ---------------------------------- | -------------------------------- |
| `--coupling`       | `-c`  | Show file coupling analysis        | `false`                          |
| `--max-commits`    | `-n`  | Limit number of commits to analyze | `0` (all)                        |
| `--branch`         | `-b`  | Analyze specific branch            | Current branch                   |
| `--author`         | `-a`  | Filter commits by author           | All authors                      |
| `--include-merges` | `-m`  | Include merge commits              | `false`                          |
| `--ignore`         | `-i`  | File patterns to ignore            | `*.md,*.txt,*.json,*.yaml,*.yml` |

### Examples

```bash
# Find coupling in last 100 commits on main branch
histui -c -b main -n 100

# Analyze a specific contributor's impact
histui -a "john@example.com"

# Full coupling analysis, include merge commits
histui --coupling --include-merges

# Custom ignore patterns (only show .go and .rs files)
histui -c --ignore "*.md,*.txt,*.json,*.yaml,*.yml,*.html,*.css"

# Analyze different repository
histui /path/to/other/repo --coupling
```

## Understanding Coupling Scores

**Coupling Score** = `(Times files changed together) / min(File A changes, File B changes)`

| Score     | Strength    | Meaning                             |
| --------- | ----------- | ----------------------------------- |
| 0.8 - 1.0 | 🔴 Critical | Files almost always change together |
| 0.5 - 0.8 | 🟠 Strong   | Frequently related changes          |
| 0.2 - 0.5 | 🟡 Moderate | Some relationship exists            |
| 0.0 - 0.2 | ⚪ Weak     | Rarely changed together             |

### Example

If `auth.go` changed 50 times and `db.go` changed 40 times, and they were modified together in 38 commits:

```
Score = 38 / min(50, 40) = 38/40 = 0.95 (Critical coupling)
```

This means **95% of the time** that `db.go` changes, `auth.go` also changes. They're tightly coupled.

## Real-World Use Cases

### 🔧 Refactoring Planning

```bash
histui -c -n 200 --ignore "*.md,test/*"
```

Identify which files will likely need updates when refactoring a module.

### 🏗️ Architecture Review

```bash
histui -c --ignore "*.md,*.json,*.yaml"
```

Find unintended coupling between supposedly independent modules.

### 👥 Team Organization

```bash
histui -c -b develop
```

Discover which files should be owned by the same team.

### 📈 Technical Debt Assessment

```bash
histui -c --since "6 months ago"
```

Track how coupling evolves over time. Increasing coupling = growing tech debt.

## How It Works

1. **Parses Git history** using `git log` with custom formatting
2. **Extracts file changes** from each commit using `--numstat`
3. **Builds co-change matrix** tracking which files appear together
4. **Calculates coupling scores** using statistical analysis
5. **Ranks and displays** the strongest relationships

All processing happens locally. No data leaves your machine.

## Performance

Git output is parsed and inserted into SQLite incrementally inside an atomic rebuild transaction, so initial indexing does not retain the selected history as a full in-memory commit slice. Targeted indexed queries avoid generating every repository-wide pair. Measure your repository with:

```bash
go test -run '^$' -bench . -benchmem ./internal/analysis ./internal/index
```

SQLite is embedded in the binary; no database service is required.

## Roadmap

- [ ] Interactive TUI mode with file selection
- [ ] Cluster detection (group related files automatically)
- [ ] Historical trend analysis (coupling over time)
- [x] Stable JSON output for targeted coupling queries
- [ ] Git hooks integration for CI/CD
- [ ] Visual graph output (HTML/SVG)

## Contributing

Contributions welcome! Please open an issue first to discuss what you'd like to change.

```bash
# Setup
git clone https://github.com/rounak-stha/histui
cd histui
go mod download

# Run
go run ./cmd/histui --coupling

# Build
go build -o histui ./cmd/histui
```

---

**[⭐ Star)** if histui helped you understand your codebase better!
