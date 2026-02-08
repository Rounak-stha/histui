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
git clone https://github.com/yourusername/histui
cd histui
go build -o histui cmd/main.go
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

- ✅ Analyzes 10,000 commits in ~4 seconds
- ✅ Minimal memory footprint
- ✅ Single binary, no dependencies

## Roadmap

- [ ] Interactive TUI mode with file selection
- [ ] Cluster detection (group related files automatically)
- [ ] Historical trend analysis (coupling over time)
- [ ] Export to JSON/CSV for custom analysis
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
go run cmd/main.go --coupling

# Build
go build -o histui cmd/main.go
```

---

**[⭐ Star)** if histui helped you understand your codebase better!
