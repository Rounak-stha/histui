# histui Coding-Agent Integration Plan

## Purpose

Extend **histui** from a human-oriented Git history report into a fast, structured source of historical context for coding agents, initially targeting the **Pi coding agent**.

Static analysis shows what code references today. Git history provides another signal: which files repeatedly changed together, who worked on an area, how a module evolved, and which files may be part of a change's hidden blast radius. Coding agents can use this evidence before a refactor, bug fix, or architectural change to inspect files they might otherwise miss.

The intended result is not to put an entire repository history into the model context. The intended result is to let an agent ask a small, targeted question at the right time and receive a compact, fast, machine-readable answer.

## Inspiration

The project already discovers file coupling by analyzing co-changes across commits. That information is especially valuable to a coding agent because an agent usually begins with only the current code, the user's prompt, and a limited context window.

The core motivating question is:

> Given the file or area I am about to change, what other files does Git history suggest I should inspect?

This is inspired by the idea that repository history is a form of architectural documentation. Repeated co-change does not prove a source-level or runtime dependency, but it can reveal tests, schemas, configuration, generated artifacts, migration files, or cross-module relationships that static inspection has not yet exposed.

## Desired outcome

A user should be able to start Pi in a Git repository and ask it to plan or perform a meaningful change. When historical context would help, Pi should be able to call a focused tool such as:

```text
git_history(file: "internal/git/cli_repo.go", maxCommits: 1000)
```

The result should arrive quickly and contain:

- historically related files
- co-change counts and coupling scores
- the history window and filters used
- index freshness information
- warnings about excluded bulk commits or weak evidence

Pi should then inspect relevant files and combine historical evidence with current code analysis. It must not treat co-change as proof of dependency.

The final integration should be publishable and easy to install as:

1. a standalone `histui` CLI
2. an Agent Skill that explains when and how to use histui
3. a Pi extension that exposes histui as a typed agent tool
4. ideally, one Pi package containing both the extension and skill

## Product principles

1. **Targeted, not exhaustive** — Agents normally need history for one file or area, not every pair in the repository.
2. **Indexed in advance, queried on demand** — Expensive history traversal should not happen on every tool call.
3. **Fast startup** — Pi startup must never block on a full repository analysis.
4. **Progressive disclosure** — Only compact results enter the model context; detailed evidence is requested only if needed.
5. **Transparent freshness** — Cached results must report which commit/ref and options they represent.
6. **Evidence, not authority** — Historical coupling supplements current code, tests, and documentation.
7. **Safe defaults** — Do not silently launch unbounded multi-minute work from an agent call.

## When this will be useful

The skill and extension should encourage histui use when:

- planning a substantial refactor
- modifying a frequently changed or central file
- estimating the likely blast radius of a change
- investigating a bug that may span modules
- looking for associated tests, schemas, migrations, configuration, or generated files
- reviewing architectural boundaries or suspected accidental coupling
- onboarding to an unfamiliar part of a mature repository
- deciding which additional files should be read before editing
- assessing ownership or contributor knowledge around an area, once contributor queries exist

## When not to use it

Histui should normally not be used when:

- the change is trivial, isolated, or purely textual
- the repository has little or no meaningful Git history
- the relevant code was recently added and has insufficient historical evidence
- generated, vendored, formatting, or mass-migration commits dominate the history
- the user only needs a direct current-code lookup that `grep`, language tooling, or tests answer better
- a fresh index would take a long time and stale results would not be acceptable
- the repository is not trusted or the user does not want local history indexed
- coupling output would add context without changing the agent's next action

The agent should not repeatedly call histui for every edited file. One query near the start of a significant task, followed by focused follow-up queries, is preferable.

## Current-state observations

The current implementation:

- is written in Go and invokes the Git CLI
- supports repository summaries and global coupling analysis
- loads `git log --numstat` output into memory using `cmd.Output()`
- materializes commits and changed files before analysis
- computes all pairs for each commit
- uses string-concatenated pair keys
- emits human-oriented terminal output
- has no stable JSON mode, persistent index, targeted `--file` query, or cache freshness contract

