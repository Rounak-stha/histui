'use client'

import { useEffect, useRef, useState } from 'react'

export default function ScrollNarrative() {
  const sectionRef = useRef<HTMLDivElement>(null)
  const [scrollProgress, setScrollProgress] = useState(0)

  useEffect(() => {
    const handleScroll = () => {
      if (!sectionRef.current) return

      const rect = sectionRef.current.getBoundingClientRect()
      const sectionStart = rect.top
      const sectionEnd = rect.bottom

      const windowHeight = window.innerHeight

      let progress = 0
      if (sectionStart < windowHeight && sectionEnd > 0) {
        progress =
          (windowHeight - sectionStart) /
          (windowHeight + rect.height) *
          100
        progress = Math.max(0, Math.min(100, progress))
      }

      setScrollProgress(progress)
    }

    window.addEventListener('scroll', handleScroll)
    return () => window.removeEventListener('scroll', handleScroll)
  }, [])

  const sections = [
    {
      step: 1,
      title: 'Tangled Dependencies',
      description: 'Your codebase is a web of interconnected files changing together.',
      visual: 'chaos',
    },
    {
      step: 2,
      title: 'Run Analysis',
      description:
        'histui scans your repository and identifies file coupling patterns.',
      visual: 'scan',
    },
    {
      step: 3,
      title: 'See Connections',
      description:
        'Visualize which files are strongly coupled and why they change together.',
      visual: 'connections',
    },
    {
      step: 4,
      title: 'Refactor Confidently',
      description:
        'Make informed decisions about code organization and refactoring opportunities.',
      visual: 'organized',
    },
  ]

  return (
    <section
      ref={sectionRef}
      className="min-h-screen bg-background py-20 px-4 relative"
    >
      <div className="max-w-6xl mx-auto">
        {sections.map((section, idx) => (
          <div
            key={idx}
            className="mb-32 last:mb-0"
            style={{
              opacity: Math.max(
                0,
                Math.min(
                  1,
                  (scrollProgress - idx * 25) / 25 + 0.3
                )
              ),
              transform: `translateY(${Math.max(0, (3 - idx) * (100 - scrollProgress) * 0.05)}px)`,
            }}
          >
            <div className="grid md:grid-cols-2 gap-12 items-center">
              <div>
                <div className="text-5xl font-bold text-accent mb-6 font-light">
                  {section.step}
                </div>
                <h2 className="text-4xl font-bold text-foreground mb-6">
                  {section.title}
                </h2>
                <p className="text-lg text-muted-foreground leading-relaxed">
                  {section.description}
                </p>
              </div>
              <div 
                className="border border-border rounded-none aspect-square bg-card/30 flex items-center justify-center relative overflow-hidden"
                suppressHydrationWarning
              >
                {section.visual === 'chaos' && (
                  <ChaosVisual progress={scrollProgress} />
                )}
                {section.visual === 'scan' && (
                  <ScanVisual progress={scrollProgress} />
                )}
                {section.visual === 'connections' && (
                  <ConnectionsVisual progress={scrollProgress} />
                )}
                {section.visual === 'organized' && (
                  <OrganizedVisual progress={scrollProgress} />
                )}
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Progress indicator */}
      <div className="fixed right-8 top-1/2 -translate-y-1/2 hidden lg:flex flex-col gap-4">
        {sections.map((_, idx) => (
          <div
            key={idx}
            className={`h-2 w-2 rounded-full transition-all ${
              scrollProgress > idx * 25
                ? 'bg-accent h-3 w-3'
                : 'bg-muted'
            }`}
          ></div>
        ))}
      </div>
    </section>
  )
}

function ChaosVisual({ progress }: { progress: number }) {
  return (
    <svg viewBox="0 0 200 200" className="w-32 h-32">
      {Array.from({ length: 8 }).map((_, i) => {
        const angle = (i * 45 * Math.PI) / 180
        const x = 100 + Math.cos(angle) * 40
        const y = 100 + Math.sin(angle) * 40
        return (
          <g key={i}>
            <circle cx={x} cy={y} r="8" fill="rgba(140, 200, 100, 0.5)" />
            <line
              x1="100"
              y1="100"
              x2={x}
              y2={y}
              stroke="rgba(140, 200, 100, 0.3)"
              strokeWidth="1"
              opacity={(progress + i * 10) % 100 / 100}
            />
          </g>
        )
      })}
    </svg>
  )
}

function ScanVisual({ progress }: { progress: number }) {
  return (
    <svg viewBox="0 0 200 200" className="w-32 h-32">
      <circle
        cx="100"
        cy="100"
        r="50"
        fill="none"
        stroke="rgba(140, 200, 100, 0.5)"
        strokeWidth="2"
        strokeDasharray="314"
        strokeDashoffset={314 - (progress % 100) * 3.14}
      />
      <circle
        cx="100"
        cy="100"
        r="30"
        fill="none"
        stroke="rgba(140, 200, 100, 0.3)"
        strokeWidth="1"
      />
    </svg>
  )
}

function ConnectionsVisual({ progress }: { progress: number }) {
  const offset = (progress % 100) * 2
  return (
    <svg viewBox="0 0 200 200" className="w-32 h-32">
      {Array.from({ length: 5 }).map((_, i) => {
        const angle = (i * 72 * Math.PI) / 180
        const x = Math.round((100 + Math.cos(angle) * 50) * 100) / 100
        const y = Math.round((100 + Math.sin(angle) * 50) * 100) / 100
        return (
          <g key={i}>
            <circle cx={x} cy={y} r="6" fill="rgba(140, 200, 100, 0.7)" />
            {Array.from({ length: 4 }).map((_, j) => {
              if (j <= i) {
                const nextAngle = ((i + 1 + j) * 72 * Math.PI) / 180
                const nextX = Math.round((100 + Math.cos(nextAngle) * 50) * 100) / 100
                const nextY = Math.round((100 + Math.sin(nextAngle) * 50) * 100) / 100
                return (
                  <line
                    key={`${i}-${j}`}
                    x1={x}
                    y1={y}
                    x2={nextX}
                    y2={nextY}
                    stroke="rgba(140, 200, 100, 0.4)"
                    strokeWidth="1"
                    opacity={(offset + i * j) % 100 / 100}
                  />
                )
              }
            })}
          </g>
        )
      })}
    </svg>
  )
}

function OrganizedVisual({ progress }: { progress: number }) {
  return (
    <svg viewBox="0 0 200 200" className="w-32 h-32">
      <defs>
        <pattern
          id="grid"
          width="20"
          height="20"
          patternUnits="userSpaceOnUse"
        >
          <rect
            x="0"
            y="0"
            width="20"
            height="20"
            fill="none"
            stroke="rgba(140, 200, 100, 0.2)"
            strokeWidth="0.5"
          />
        </pattern>
      </defs>
      <rect x="0" y="0" width="200" height="200" fill="url(#grid)" />
      {Array.from({ length: 9 }).map((_, i) => {
        const x = (i % 3) * 60 + 30
        const y = Math.floor(i / 3) * 60 + 30
        const opacity = Math.min(1, (progress - 75) / 25)
        return (
          <g key={i} opacity={opacity}>
            <rect
              x={x - 15}
              y={y - 15}
              width="30"
              height="30"
              fill="none"
              stroke="rgba(140, 200, 100, 0.6)"
              strokeWidth="1"
            />
            <circle cx={x} cy={y} r="4" fill="rgba(140, 200, 100, 0.8)" />
          </g>
        )
      })}
    </svg>
  )
}
