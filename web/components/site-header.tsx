'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { Github, Menu, X } from 'lucide-react'
import { useState } from 'react'

const links = [
  { href: '/', label: 'Overview' },
  { href: '/docs', label: 'Docs' },
]

export default function SiteHeader() {
  const pathname = usePathname()
  const [open, setOpen] = useState(false)

  return (
    <header className="site-header">
      <div className="shell nav-shell">
        <Link href="/" className="wordmark" aria-label="histui home">
          <span className="mark">h</span>
          <span>histui</span>
          <span className="version">alpha</span>
        </Link>

        <nav className="desktop-nav" aria-label="Main navigation">
          {links.map((link) => (
            <Link key={link.href} href={link.href} className={pathname === link.href ? 'active' : ''}>
              {link.label}
            </Link>
          ))}
          <a className="github-link" href="https://github.com/Rounak-stha/histui" target="_blank" rel="noreferrer">
            <Github size={16} /> GitHub
          </a>
        </nav>

        <button className="menu-button" onClick={() => setOpen(!open)} aria-expanded={open} aria-label="Toggle navigation">
          {open ? <X /> : <Menu />}
        </button>
      </div>
      {open && (
        <nav className="mobile-nav" aria-label="Mobile navigation">
          {links.map((link) => <Link key={link.href} href={link.href} onClick={() => setOpen(false)}>{link.label}</Link>)}
          <a href="https://github.com/Rounak-stha/histui">GitHub</a>
        </nav>
      )}
    </header>
  )
}