For a commit touching `k` files, global pair generation is approximately:

```text
k * (k - 1) / 2
```

Bulk formatting, vendoring, generated-code, and repository-wide migration commits can therefore hurt both performance and result quality.

## Proposed user-facing CLI

Exact Cobra command structure can be refined during implementation, but the intended interface is:

```bash
# Build or rebuild an index
histui index [path] --max-commits 5000

# Incrementally update an existing index
histui index [path] --update

# Query files historically coupled to one file
histui query coupling [path] \
  --file internal/git/cli_repo.go \
  --max-commits 1000 \
  --limit 20 \
  --format json

# Query several files as a change set
histui query coupling [path] \
  --files internal/git/cli_repo.go,internal/git/models.go \
  --format json

# Preserve a global report for human use
histui query coupling [path] --top 20
```

Backward compatibility with `histui --coupling` should be retained initially, with documentation directing agent integrations toward `histui query ... --format json`.

Potential later queries:

```bash
histui query commits --file path/to/file.go --format json
histui query contributors --file path/to/file.go --format json
histui query related --directory internal/git --format json
histui query change-set --files a.go,b.go --format json
```

## Structured output contract

Add `--format json` and keep it stable enough for integrations. In JSON mode:

- `stdout` must contain JSON only
- progress and diagnostic messages go to `stderr`
- paths must not be visually truncated
- invalid arguments, unavailable indexes, and Git failures return nonzero exit codes
- field names and schema versions must be stable
- terminal color and decorative output must be disabled

Example targeted response:

```json
{
  "schemaVersion": 1,
  "repository": {
    "path": "/project",
    "ref": "main",
    "currentHead": "def456"
  },
  "query": {
    "type": "file-coupling",
    "file": "internal/git/cli_repo.go",
    "maxCommits": 1000,
    "limit": 20
  },
  "index": {
    "status": "stale",
    "indexedHead": "abc123",
    "analyzedCommits": 1000,
    "createdAt": "2026-07-19T12:00:00Z"
  },
  "results": [
    {
      "path": "internal/git/models.go",
      "coChanges": 18,
      "targetChanges": 25,
      "relatedFileChanges": 21,
      "score": 0.857,
      "strength": "critical"
    }
  ],
  "exclusions": {
    "bulkCommits": 14,
    "ignoredFiles": 82
  },
  "warnings": [
    "Historical coupling is evidence, not proof of a current dependency."
  ]
}
```

## File-focused coupling semantics

The first high-value query is coupling for a target file.

For a requested file:

1. Find indexed commits containing the target file.
2. Count the other files in those commits.
3. Determine each file's total changes in the selected history window.
4. Calculate the existing coupling score:

```text
coChanges / min(targetChanges, relatedFileChanges)
```

5. Apply a minimum evidence threshold, initially three co-changes unless overridden.
6. Sort deterministically by score, then co-change count, then path.
7. Return only the requested limit.

Consider exposing additional measures later, such as confidence, Jaccard similarity, lift, or directional probabilities. Avoid changing the meaning of the existing score without versioning the output.

For a multi-file change set, define semantics explicitly. A useful first version can return the union of related files with evidence grouped by each requested file. A later version can score relationships to the change set as a whole.

## Index and cache architecture

Use a hybrid model:

- **Index history in advance.**
- **Update the index incrementally.**
- **Run small, targeted queries on demand.**
- **Never inject the full index into Pi's context.**

SQLite is a good initial storage choice because it is local, portable, transactional, queryable, and available from Go libraries. Avoid requiring a separate database service.

A starting schema:

