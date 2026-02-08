'use client'

import { useEffect, useState } from 'react'

interface TerminalLine {
  text: string
  type: 'input' | 'output' | 'header'
}

export default function Terminal() {
  const [mounted, setMounted] = useState(false)
  const [lines, setLines] = useState<TerminalLine[]>([
    { text: '$ histui --coupling', type: 'input' },
  ])
  const [isAnimating, setIsAnimating] = useState(true)

  useEffect(() => {
    setMounted(true)
  }, [])

  useEffect(() => {
    if (!isAnimating) return

    const outputLines: TerminalLine[] = [
      {
        text: 'Analyzing repository structure...',
        type: 'output',
      },
      {
        text: '',
        type: 'output',
      },
      {
        text: 'FILE COUPLING ANALYSIS',
        type: 'header',
      },
      {
        text: '────────────────────────────────────────',
        type: 'output',
      },
      {
        text: 'hooks/useAuth.ts          ←→  contexts/AuthContext.ts       [94%]',
        type: 'output',
      },
      {
        text: 'pages/Dashboard.tsx       ←→  components/Sidebar.tsx        [87%]',
        type: 'output',
      },
      {
        text: 'utils/api.ts              ←→  services/DataFetcher.ts      [81%]',
        type: 'output',
      },
      {
        text: 'types/index.ts            ←→  schemas/validation.ts        [76%]',
        type: 'output',
      },
      {
        text: 'lib/db.ts                 ←→  models/User.ts               [73%]',
        type: 'output',
      },
      {
        text: '────────────────────────────────────────',
        type: 'output',
      },
      {
        text: 'Coupling score: 8.2/10',
        type: 'output',
      },
      {
        text: 'Refactor opportunities: 3 high-impact areas detected',
        type: 'output',
      },
    ]

    let lineIndex = 0
    const interval = setInterval(() => {
      if (lineIndex < outputLines.length) {
        setLines((prev) => [...prev, outputLines[lineIndex]])
        lineIndex++
      } else {
        setIsAnimating(false)
        clearInterval(interval)
      }
    }, 200)

    return () => clearInterval(interval)
  }, [isAnimating])

  if (!mounted) {
    return (
      <section className="min-h-screen w-full bg-background py-20 px-4 border-b border-border">
        <div className="max-w-5xl mx-auto">
          <div className="bg-card rounded-none border border-border overflow-hidden">
            <div className="bg-secondary px-4 py-3 flex items-center gap-3 border-b border-border">
              <div className="flex gap-2">
                <div className="w-3 h-3 rounded-full bg-red-500"></div>
                <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
                <div className="w-3 h-3 rounded-full bg-green-500"></div>
              </div>
              <span className="text-sm text-muted-foreground flex-1 text-center font-mono">
                histui-analysis.sh
              </span>
            </div>
            <div className="p-6 font-mono text-sm bg-card min-h-96" />
          </div>
        </div>
      </section>
    )
  }

  return (
    <section className="min-h-screen w-full bg-background py-20 px-4 border-b border-border">
      <div className="max-w-5xl mx-auto">
        <div className="bg-card rounded-none border border-border overflow-hidden">
          {/* Terminal Header */}
          <div className="bg-secondary px-4 py-3 flex items-center gap-3 border-b border-border">
            <div className="flex gap-2">
              <div className="w-3 h-3 rounded-full bg-red-500"></div>
              <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
              <div className="w-3 h-3 rounded-full bg-green-500"></div>
            </div>
            <span className="text-sm text-muted-foreground flex-1 text-center font-mono">
              histui-analysis.sh
            </span>
          </div>

          {/* Terminal Content */}
          <div className="p-6 font-mono text-sm bg-card min-h-96 space-y-1">
            {lines?.map((line, idx) => {
              if (!line) return null
              return (
                <div
                  key={idx}
                  className={`animate-in fade-in duration-300 ${
                    line.type === 'input'
                      ? 'text-accent font-semibold'
                      : line.type === 'header'
                        ? 'text-accent font-bold'
                        : 'text-foreground'
                  }`}
                >
                  {line.type === 'input' ? '$ ' : ''}
                  {line.text}
                  {line.type === 'input' && isAnimating && (
                    <span className="animate-pulse">▌</span>
                  )}
                </div>
              )
            })}
            {isAnimating && (
              <div className="text-foreground animate-pulse">
                $ ▌
              </div>
            )}
          </div>
        </div>

        <div className="mt-12 grid md:grid-cols-3 gap-8">
          <div className="border border-border p-6 bg-card/50">
            <div className="text-3xl font-bold text-accent mb-2">94%</div>
            <div className="text-sm text-muted-foreground">
              Highest Coupling
            </div>
            <div className="text-xs text-muted-foreground mt-2 font-mono">
              useAuth ←→ AuthContext
            </div>
          </div>
          <div className="border border-border p-6 bg-card/50">
            <div className="text-3xl font-bold text-accent mb-2">8.2/10</div>
            <div className="text-sm text-muted-foreground">
              Overall Coupling Score
            </div>
            <div className="text-xs text-muted-foreground mt-2">
              Moderate - 3 refactor opportunities
            </div>
          </div>
          <div className="border border-border p-6 bg-card/50">
            <div className="text-3xl font-bold text-accent mb-2">12.4ms</div>
            <div className="text-sm text-muted-foreground">
              Analysis Time
            </div>
            <div className="text-xs text-muted-foreground mt-2">
              Fast scanning & indexing
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
