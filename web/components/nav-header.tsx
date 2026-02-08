export default function NavHeader() {
  return (
    <header className="fixed top-0 left-0 right-0 z-40 bg-background/80 backdrop-blur-md border-b border-border/50">
      <div className="max-w-6xl mx-auto px-4 py-4 flex items-center justify-between">
        {/* Logo */}
        <div className="text-lg font-bold text-foreground font-mono">
          histui
        </div>

        {/* Center navigation */}
        <nav className="hidden md:flex items-center gap-8 text-sm">
          <a
            href="/"
            className="text-muted-foreground hover:text-foreground transition-colors"
          >
            Home
          </a>
          <a
            href="https://github.com/Rounak-stha/histui"
            className="text-muted-foreground hover:text-foreground transition-colors"
          >
            Github
          </a>
        </nav>
      </div>
    </header>
  )
}