```sql
CREATE TABLE metadata (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE commits (
    id INTEGER PRIMARY KEY,
    sha TEXT UNIQUE NOT NULL,
    timestamp INTEGER NOT NULL,
    parent_shas TEXT,
    author_name TEXT,
    author_email TEXT,
    subject TEXT,
    is_merge INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE files (
    id INTEGER PRIMARY KEY,
    path TEXT UNIQUE NOT NULL
);

CREATE TABLE commit_files (
    commit_id INTEGER NOT NULL,
    file_id INTEGER NOT NULL,
    additions INTEGER NOT NULL DEFAULT 0,
    deletions INTEGER NOT NULL DEFAULT 0,
    old_path TEXT,
    change_type TEXT,
    PRIMARY KEY (commit_id, file_id),
    FOREIGN KEY (commit_id) REFERENCES commits(id),
    FOREIGN KEY (file_id) REFERENCES files(id)
);

CREATE INDEX commit_files_by_file
ON commit_files(file_id, commit_id);

CREATE INDEX commits_by_timestamp
ON commits(timestamp);
```

Store index metadata including:

- schema/index version
- canonical repository identity and path
- selected ref/branch
- indexed HEAD
- oldest/newest indexed commit
- number of analyzed commits
- merge policy
- ignore patterns
- maximum-files-per-commit policy
- creation and update timestamps

Do not precompute and persist every pair by default. Store commit-to-file membership and calculate targeted coupling from the subset of commits associated with the requested file. Optionally cache frequently requested file results or precompute a bounded list of global top pairs later.

Choose a cache location deliberately. Options include `.git/histui/`, the platform user-cache directory keyed by repository identity, or a configurable path. The index should not be committed to the repository. Document privacy implications: commit metadata, paths, authors, and subjects remain local unless users explicitly export them.

## Freshness and incremental updates

Before a query, compare the indexed HEAD and requested ref to the current HEAD.

### Fresh index

If both HEADs match, answer immediately.

### Fast-forward update available

Use:

```bash
git merge-base --is-ancestor <indexed-head> <current-head>
```

If true, process only:

```text
<indexed-head>..HEAD
```

### Diverged history

If the indexed HEAD is not an ancestor of the current HEAD, history may have been rebased, force-updated, or switched. Rebuild the affected ref's index or maintain separate index identities per ref.

### Stale but usable

A query may return stale results when policy allows, but it must include:

- indexed HEAD
- current HEAD
- stale/fresh status
- analyzed history window
- relevant warnings

Support a freshness policy:

```text
cached  - return existing data immediately; do not block for refresh
refresh - incrementally refresh when inexpensive, otherwise use stale data with metadata
strict  - answer only from an index matching the requested HEAD
```

The Pi tool should default to `cached` or a bounded `refresh`, not `strict`.

### Stream Git output

Replace whole-output buffering where practical:

```go
out, err := cmd.Output()
```

with `StdoutPipe` and incremental parsing. Be mindful of `bufio.Scanner`'s default token limit; increase its buffer or use `bufio.Reader`. Index commits in batches inside SQLite transactions.

### Handle bulk commits

Add a configurable policy such as:

```bash
--max-files-per-commit 200
```

Possible behavior:

- exclude oversized commits from coupling evidence by default
- retain their metadata for transparency
- report excluded counts in query results
- optionally allow down-weighting or explicit inclusion later

This protects runtime and avoids treating mass formatting or generated changes as architecture.

### Improve in-memory global analysis

If global pair analysis remains supported:

- intern file paths to integer IDs
- replace concatenated string keys with a comparable struct such as `{A uint32, B uint32}`
- skip ignored files before deciding whether a commit has enough valid files
- avoid duplicate paths within one commit
- apply deterministic tie-breaking when sorting
- consider bounded top-N algorithms where complete pair retention is unnecessary

### Profile before optimizing further

Measure separately:

- Git history traversal
- parsing
- database writes
- targeted query execution
- global pair generation
- sorting
- JSON/human output formatting
- peak memory and allocations

Use Go benchmarks and `pprof` on small, medium, and monorepo-sized repositories, including histories with large bulk commits. Consider alternatives like Rust only if optimized, indexed Go code remains demonstrably CPU- or memory-bound and a focused prototype shows a meaningful end-to-end gain.

