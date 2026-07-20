import assert from "node:assert/strict";
import test from "node:test";
import { explainFailure, repositoryRelativeFile, summarize } from "./histui.ts";

test("repositoryRelativeFile accepts normalized paths", () => {
	assert.equal(repositoryRelativeFile("/repo", "@internal/../src/file.go"), "src/file.go");
});

test("repositoryRelativeFile rejects absolute, parent, and root paths", () => {
	for (const path of ["/etc/passwd", "../outside.go", ".", ""]) {
		assert.throws(() => repositoryRelativeFile("/repo", path));
	}
});

test("explainFailure distinguishes a missing index from a missing executable", () => {
	assert.match(explainFailure("history index not found; run: histui index .", 1, false), /Build a bounded index explicitly/);
	assert.doesNotMatch(explainFailure("history index not found", 1, false), /executable was not found/);
	assert.match(explainFailure("spawn histui ENOENT", 1, false), /executable was not found/);
	assert.match(explainFailure("", 1, true), /timed out or was cancelled/);
	assert.equal(explainFailure("git failed", 2, false), "git failed");
});

test("summarize remains compact and includes warnings", () => {
	const text = summarize({
		schemaVersion: 1,
		repository: { path: "/repo", ref: "main", currentHead: "def" },
		query: { file: "target.go", maxCommits: 1000, limit: 2, minCoChanges: 3 },
		index: { status: "stale", indexedHead: "abc", analyzedCommits: 1000 },
		results: [{ path: "related.go", coChanges: 7, targetChanges: 10, relatedFileChanges: 8, score: 0.875, strength: "critical" }],
		exclusions: { bulkCommits: 2, ignoredFiles: 4 },
		warnings: ["Historical coupling is evidence."],
	});
	assert.match(text, /History index: stale/);
	assert.match(text, /Target: target.go \(10 historical changes\)/);
	assert.match(text, /related.go — score 0.875, 7 co-changes/);
	assert.match(text, /Excluded 2 bulk commits/);
	assert.match(text, /Warning: Historical coupling is evidence/);
	assert.ok(text.length < 500);
});

test("summarize does not invent a zero target count for empty results", () => {
	const text = summarize({
		schemaVersion: 1,
		repository: { path: "/repo", ref: "main", currentHead: "abc" },
		query: { file: "new.go", maxCommits: 10, limit: 2, minCoChanges: 3 },
		index: { status: "fresh", indexedHead: "abc", analyzedCommits: 10 },
		results: [],
		exclusions: { bulkCommits: 0, ignoredFiles: 0 },
		warnings: [],
	});
	assert.match(text, /Target: new.go\n/);
	assert.doesNotMatch(text, /0 historical changes/);
});
