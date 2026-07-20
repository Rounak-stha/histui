# JSON output schema v1

`histui query coupling --format json` writes exactly one JSON object to stdout. Diagnostics and errors are written to stderr and return a nonzero exit status.

## Compatibility

- `schemaVersion` is currently `1`.
- Fields documented here will retain their meaning for schema v1.
- Consumers must reject unsupported major schema versions and should ignore unknown fields.
- New optional fields may be added without changing `schemaVersion`.

## File-coupling response

| Field | Type | Meaning |
| --- | --- | --- |
| `schemaVersion` | integer | Output contract version. |
| `repository.path` | string | Absolute repository path. |
| `repository.ref` | string | Checked-out branch or explicitly requested ref. |
| `repository.currentHead` | string | Current HEAD SHA. |
| `query.type` | string | Always `file-coupling` for this command. |
| `query.file` | string | Normalized repository-relative target path. |
| `query.maxCommits` | integer | Bounded history window requested. |
| `query.limit` | integer | Maximum result count. |
| `query.minCoChanges` | integer | Minimum evidence threshold. |
| `index.status` | string | `fresh` when indexed HEAD/ref match, otherwise `stale`. |
| `index.indexedHead` | string | HEAD represented by the analysis. |
| `index.analyzedCommits` | integer | Commits actually loaded in the selected window. |
| `index.createdAt` | RFC 3339 string | Time the current index was first built. |
| `index.updatedAt` | RFC 3339 string | Time the index was last rebuilt or incrementally updated. |
| `results` | array | Deterministically sorted related-file evidence. |
| `exclusions.bulkCommits` | integer | Commits excluded by `--max-files-per-commit`. |
| `exclusions.ignoredFiles` | integer | Ignored file occurrences across non-bulk commits in the selected history window. |
| `warnings` | string array | Interpretation and quality caveats. |

Each result contains `path`, `coChanges`, `targetChanges`, `relatedFileChanges`, `score`, and `strength`. The score is:

```text
coChanges / min(targetChanges, relatedFileChanges)
```

Scores are rounded to three decimal places. Results sort by score descending, co-change count descending, then path ascending. Historical correlation is evidence, not proof of a current dependency.

Bulk commits are excluded from both change totals and coupling evidence. Duplicate paths within one commit count once. Renames use the new path emitted by Git's rename detection.

`--ignore` accepts repeatable or comma-separated repository-relative patterns. Patterns use `/` separators: basename patterns such as `*.md` match at any depth; a plain directory name or trailing slash such as `vendor` or `vendor/` excludes its subtree; `*` matches one path segment; and `**` matches any number of segments. Ignored files are removed from totals and coupling evidence, except that an explicitly requested target remains included.

## Index lifecycle and privacy

Build with `histui index [path] --max-commits 5000`. Initial history is streamed from Git directly into one atomic SQLite rebuild transaction rather than retained as a complete in-memory commit list. By default the SQLite file is stored below the platform user-cache directory and keyed by canonical repository path; `--index-path` overrides it. The index remains local and stores paths, commit subjects/timestamps, parent SHAs, and commit-to-file membership. Deleting the SQLite file and adjacent WAL/SHM files removes the cache.

Queries never build a missing index implicitly. A missing index is an error with setup instructions. A HEAD mismatch returns cached evidence with `index.status: "stale"` under `--freshness cached`; a requested ref that differs from the indexed ref is rejected because its evidence may describe a different history. Use `--index-path` to maintain separate indexes for multiple refs. `--freshness refresh` appends a bounded fast-forward range; divergence or an oversized refresh remains stale with a warning. `--freshness strict` returns a nonzero error unless HEAD matches. `histui index --update` performs an explicit unbounded fast-forward update and refuses divergence or option mismatches.

The internal SQLite schema is versioned separately from this JSON contract. Incompatible index schema changes require rebuilding the local cache and do not change `schemaVersion` unless the JSON response itself changes.