## Pi extension

Create a TypeScript Pi extension that registers a typed tool, tentatively named `histui`.

Suggested parameters:

```text
file?: string
files?: string[]
maxCommits?: number
limit?: number
freshness?: "cached" | "refresh" | "strict"
includeMerges?: boolean
```

The tool should:

1. Use `ctx.cwd` as the repository path.
2. Validate that requested files are repository-relative and remain within the repository.
3. Spawn `histui` directly with an argument array; do not construct a shell command.
4. Request JSON output and parse it.
5. Respect the provided `AbortSignal` and terminate the child process on cancellation.
6. Apply a configurable timeout.
7. Return compact text for the model plus structured details for session persistence/debugging.
8. Report index absence, staleness, timeout, incompatible schema, and binary-not-found errors clearly.
9. Never silently start an unlimited full index build.
10. Avoid exposing author emails or unnecessary commit content to the model for a coupling query.

The tool description and prompt guidance should tell the model:

- use the tool before significant refactors or when hidden file relationships may affect a change
- do not use it for trivial edits
- treat results as clues and inspect current code before acting
- avoid repeated calls that do not change the next action

Example tool result shown to the model:

```text
History index: stale by 8 commits; 1,000 commits analyzed.
Target: internal/git/cli_repo.go (25 historical changes)
Related files:
1. internal/git/models.go — score 0.857, 18 co-changes
2. internal/git/repository.go — score 0.720, 12 co-changes
Excluded 4 bulk commits (>200 files).
Historical coupling is not proof of a current dependency.
```

### Extension startup behavior

On Pi startup, perform only cheap checks if any:

- locate the histui executable
- optionally inspect index metadata
- optionally show a small status indicator

Do not traverse Git history during `session_start`. The first relevant tool call can use cached data, request an incremental refresh, or explain how to build an index.

### Missing-index behavior

For the initial release, prefer explicit consent and predictable work:

- If no index exists, return instructions to run `histui index --max-commits <bounded-value>`.
- Optionally allow the tool to offer a small bounded fallback, such as 200 commits, with a timeout.
- A future TUI command may ask the user whether to build an index.
- Print/JSON/RPC modes must not depend on an interactive confirmation.

## Agent Skill

Create a standards-compatible skill, for example:

```text
skills/histui/SKILL.md
```

Frontmatter should use a specific discovery description, such as:

```yaml
---
name: histui
description: Uses indexed Git history to find files that repeatedly changed with a target file, estimate change blast radius, and surface historical contributors or commits. Use when planning substantial refactors, investigating cross-file bugs, or exploring unfamiliar mature code. Do not use for trivial edits or as proof of a current dependency.
---
```

The skill should teach the agent:

1. when historical analysis is valuable
2. when not to invoke it
3. how to check/build/update the index
4. how to begin with a targeted file query
5. how to interpret coupling score, co-change count, history window, and freshness
6. how bulk commits and ignored patterns affect evidence
7. how to combine history with `read`, `grep`, tests, and static analysis
8. how to avoid overclaiming from correlation
9. how to limit output and preserve context
10. how to proceed when histui is unavailable or an index is stale

The skill can support both the Pi extension tool and direct CLI use, but should prefer the typed extension tool when available.

## Packaging and publishing

Publish both the extension and skill so users can install them together. A package layout could be:

```text
package.json
extensions/
  histui.ts
skills/
  histui/
    SKILL.md
README.md
LICENSE
```

Example package metadata:

```json
{
  "name": "histui-pi",
  "keywords": ["pi-package", "git-history", "coding-agent"],
  "peerDependencies": {
    "@earendil-works/pi-coding-agent": "*",
    "typebox": "*"
  },
  "pi": {
    "extensions": ["./extensions"],
    "skills": ["./skills"]
  }
}
```

Decide whether this package lives in the histui repository or a dedicated package repository. Keeping it in this repository initially makes CLI and integration versions easier to coordinate.

