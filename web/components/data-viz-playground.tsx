'use client'

import { useEffect, useRef, useState } from 'react'

interface Node {
  id: number
  x: number
  y: number
  vx: number
  vy: number
  file: string
  coupling: number
}

interface Link {
  source: number
  target: number
  strength: number
}

export default function DataVizPlayground() {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const [repo, setRepo] = useState('react')
  const [selectedNode, setSelectedNode] = useState<number | null>(null)

  const repos: Record<string, { nodes: Node[]; links: Link[] }> = {
    react: {
      nodes: [
        {
          id: 0,
          x: 0,
          y: 0,
          vx: 0,
          vy: 0,
          file: 'hooks.ts',
          coupling: 92,
        },
        {
          id: 1,
          x: 0,
          y: 0,
          vx: 0,
          vy: 0,
          file: 'component.ts',
          coupling: 87,
        },
        {
          id: 2,
          x: 0,
          y: 0,
          vx: 0,
          vy: 0,
          file: 'render.ts',
          coupling: 81,
        },
        {
          id: 3,
          x: 0,
          y: 0,
          vx: 0,
          vy: 0,
          file: 'state.ts',
          coupling: 76,
        },
        {
          id: 4,
          x: 0,
          y: 0,
          vx: 0,
          vy: 0,
          file: 'utils.ts',
          coupling: 65,
        },
      ],
      links: [
        { source: 0, target: 1, strength: 0.92 },
        { source: 0, target: 2, strength: 0.81 },
        { source: 1, target: 2, strength: 0.87 },
        { source: 1, target: 3, strength: 0.76 },
        { source: 2, target: 4, strength: 0.65 },
      ],
    },
    vue: {
      nodes: [
        {
          id: 0,
          x: 0,
          y: 0,
          vx: 0,
          vy: 0,
          file: 'reactivity.ts',
          coupling: 88,
        },
        {
          id: 1,
          x: 0,
          y: 0,
          vx: 0,
          vy: 0,
          file: 'template.ts',
          coupling: 84,
        },
        {
          id: 2,
          x: 0,
          y: 0,
          vx: 0,
          vy: 0,
          file: 'watcher.ts',
          coupling: 79,
        },
        {
          id: 3,
          x: 0,
          y: 0,
          vx: 0,
          vy: 0,
          file: 'parser.ts',
          coupling: 72,
        },
      ],
      links: [
        { source: 0, target: 1, strength: 0.88 },
        { source: 0, target: 2, strength: 0.79 },
        { source: 1, target: 3, strength: 0.72 },
      ],
    },
  }

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    canvas.width = canvas.offsetWidth
    canvas.height = canvas.offsetHeight

    const data = repos[repo as keyof typeof repos]

    // Initialize positions
    data.nodes.forEach((node) => {
      if (node.x === 0 && node.y === 0) {
        node.x = Math.random() * canvas.width
        node.y = Math.random() * canvas.height
      }
    })

    const animate = () => {
      ctx.clearRect(0, 0, canvas.width, canvas.height)

      // Apply force-directed layout
      data.nodes.forEach((node, i) => {
        data.nodes.forEach((other, j) => {
          if (i === j) return

          const dx = other.x - node.x
          const dy = other.y - node.y
          const dist = Math.sqrt(dx * dx + dy * dy) || 1

          // Repulsion
          const repulsion = 50 / (dist * dist + 1)
          node.vx -= (dx / dist) * repulsion
          node.vy -= (dy / dist) * repulsion
        })

        // Apply attraction for linked nodes
        data.links.forEach((link) => {
          if (link.source === node.id) {
            const target = data.nodes[link.target]
            const dx = target.x - node.x
            const dy = target.y - node.y
            const dist = Math.sqrt(dx * dx + dy * dy) || 1
            const attraction =
              link.strength * 0.1 * (dist - 100)
            node.vx += (dx / dist) * attraction
            node.vy += (dy / dist) * attraction
          }
        })

        // Damping
        node.vx *= 0.95
        node.vy *= 0.95

        // Update position
        node.x += node.vx
        node.y += node.vy

        // Boundaries
        node.x = Math.max(20, Math.min(canvas.width - 20, node.x))
        node.y = Math.max(20, Math.min(canvas.height - 20, node.y))
      })

      // Draw links
      ctx.strokeStyle = 'rgba(140, 200, 100, 0.2)'
      ctx.lineWidth = 1
      data.links.forEach((link) => {
        const source = data.nodes[link.source]
        const target = data.nodes[link.target]

        ctx.beginPath()
        ctx.moveTo(source.x, source.y)
        ctx.lineTo(target.x, target.y)
        ctx.stroke()

        // Draw strength indicator
        ctx.fillStyle = `rgba(140, 200, 100, ${link.strength * 0.3})`
        ctx.beginPath()
        ctx.arc(
          (source.x + target.x) / 2,
          (source.y + target.y) / 2,
          2,
          0,
          Math.PI * 2
        )
        ctx.fill()
      })

      // Draw nodes
      data.nodes.forEach((node) => {
        const isSelected = selectedNode === node.id
        const size = isSelected ? 10 : 6

        ctx.fillStyle = isSelected
          ? 'rgba(140, 200, 100, 1)'
          : `rgba(140, 200, 100, ${0.5 + node.coupling / 200})`
        ctx.beginPath()
        ctx.arc(node.x, node.y, size, 0, Math.PI * 2)
        ctx.fill()

        // Label
        if (isSelected) {
          ctx.fillStyle = 'rgba(140, 200, 100, 0.8)'
          ctx.font = '12px monospace'
          ctx.textAlign = 'center'
          ctx.fillText(node.file, node.x, node.y - 15)
          ctx.fillText(`${node.coupling}%`, node.x, node.y + 20)
        }
      })

      requestAnimationFrame(animate)
    }

    animate()
  }, [repo, selectedNode])

  return (
    <section className="min-h-screen bg-background py-20 px-4 border-t border-border">
      <div className="max-w-6xl mx-auto">
        <h2 className="text-5xl font-bold text-foreground mb-4">
          Interactive Coupling Explorer
        </h2>
        <p className="text-lg text-muted-foreground mb-12">
          Visualize file dependencies in real-time. Force-directed graphs show
          how files cluster together based on change patterns.
        </p>

        <div className="grid md:grid-cols-3 gap-8 mb-8">
          {Object.keys(repos).map((repoName) => (
            <button
              key={repoName}
              onClick={() => setRepo(repoName)}
              className={`px-4 py-3 rounded-sm font-mono text-sm transition-all ${
                repo === repoName
                  ? 'bg-accent text-accent-foreground'
                  : 'bg-card border border-border text-foreground hover:bg-card/80'
              }`}
            >
              {repoName.charAt(0).toUpperCase() + repoName.slice(1)}
            </button>
          ))}
        </div>

        <div className="border border-border rounded-none bg-card/30 overflow-hidden">
          <canvas
            ref={canvasRef}
            className="w-full h-96 block"
            onClick={(e) => {
              const canvas = canvasRef.current
              if (!canvas) return

              const rect = canvas.getBoundingClientRect()
              const x = e.clientX - rect.left
              const y = e.clientY - rect.top

              const data =
                repos[repo as keyof typeof repos]
              let clicked = false

              data.nodes.forEach((node) => {
                const dist = Math.sqrt(
                  Math.pow(x - node.x, 2) +
                    Math.pow(y - node.y, 2)
                )
                if (dist < 10) {
                  setSelectedNode(
                    selectedNode === node.id ? null : node.id
                  )
                  clicked = true
                }
              })

              if (!clicked) {
                setSelectedNode(null)
              }
            }}
          />
        </div>

        <div className="mt-8 grid md:grid-cols-2 gap-8">
          <div className="border border-border p-6 bg-card/50">
            <h3 className="text-foreground font-semibold mb-4">
              How to Use
            </h3>
            <ul className="space-y-2 text-sm text-muted-foreground">
              <li>• Switch repositories to compare patterns</li>
              <li>• Click on nodes to see detailed coupling</li>
              <li>• Node size indicates coupling strength</li>
              <li>• Lines show file change relationships</li>
            </ul>
          </div>
          <div className="border border-border p-6 bg-card/50">
            <h3 className="text-foreground font-semibold mb-4">
              Key Insights
            </h3>
            <ul className="space-y-2 text-sm text-muted-foreground">
              <li>• Clustered nodes = high interdependency</li>
              <li>• Identify refactoring opportunities</li>
              <li>• Plan feature changes with impact analysis</li>
              <li>• Improve code organization systematically</li>
            </ul>
          </div>
        </div>
      </div>
    </section>
  )
}
