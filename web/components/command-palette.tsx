'use client'

import { useEffect, useRef } from 'react'

interface Command {
  id: string
  label: string
  description: string
  icon: string
}

interface CommandPaletteProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export default function CommandPalette({
  open,
  onOpenChange,
}: CommandPaletteProps) {
  const inputRef = useRef<HTMLInputElement>(null)

  const commands: Command[] = [
    {
      id: 'analyze',
      label: 'Analyze Repository',
      description: 'Run coupling analysis on current repo',
      icon: '📊',
    },
    {
      id: 'docs',
      label: 'Documentation',
      description: 'View complete histui documentation',
      icon: '📚',
    },
    {
      id: 'github',
      label: 'GitHub Repository',
      description: 'Visit the GitHub repository',
      icon: '🔗',
    },
    {
      id: 'install',
      label: 'Installation Guide',
      description: 'How to install histui',
      icon: '⚙️',
    },
    {
      id: 'api',
      label: 'API Reference',
      description: 'Explore the histui API',
      icon: '🔧',
    },
    {
      id: 'examples',
      label: 'Examples',
      description: 'View real-world examples',
      icon: '💡',
    },
  ]

  useEffect(() => {
    if (open && inputRef.current) {
      inputRef.current.focus()
    }
  }, [open])

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && open) {
        onOpenChange(false)
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [open, onOpenChange])

  if (!open) return null

  return (
    <>
      {/* Overlay */}
      <div
        className="fixed inset-0 z-40 bg-black/50 backdrop-blur-sm"
        onClick={() => onOpenChange(false)}
      />

      {/* Command Palette */}
      <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 z-50 w-full max-w-2xl">
        <div className="bg-card border border-border rounded-lg shadow-2xl overflow-hidden">
          {/* Input */}
          <div className="border-b border-border p-4 flex items-center gap-3">
            <span className="text-muted-foreground">$</span>
            <input
              ref={inputRef}
              type="text"
              placeholder="Type a command..."
              className="flex-1 bg-transparent text-foreground outline-none placeholder:text-muted-foreground font-mono"
            />
            <span className="text-xs text-muted-foreground">
              ESC to close
            </span>
          </div>

          {/* Commands */}
          <div className="max-h-96 overflow-y-auto">
            {commands.map((command, idx) => (
              <button
                key={command.id}
                onClick={() => {
                  onOpenChange(false)
                }}
                className="w-full px-4 py-3 text-left hover:bg-secondary/50 border-b border-border/50 last:border-b-0 transition-colors flex items-start gap-3 group"
              >
                <span className="text-lg mt-1">{command.icon}</span>
                <div className="flex-1">
                  <div className="text-foreground font-semibold group-hover:text-accent transition-colors">
                    {command.label}
                  </div>
                  <div className="text-sm text-muted-foreground">
                    {command.description}
                  </div>
                </div>
                <kbd className="hidden sm:flex text-xs text-muted-foreground bg-secondary rounded px-2 py-1 font-mono h-fit">
                  ⏎
                </kbd>
              </button>
            ))}
          </div>

          {/* Footer */}
          <div className="bg-secondary/30 px-4 py-3 border-t border-border text-xs text-muted-foreground font-mono">
            Type to filter • ↑↓ to navigate • ⏎ to select
          </div>
        </div>
      </div>
    </>
  )
}
