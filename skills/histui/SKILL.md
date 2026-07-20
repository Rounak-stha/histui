---
name: histui
description: Uses indexed Git history to find files that repeatedly changed with a target file and estimate hidden change blast radius. Use when planning substantial refactors, investigating cross-file bugs, or exploring unfamiliar mature code. Do not use for trivial edits or as proof of a current dependency.
license: MIT
compatibility: Requires Git history and the histui CLI; prefers Pi's git_history tool when available.
---

# Histui historical context

Use history selectively. The goal is to discover additional files worth inspecting, not to inject broad repository history or replace current-code analysis.

## When to use

Use one targeted query near the start of:

- a substantial refactor or architectural change
- a bug investigation likely to cross modules
- work on a frequently changed central file
- unfamiliar mature code where associated tests, schemas, migrations, configuration, or generated files may be easy to miss

Do not use it for trivial/textual edits, repositories with little relevant history, or questions answered directly by grep, language tooling, or tests. Avoid repeated calls unless another target would change the next action.

## Workflow

1. Identify the most central repository-relative file in the planned change.
2. Prefer the typed `git_history` tool:

   ```text
   git_history(file: "internal/git/cli_repo.go", maxCommits: 1000, limit: 20, freshness: "cached")
   ```

3. If the tool is unavailable, use the CLI:

   ```bash
   histui query coupling . --file internal/git/cli_repo.go --max-commits 1000 --limit 20 --freshness cached --format json
   ```

4. Inspect only related files that could affect the task using read/grep and static analysis.
5. Confirm relationships in current code and tests before planning or editing.
6. Mention historical evidence as correlation, including weak/stale evidence and exclusions.

Keep the result bounded. Start with 10–20 related files and 1,000 commits. Follow up only when the first query identifies a meaningful area. When generated, vendored, documentation, or fixture trees create noise, pass ignore patterns such as `*.md`, `vendor`, or `generated/**`; folder names and trailing-slash patterns exclude their subtrees recursively.

## Index setup and freshness

The tool never silently builds a full index. If missing, tell the user or run explicitly when requested:

```bash
histui index . --max-commits 5000
```

Update an existing fast-forward index with:

```bash
histui index . --update
```

Freshness modes:

- `cached`: return immediately; clearly note stale status
- `refresh`: attempt a bounded fast-forward update, otherwise return stale evidence with warnings
- `strict`: fail unless the index matches selected HEAD

Prefer `cached` for interactive planning. Use `refresh` when recent commits matter and bounded work is acceptable. Use `strict` only when stale evidence would be misleading. Diverged/rebased history requires rebuilding the index.

## Interpretation

The coupling score is:

```text
coChanges / min(targetChanges, relatedFileChanges)
```

Interpret it alongside evidence volume:

- High score with many co-changes is stronger evidence.
- High score from only three co-changes remains weak.
- No result may mean isolation, ignored files, a short history window, recent code, or insufficient evidence.
- Bulk commits are excluded by default because formatting, vendoring, generated code, and mass migrations create noisy relationships.
- Ignore patterns and merge policy alter the evidence window.
- Stale indexes describe the indexed HEAD, not necessarily current code.

Never claim that coupling proves a source-level, runtime, or architectural dependency. It may reveal tests, release files, schemas, migrations, or workflow conventions instead.

## If unavailable

If the binary is missing, the index is unavailable, or a fresh query would be too expensive, continue with current-code tools. Use repository search, language references, tests, build configuration, and nearby naming conventions to estimate blast radius. Do not block a task solely because historical evidence is unavailable.
