import assert from "node:assert/strict";
import test from "node:test";
import { createHistuiTool, type HistuiResponse } from "./histui.ts";

const valid: HistuiResponse = {
	schemaVersion: 1,
	repository: { path: "/repo", ref: "main", currentHead: "abc" },
	query: { file: "target.go", maxCommits: 1000, limit: 20, minCoChanges: 3 },
	index: { status: "fresh", indexedHead: "abc", analyzedCommits: 3 },
	results: [], exclusions: { bulkCommits: 0, ignoredFiles: 0 }, warnings: [],
};

function harness(histuiResult: { stdout: string; stderr?: string; code?: number; killed?: boolean }) {
	const calls: Array<{ command: string; args: string[]; options: unknown }> = [];
	const pi = {
		exec: async (command: string, args: string[], options: unknown) => {
			calls.push({ command, args, options });
			if (command === "git") return { stdout: "/repo\n", stderr: "", code: 0, killed: false };
			return { stderr: "", code: 0, killed: false, ...histuiResult };
		},
	} as any;
	return { tool: createHistuiTool(pi), calls };
}

async function execute(tool: ReturnType<typeof createHistuiTool>, params: Record<string, unknown> = {}) {
	return tool.execute("call", { file: "target.go", ...params } as any, new AbortController().signal, undefined, { cwd: "/repo/subdir" } as any);
}

test("execute constructs a direct bounded query and parses details", async () => {
	const { tool, calls } = harness({ stdout: JSON.stringify(valid) });
	const result = await execute(tool, { maxCommits: 50, limit: 5, freshness: "strict", includeMerges: true, ignore: ["vendor", "generated/**"] });
	assert.equal(calls.length, 2);
	assert.equal(calls[1].command, "histui");
	assert.deepEqual(calls[1].args, ["query", "coupling", "/repo", "--file", "target.go", "--max-commits", "50", "--limit", "5", "--freshness", "strict", "--format", "json", "--include-merges", "--ignore", "vendor", "--ignore", "generated/**"]);
	assert.ok(!calls[1].args.includes("index"));
	assert.deepEqual(result.details, valid);
});

test("execute rejects malformed output and incompatible schemas", async () => {
	await assert.rejects(() => execute(harness({ stdout: "not-json" }).tool), /malformed JSON/);
	await assert.rejects(() => execute(harness({ stdout: JSON.stringify({ ...valid, schemaVersion: 2 }) }).tool), /Unsupported histui schema 2/);
	await assert.rejects(() => execute(harness({ stdout: JSON.stringify({ schemaVersion: 1, results: [] }) }).tool), /missing required schema fields/);
});

test("execute reports nonzero exits and propagates the abort signal", async () => {
	const controller = new AbortController();
	const { tool, calls } = harness({ stdout: "", stderr: "history index not found", code: 1 });
	await assert.rejects(() => tool.execute("call", { file: "target.go" }, controller.signal, undefined, { cwd: "/repo" } as any), /Build a bounded index/);
	assert.equal((calls[0].options as any).signal, controller.signal);
	assert.equal((calls[1].options as any).signal, controller.signal);
});
