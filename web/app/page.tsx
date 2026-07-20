import Link from 'next/link'
import { ArrowRight, Bot, Braces, Database, GitCommitHorizontal, Search, ShieldCheck, Sparkles } from 'lucide-react'
import CopyCommand from '@/components/copy-command'
import SiteFooter from '@/components/site-footer'
import SiteHeader from '@/components/site-header'

const evidence = [
  { path: 'internal/git/models.go', score: '.857', changes: '18' },
  { path: 'internal/git/repository.go', score: '.720', changes: '12' },
  { path: 'internal/index/store.go', score: '.514', changes: '9' },
]

export default function Home() {
  return (
    <>
      <SiteHeader />
      <main>
        <section className="hero shell">
          <div className="hero-copy">
            <div className="eyebrow"><Sparkles size={14} /> Git history, made useful</div>
            <h1>See the files your next change might <em>miss.</em></h1>
            <p className="hero-lede">histui turns repeated Git co-changes into focused evidence for refactors, bug hunts, architecture work, and coding agents—without sending your history anywhere.</p>
            <div className="hero-actions">
              <Link className="button primary" href="/docs">Get started <ArrowRight size={17} /></Link>
              <a className="button secondary" href="https://github.com/Rounak-stha/histui">View source</a>
            </div>
            <CopyCommand command="go install github.com/Rounak-stha/histui/cmd/histui@latest" />
          </div>

          <div className="hero-visual" aria-label="Example histui coupling result">
            <div className="window-bar"><span /><span /><span /><b>history / coupling</b></div>
            <div className="query-line"><span>target</span> internal/git/cli_repo.go</div>
            <div className="signal-header"><span>historically related</span><span>score · co-changes</span></div>
            {evidence.map((item, index) => (
              <div className="evidence-row" key={item.path}>
                <span className="rank">0{index + 1}</span>
                <div><strong>{item.path}</strong><i style={{ width: `${Number(item.score) * 100}%` }} /></div>
                <code>{item.score} · {item.changes}</code>
              </div>
            ))}
            <div className="window-note"><span className="fresh-dot" /> fresh index · 1,000 commits · 4 bulk commits excluded</div>
          </div>
        </section>

        <section className="proof-strip">
          <div className="shell proof-grid">
            <div><b>&lt; 1s</b><span>cached targeted query</span></div>
            <div><b>100%</b><span>local processing</span></div>
            <div><b>SQLite</b><span>portable history index</span></div>
            <div><b>JSON v1</b><span>stable agent contract</span></div>
          </div>
        </section>

        <section className="section shell" id="why">
          <div className="section-heading">
            <span className="kicker">The missing signal</span>
            <h2>Static analysis shows today.<br />History shows what moves together.</h2>
            <p>Repeated co-change is not proof of dependency. It is a practical clue that helps you find tests, schemas, migrations, configuration, and boundaries before editing.</p>
          </div>
          <div className="feature-grid">
            <article className="feature feature-wide"><Search /><span>01</span><h3>Targeted coupling</h3><p>Ask one small question: which files repeatedly changed with this file? Results are bounded, ranked, and deterministic.</p><div className="mini-code">histui query coupling . --file src/auth.go</div></article>
            <article className="feature"><Database /><span>02</span><h3>Index once, query fast</h3><p>Stream history into a local SQLite index, then incrementally refresh only fast-forward commits.</p></article>
            <article className="feature"><Bot /><span>03</span><h3>Built for agents</h3><p>A typed Pi tool and Agent Skill provide compact context without flooding the model window.</p></article>
            <article className="feature"><ShieldCheck /><span>04</span><h3>Private by default</h3><p>No service, telemetry pipeline, or history upload. Your repository evidence remains local.</p></article>
            <article className="feature feature-wide"><GitCommitHorizontal /><span>05</span><h3>Honest evidence</h3><p>Freshness, history windows, ignored files, merge policy, and excluded bulk commits are explicit in every machine-readable result.</p><div className="tags"><i>freshness: cached</i><i>bulk threshold: 200</i><i>minimum evidence: 3</i></div></article>
          </div>
        </section>

        <section className="workflow-section">
          <div className="shell workflow-grid">
            <div className="section-heading left"><span className="kicker">Three commands</span><h2>From repository to useful context.</h2><p>Index deliberately. Query narrowly. Validate every historical clue against current code and tests.</p></div>
            <div className="steps">
              <div className="step"><b>01</b><div><h3>Build a bounded index</h3><CopyCommand command="histui index . --max-commits 5000" /></div></div>
              <div className="step"><b>02</b><div><h3>Ask about one file</h3><CopyCommand command="histui query coupling . --file src/app.go --format json" /></div></div>
              <div className="step"><b>03</b><div><h3>Refresh after new commits</h3><CopyCommand command="histui index . --update" /></div></div>
            </div>
          </div>
        </section>

        <section className="agent-section shell">
          <div className="agent-card">
            <div><span className="kicker">Pi integration</span><h2>Historical context, on demand.</h2><p>Install the extension and skill together. Pi can query cached evidence before a substantial refactor, while avoiding hidden full-history work.</p><Link href="/docs#pi" className="text-link">Read the integration guide <ArrowRight size={16} /></Link></div>
            <div className="tool-call"><div><Braces size={16} /> git_history</div><pre>{`{
  "file": "internal/git/cli_repo.go",
  "maxCommits": 1000,
  "freshness": "cached",
  "ignore": ["vendor", "*.md"]
}`}</pre></div>
          </div>
        </section>

        <section className="final-cta shell">
          <span className="kicker">Know before you change</span>
          <h2>Your Git history already documented the architecture.</h2>
          <p>histui makes that evidence searchable, bounded, and useful.</p>
          <div className="hero-actions"><Link className="button primary" href="/docs">Read the docs <ArrowRight size={17} /></Link><a className="button secondary" href="https://github.com/Rounak-stha/histui">Star on GitHub</a></div>
        </section>
      </main>
      <SiteFooter />
    </>
  )
}