The published integration needs a clear binary strategy:

- require users to install `histui` separately and detect it on `PATH`; or
- distribute platform-specific binaries and select the correct one; or
- download verified release assets during an explicit setup command.

The safest first release is to require a separately installed binary and provide actionable setup instructions. Avoid unannounced downloads or install-time scripts. If automatic downloads are added later, pin versions, verify checksums/signatures, support offline installation, and require explicit user action.

Potential installation experience:

```bash
# Install histui CLI by documented platform method first
histui --version

# Install extension and skill globally
pi install npm:histui-pi

# Or install from Git while developing
pi install git:github.com/rounak-stha/histui
```

A project-local Pi install can be documented for teams that want shared behavior:

```bash
pi install -l npm:histui-pi
```

Remember that Pi extensions execute with user permissions and project-local resources require project trust. Document that users should review the package source and that history/index data stays local by default.

## Documentation requirements

Update the main README to include:

- coding-agent use case
- indexing versus querying
- example file-focused JSON query
- cache location and privacy
- freshness semantics
- bulk-commit behavior
- Pi installation and setup
- direct skill invocation example
- troubleshooting for missing/stale indexes
- guidance on useful and inappropriate usage

Fix outdated build examples to use the actual command path:

```bash
go build -o histui ./cmd/histui
```

Add schema documentation for JSON consumers and a compatibility/versioning policy.

## Testing strategy

### Go unit tests

Add tests for:

- targeted coupling calculations
- minimum co-change threshold
- ignored files
- duplicate paths in a commit
- bulk-commit exclusion
- rename handling
- deterministic ordering
- JSON schema fields
- cache metadata validation
- fresh, fast-forward, and diverged HEAD detection

### Integration tests

Create temporary Git repositories that cover:

- no commits
- one commit
- single-file commits only
- known repeated co-change patterns
- branch switching
- incremental commits after indexing
- rebase/divergence
- merge inclusion/exclusion
- large/bulk commits
- renames
- paths with spaces and unusual characters

Verify that JSON mode keeps `stdout` clean and sends progress to `stderr`.

### Extension tests

Test:

- successful parsing of histui JSON
- missing executable
- no index
- stale index
- schema mismatch
- malformed output
- timeout and cancellation
- nonzero process exit
- paths outside the repository
- compact model-facing output
- behavior in TUI, print, JSON, and RPC modes where applicable

### Performance acceptance tests

Establish baselines before implementation and targets after profiling. Suggested goals, subject to real measurements:

- cached targeted query: comfortably below one second on a large repository
- incremental update after a few commits: a few seconds or less
- bounded memory during initial indexing through streaming/batching
- Pi startup: no measurable full-history work
- timeout or explicit confirmation before any potentially long fallback

## Delivery phases

### Phase 0 — Baseline and specification

- Add phase-level timing and representative benchmarks.
- Define CLI command structure and JSON schema version 1.
- Decide index location and repository identity rules.
- Document coupling semantics and freshness policy.

**Exit criteria:** Existing performance is measured, and the external contracts are written before implementation.

### Phase 1 — Agent-friendly output and targeted query

- Add JSON output with clean stdout.
- Add `--file`, `--limit`, and bounded `--max-commits` behavior.
- Implement targeted coupling without generating every global pair.
- Add deterministic sorting and warnings.
- Add bulk-commit exclusion policy.
- Preserve the existing human-readable command.

**Exit criteria:** A direct targeted CLI query produces compact stable JSON and is meaningfully faster than full global coupling.

### Phase 2 — Persistent index

- Add SQLite storage and metadata.
- Stream Git history into batched transactions.
- Implement `histui index` and index inspection.
- Query commit/file membership without loading all history.
- Document cache privacy and deletion.

**Exit criteria:** Repeated targeted queries do not retraverse Git history.

### Phase 3 — Incremental refresh and stale policy

