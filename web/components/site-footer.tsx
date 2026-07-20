import Link from 'next/link'

export default function SiteFooter() {
  return (
    <footer className="site-footer">
      <div className="shell footer-grid">
        <div>
          <div className="wordmark"><span className="mark">h</span><span>histui</span></div>
          <p>Local Git history as focused engineering context.</p>
        </div>
        <div className="footer-links">
          <Link href="/docs">Documentation</Link>
          <a href="https://github.com/Rounak-stha/histui">Source</a>
          <a href="https://github.com/Rounak-stha/histui/issues">Issues</a>
        </div>
      </div>
      <div className="shell footer-bottom">
        <span>MIT licensed · history stays on your machine</span>
        <span>Built for humans and coding agents</span>
      </div>
    </footer>
  )
}
