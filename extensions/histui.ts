import type { ExtensionAPI, ToolDefinition } from "@earendil-works/pi-coding-agent";
import { Type, type Static } from "typebox";
import { isAbsolute, relative, resolve, sep } from "node:path";

const schema = Type.Object({
	file: Type.String({ description: "Repository-relative file to investigate" }),
	maxCommits: Type.Optional(Type.Integer({ minimum: 1, maximum: 100000, description: "Bounded indexed history window" })),
	limit: Type.Optional(Type.Integer({ minimum: 1, maximum: 50, description: "Maximum related files" })),
	freshness: Type.Optional(Type.Union([
		Type.Literal("cached"),
		Type.Literal("refresh"),
		Type.Literal("strict"),
	])),
	includeMerges: Type.Optional(Type.Boolean()),
	ignore: Type.Optional(Type.Array(Type.String(), { maxItems: 100, description: "File/folder ignore patterns, for example *.md, vendor, or generated/**" })),
});

export type HistuiToolInput = Static<typeof schema>;

type RelatedFile = {
	path: string;
	coChanges: number;
	targetChanges: number;
	relatedFileChanges: number;
	score: number;
	strength: string;
};

export type HistuiResponse = {
	schemaVersion: number;
	repository: { path: string; ref: string; currentHead: string };
	query: { file: string; maxCommits: number; limit: number; minCoChanges: number };
	index: { status: string; indexedHead: string; analyzedCommits: number; createdAt?: string; updatedAt?: string };
	results: RelatedFile[];
	exclusions: { bulkCommits: number; ignoredFiles: number };
	warnings: string[];
};

export function repositoryRelativeFile(cwd: string, input: string): string {
	const cleaned = input.startsWith("@") ? input.slice(1) : input;
	if (!cleaned || isAbsolute(cleaned)) throw new Error("file must be repository-relative");
	const absolute = resolve(cwd, cleaned);
	const rel = relative(cwd, absolute);
	if (!rel || rel === ".." || rel.startsWith(`..${sep}`) || isAbsolute(rel)) {
		throw new Error("file must remain within the repository");
	}
	return rel.split(sep).join("/");
}

export function summarize(data: HistuiResponse): string {
	const targetChanges = data.results[0]?.targetChanges;
	const lines = [
		`History index: ${data.index.status}; ${data.index.analyzedCommits.toLocaleString()} commits analyzed.`,
		targetChanges === undefined
			? `Target: ${data.query.file}`
			: `Target: ${data.query.file} (${targetChanges} historical changes)`,
	];
	if (data.results.length === 0) lines.push("No related files met the evidence threshold.");
	else {
		lines.push("Related files:");
		for (const [index, result] of data.results.entries()) {
			lines.push(`${index + 1}. ${result.path} — score ${result.score.toFixed(3)}, ${result.coChanges} co-changes`);
		}
	}
	if (data.exclusions.bulkCommits > 0) lines.push(`Excluded ${data.exclusions.bulkCommits} bulk commits.`);
	for (const warning of data.warnings) lines.push(`Warning: ${warning}`);
	return lines.join("\n");
}

export function explainFailure(stderr: string, code: number | null, killed: boolean): string {
	const message = stderr.trim();
	if (/history index not found/i.test(message)) return `${message}\nBuild a bounded index explicitly; the tool will not start a full index automatically.`;
	if (/command not found|ENOENT|executable.*not found/i.test(message)) return "histui executable was not found. Install the histui CLI and ensure it is on PATH.";
	if (killed) return "histui timed out or was cancelled before producing a result.";
	return message || `histui exited with status ${code ?? "unknown"}`;
}

export function createHistuiTool(pi: ExtensionAPI): ToolDefinition<typeof schema, HistuiResponse> {
	return {
		name: "git_history",
		label: "Git History",
		description: "Query indexed Git history for files repeatedly changed with a target. Use as supporting evidence before substantial refactors, cross-file bug fixes, or work in unfamiliar mature code. Do not use for trivial edits or treat correlation as proof. Requires an explicitly built histui index.",
		promptSnippet: "Find files historically coupled to a target file using the local histui index",
		promptGuidelines: [
			"Use git_history near the start of substantial refactors or cross-file investigations when hidden relationships may change which files should be inspected.",
			"Do not repeatedly call git_history for trivial edits, and inspect current code and tests before acting on historical correlation.",
		],
		parameters: schema,
		async execute(_toolCallId, params, signal, onUpdate, ctx) {
			let rootResult;
			try {
				rootResult = await pi.exec("git", ["-C", ctx.cwd, "rev-parse", "--show-toplevel"], { signal, timeout: 3_000 });
			} catch (error) {
				throw new Error(`Unable to locate the Git repository: ${error instanceof Error ? error.message : String(error)}`);
			}
			if (rootResult.code !== 0 || !rootResult.stdout.trim()) throw new Error("git_history requires a Git repository");
			const repositoryRoot = rootResult.stdout.trim();
			const file = repositoryRelativeFile(repositoryRoot, params.file);
			const args = [
				"query", "coupling", repositoryRoot,
				"--file", file,
				"--max-commits", String(params.maxCommits ?? 1000),
				"--limit", String(params.limit ?? 20),
				"--freshness", params.freshness ?? "cached",
				"--format", "json",
			];
			if (params.includeMerges) args.push("--include-merges");
			for (const pattern of params.ignore ?? []) {
				if (!pattern.trim()) throw new Error("ignore patterns must not be empty");
				args.push("--ignore", pattern);
			}
			let result;
			try {
				result = await pi.exec(process.env.HISTUI_BIN || "histui", args, { signal, timeout: 15_000 });
			} catch (error) {
				throw new Error(error instanceof Error && /ENOENT/i.test(error.message)
					? "histui executable was not found. Install the histui CLI and ensure it is on PATH."
					: `Failed to run histui: ${error instanceof Error ? error.message : String(error)}`);
			}
			if (result.code !== 0) throw new Error(explainFailure(result.stderr, result.code, result.killed));
			let data: HistuiResponse;
			try {
				data = JSON.parse(result.stdout) as HistuiResponse;
			} catch {
				throw new Error("histui returned malformed JSON; verify that the CLI supports query coupling --format json");
			}
			if (data.schemaVersion !== 1) throw new Error(`Unsupported histui schema ${data.schemaVersion}; this extension requires schema 1.`);
			if (!data.repository || !data.query || !Array.isArray(data.results) || !data.index || !data.exclusions || !Array.isArray(data.warnings)) {
				throw new Error("histui JSON is missing required schema fields");
			}
			return { content: [{ type: "text", text: summarize(data) }], details: data };
		},
	};
}

export default function histuiExtension(pi: ExtensionAPI) {
	pi.registerTool(createHistuiTool(pi));
}
