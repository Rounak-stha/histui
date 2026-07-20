import type { Metadata } from 'next'
import Link from 'next/link'
import { AlertTriangle, ArrowUpRight, CheckCircle2, Info } from 'lucide-react'
import CopyCommand from '@/components/copy-command'
import SiteFooter from '@/components/site-footer'
import SiteHeader from '@/components/site-header'

export const metadata: Metadata = {
  title: 'Documentation · histui',
  description: 'Install, index, and query Git history with histui. CLI, JSON schema, freshness, ignore patterns, and Pi integration.',
}

const sections = [
  ['start', 'Quick start'], ['indexing', 'Indexing'], ['querying', 'Coupling query'], ['ignore', 'Ignore patterns'],
  ['freshness', 'Freshness'], ['json', 'JSON output'], ['pi', 'Pi integration'], ['interpret', 'Interpretation'], ['troubleshooting', 'Troubleshooting'],
]

function Code({ children }: { children: string }) {
  return <pre className="docs-code"><code>{children}</code></pre>
}

export default function Docs() {
  return (
    <>
      <SiteHeader />
      <main className="docs-shell shell">
        <aside className="docs-sidebar">
          <p>Documentation</p>
          <nav>{sections.map(([id, label]) => <a key={id} href={`#${id}`}>{label}</a>)}</nav>
          <a className="edit-link" href="https://github.com/Rounak-stha/histui">View source <ArrowUpRight size={13} /></a>
        </aside>

        <article className="docs-content">
          <header className="docs-hero">
            <span className="kicker">Documentation</span>
            <h1>Git history as focused engineering context.</h1>
            <p>Everything needed to build a local index, run targeted coupling queries, consume stable JSON, and connect histui to Pi.</p>
          </header>

          <section id="start">
            <h2>Quick start</h2>
            <p>Build histui from source, create a bounded local index, and ask which files changed repeatedly with a target.</p>
            <CopyCommand label="Install" command="go install github.com/Rounak-stha/histui/cmd/histui@latest" />
            <CopyCommand label="Index" command="histui index . --max-commits 5000" />
            <CopyCommand label="Query" command="histui query coupling . --file internal/git/cli_repo.go --format json" />
            <div className="callout info"><Info /><div><b>Local by design</b><p>The index contains paths and commit metadata, lives in your platform cache directory, and never leaves your machine unless you export it.</p></div></div>
          </section>

          <section id="indexing">
            <h2>Indexing</h2>
            <p>Initial indexing streams a bounded Git history into an atomic SQLite transaction. Queries never silently create a missing index.</p>
            <Code>{`# Build or replace an index
histui index . --max-commits 5000

# Incrementally append fast-forward commits
histui index . --update

# Inspect freshness and policy
histui index status . --format json

# Keep a separate index for another ref
histui index . --branch release --index-path ~/.cache/histui/release.sqlite`}</Code>
            <h3>Index options</h3>
            <div className="option-table">
              <div><code>--max-commits</code><span>Maximum commits indexed. Default: 5000.</span></div>
              <div><code>--max-files-per-commit</code><span>Mark larger commits as bulk noise. Default: 200.</span></div>
              <div><code>--branch</code><span>Index a specific branch or ref.</span></div>
              <div><code>--include-merges</code><span>Include merge commits; query policy must match.</span></div>
              <div><code>--index-path</code><span>Override the platform cache location.</span></div>
            </div>
          </section>

          <section id="querying">
            <h2>Targeted coupling query</h2>
            <p>A query finds indexed commits containing the target, counts other files in those commits, and ranks them by historical coupling.</p>
            <Code>{`histui query coupling . \\
  --file src/auth/session.go \\
  --max-commits 1000 \\
  --limit 20 \\
  --min-co-changes 3 \\
  --freshness cached \\
  --format json`}</Code>
            <div className="option-table">
              <div><code>--file</code><span>Required repository-relative target file.</span></div>
              <div><code>--limit</code><span>Maximum related files returned. Default: 20.</span></div>
              <div><code>--min-co-changes</code><span>Minimum evidence required. Default: 3.</span></div>
              <div><code>--max-commits</code><span>Bounded query window. Default: 1000.</span></div>
              <div><code>--format</code><span><code>human</code> or stable <code>json</code>.</span></div>
            </div>
          </section>

          <section id="ignore">
            <h2>Ignore files and folders</h2>
            <p>Always quote shell glob patterns. Otherwise your shell expands them into positional arguments before histui starts.</p>
            <Code>{`histui --coupling \\
  --ignore '**/i18n/**' \\
  --ignore 'package.json' \\
  --ignore '*.d.ts'

# Or as one quoted, comma-separated value
histui --coupling --ignore '**/i18n/**,package.json,*.d.ts'`}</Code>
            <div className="pattern-grid">
              <div><code>*.md</code><span>matching filename at any depth</span></div>
              <div><code>vendor</code><span>folder and complete subtree</span></div>
              <div><code>vendor/</code><span>same recursive folder behavior</span></div>
              <div><code>docs/*</code><span>direct children only</span></div>
              <div><code>docs/**</code><span>all descendants</span></div>
              <div><code>**/i18n/**</code><span>matching folder at any depth</span></div>
            </div>
            <div className="callout warning"><AlertTriangle /><div><b>Why “accepts at most 1 arg” happens</b><p>Unquoted <code>**/i18n</code> or <code>*.d.ts</code> is expanded by Bash/Zsh into many filenames. Quote every pattern containing <code>*</code>, <code>?</code>, or brackets.</p></div></div>
          </section>

          <section id="freshness">
            <h2>Freshness policies</h2>
            <div className="policy-grid">
              <article><b>cached</b><p>Return immediately. Stale evidence is allowed and labeled.</p></article>
              <article><b>refresh</b><p>Append a bounded fast-forward range; otherwise return stale evidence with a warning.</p></article>
              <article><b>strict</b><p>Fail unless the index exactly represents the selected HEAD.</p></article>
            </div>
            <p>Diverged or rebased history must be rebuilt. An index built for another ref is rejected; use separate index paths for multiple refs.</p>
          </section>

          <section id="json">
            <h2>Stable JSON output</h2>
            <p>JSON mode writes exactly one untruncated object to stdout. Diagnostics go to stderr and failures return a nonzero status.</p>
            <Code>{`{
  "schemaVersion": 1,
  "repository": { "ref": "main", "currentHead": "def456" },
  "query": { "type": "file-coupling", "file": "src/auth.go" },
  "index": { "status": "fresh", "analyzedCommits": 1000 },
  "results": [{
    "path": "src/session.go",
    "coChanges": 18,
    "targetChanges": 25,
    "relatedFileChanges": 21,
    "score": 0.857,
    "strength": "critical"
  }],
  "exclusions": { "bulkCommits": 4, "ignoredFiles": 82 },
  "warnings": ["Historical coupling is evidence, not proof..."]
}`}</Code>
            <p>Consumers should reject unsupported schema versions and ignore unknown fields. See the <a href="https://github.com/Rounak-stha/histui/blob/main/docs/json-schema-v1.md">full schema contract</a>.</p>
          </section>

          <section id="pi">
            <h2>Pi coding-agent integration</h2>
            <p>The package bundles a typed <code>git_history</code> tool and an Agent Skill. Install the CLI first, then the package.</p>
            <Code>{`# From a local checkout
pi install /absolute/path/to/histui

# From Git
pi install git:github.com/Rounak-stha/histui

# Build the index explicitly before agent queries
histui index . --max-commits 5000`}</Code>
            <Code>{`git_history({
  "file": "internal/git/cli_repo.go",
  "maxCommits": 1000,
  "limit": 20,
  "freshness": "cached",
  "ignore": ["vendor", "generated/**", "*.md"]
})`}</Code>
            <div className="callout success"><CheckCircle2 /><div><b>No hidden expensive work</b><p>The extension applies cancellation and timeout, invokes the CLI without a shell, and never starts a missing full index automatically.</p></div></div>
          </section>

          <section id="interpret">
            <h2>Interpret the evidence</h2>
            <p>The v1 coupling score is:</p>
            <div className="formula">coChanges / min(targetChanges, relatedFileChanges)</div>
            <ul>
              <li>High score with many co-changes is stronger evidence.</li>
              <li>A high score from only three commits remains weak.</li>
              <li>No result may mean isolation, recent code, ignored files, or an insufficient window.</li>
              <li>Bulk commits are excluded because mass formatting and generated changes create noise.</li>
              <li>Correlation does not prove a source-level, runtime, or architectural dependency.</li>
            </ul>
          </section>

          <section id="troubleshooting">
            <h2>Troubleshooting</h2>
            <div className="faq">
              <details open><summary>History index not found</summary><p>Run <code>histui index . --max-commits 5000</code>. Queries intentionally do not create indexes.</p></details>
              <details><summary>Index is stale or diverged</summary><p>Use <code>histui index . --update</code> for fast-forward history. Rebuild without <code>--update</code> after a rebase or branch divergence.</p></details>
              <details><summary>Merge or bulk policy does not match</summary><p>Query flags must match index policy. Use the requested flag value or rebuild with the desired settings.</p></details>
              <details><summary>No related files returned</summary><p>Try a larger bounded window or lower <code>--min-co-changes</code>, then verify whether ignores or bulk commits removed evidence.</p></details>
            </div>
          </section>

          <div className="docs-next"><span>Next</span><Link href="https://github.com/Rounak-stha/histui">Explore the source <ArrowUpRight size={16} /></Link></div>
        </article>
      </main>
      <SiteFooter />
    </>
  )
}