- Detect fresh, fast-forward, and diverged histories.
- Process only new commits after a fast-forward.
- Support `cached`, `refresh`, and `strict` query policies.
- Report freshness in JSON.
- Add optional, documented post-commit refresh guidance without installing hooks automatically.

**Exit criteria:** Normal updates after a few commits are fast and stale results are never presented as fresh.

### Phase 4 — Pi extension

- Register the typed history tool.
- Add safe process execution, cancellation, and timeout.
- Produce compact model-facing summaries.
- Handle binary/index errors and freshness metadata.
- Ensure no full indexing occurs during Pi startup.

**Exit criteria:** Pi can use a cached file-focused query during a real refactoring workflow without manual output piping.

### Phase 5 — Skill and package publication

- Write and validate the histui skill.
- Package extension and skill together.
- Add installation, security, privacy, and usage docs.
- Test local, Git, and npm package installation as applicable.
- Publish versioned releases with compatibility notes.

**Exit criteria:** A new user can install the CLI and Pi package, build an index, and have Pi correctly decide when to query it.

### Phase 6 — Follow-up capabilities

Based on usage and profiling, consider:

- file-specific commit summaries
- contributor/ownership queries
- directory or module-level coupling
- change-set queries for several planned files
- evidence sampling with relevant commit SHAs/subjects
- bounded global top-pair caches
- optional background refresh or explicit setup UI
- integrations for coding agents other than Pi

Do not add these before the targeted coupling path is reliable and fast.

## Risks and mitigations

### Correlation mistaken for dependency

**Risk:** The agent treats a high coupling score as proof.

**Mitigation:** Include warnings in tool descriptions, skill instructions, and results; require inspection of current code and tests.

### Noisy bulk commits

**Risk:** Formatting or generated changes create false coupling and quadratic work.

**Mitigation:** Exclude or down-weight oversized commits and report exclusions.

### Stale results

**Risk:** A rebase or branch switch invalidates the index.

**Mitigation:** Track ref/HEAD/options, detect ancestry, expose freshness, and rebuild on divergence.

### Long-running agent calls

**Risk:** Pi silently starts a full index and appears stuck.

**Mitigation:** Cache-first queries, timeouts, bounded fallback, explicit setup, and no startup indexing.

### Context pollution

**Risk:** Huge JSON results consume the model context.

**Mitigation:** Default limits, compact summaries, targeted queries, and detailed data only on request.

### Privacy

**Risk:** Commit authors, messages, or proprietary paths are unnecessarily exposed.

**Mitigation:** Keep indexes local, return only required fields, document storage, and avoid sending author emails in coupling responses.

### Packaging and binary mismatch

**Risk:** The TypeScript extension and Go binary support incompatible schemas.

**Mitigation:** Add `histui --version`, JSON schema versions, supported-version checks, and clear errors.

## Definition of done

The initial coding-agent integration is complete when:

- histui can build and incrementally update a local history index
- a file-focused coupling query returns stable compact JSON
- a cached targeted query is fast enough for interactive agent use
- cache freshness and exclusions are explicit
- bulk commits cannot accidentally dominate default analysis
- the Go implementation has benchmark/profile evidence supporting its performance
- Pi has a safe typed extension tool with cancellation and timeout
- Pi has a skill that explains when, why, and how to use the tool—and when not to
- both extension and skill are packaged and documented for publication
- the agent uses history as supporting evidence and reads current related files before making conclusions

## Starting instructions for the next session

Begin with **Phase 0**, not the extension. Inspect the current command and Git parsing code, then:

1. establish benchmark repositories and phase-level timings
2. write the JSON schema and CLI behavior as tests/specification
3. implement the targeted in-memory `--file` coupling query before SQLite
4. add bulk-commit handling and deterministic output
5. verify the value and performance of that API
6. then design and implement the persistent index

Keep changes incremental and testable. Do not add automatic background hooks or build a complex Pi UI before the targeted CLI contract is stable. The Pi extension and skill depend on that contract and should be implemented after it is reliable.
