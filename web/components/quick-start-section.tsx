'use client'

import { useState } from 'react'

export default function QuickStartSection() {
  const [activeTab, setActiveTab] = useState<'npm' | 'yarn' | 'pnpm' | 'cargo'>(
    'npm'
  )

  const commands = {
    npm: 'npm install -g histui',
    yarn: 'yarn global add histui',
    pnpm: 'pnpm install -g histui',
    cargo: 'cargo install histui',
  }

  const steps = [
    {
      number: 1,
      title: 'Install histui',
      description: 'Add histui to your system using your package manager',
      code: commands[activeTab],
    },
    {
      number: 2,
      title: 'Navigate to your repository',
      description: 'Run histui in any Git repository',
      code: 'cd your-project',
    },
    {
      number: 3,
      title: 'Analyze file coupling',
      description: 'Generate your first coupling analysis',
      code: 'histui --coupling',
    },
    {
      number: 4,
      title: 'Explore the results',
      description: 'Review coupling scores and refactor opportunities',
      code: 'histui --coupling --json > report.json',
    },
  ]

  return (
    <section className="w-full bg-background py-20 px-4 border-b border-border">
      <div className="max-w-5xl mx-auto">
        <div className="mb-16">
          <h2 className="text-5xl font-bold text-foreground mb-4">
            Quick Start
          </h2>
          <p className="text-xl text-muted-foreground max-w-2xl">
            Get histui running in your repository in minutes.
          </p>
        </div>

        {/* Installation Tabs */}
        <div className="mb-12">
          <div className="flex border-b border-border mb-6">
            {(['npm', 'yarn', 'pnpm', 'cargo'] as const).map((pm) => (
              <button
                key={pm}
                onClick={() => setActiveTab(pm)}
                className={`px-4 py-3 font-mono text-sm transition-colors ${
                  activeTab === pm
                    ? 'text-accent border-b-2 border-accent'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                {pm}
              </button>
            ))}
          </div>

          <div className="bg-card border border-border p-6">
            <code className="text-accent font-mono">
              $ {commands[activeTab]}
            </code>
          </div>
        </div>

        {/* Steps */}
        <div className="space-y-8">
          {steps.map((step, idx) => (
            <div key={idx} className="flex gap-6">
              <div className="flex-shrink-0">
                <div className="flex items-center justify-center h-12 w-12 rounded-full bg-accent text-accent-foreground font-bold">
                  {step.number}
                </div>
              </div>
              <div className="flex-1">
                <h3 className="text-lg font-semibold text-foreground mb-2">
                  {step.title}
                </h3>
                <p className="text-muted-foreground mb-4">{step.description}</p>
                <div className="bg-card/50 border border-border/50 p-4 rounded-sm">
                  <code className="text-accent font-mono text-sm">
                    $ {step.code}
                  </code>
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* Additional Resources */}
        <div className="mt-16 grid md:grid-cols-3 gap-8">
          <a
            href="#"
            className="group p-6 border border-border bg-card/50 hover:bg-card/80 hover:border-accent/50 transition-all"
          >
            <h4 className="font-semibold text-foreground mb-2 group-hover:text-accent transition-colors">
              Documentation
            </h4>
            <p className="text-sm text-muted-foreground">
              Explore the complete API reference and configuration options
            </p>
          </a>
          <a
            href="#"
            className="group p-6 border border-border bg-card/50 hover:bg-card/80 hover:border-accent/50 transition-all"
          >
            <h4 className="font-semibold text-foreground mb-2 group-hover:text-accent transition-colors">
              Examples
            </h4>
            <p className="text-sm text-muted-foreground">
              See how to use histui with real-world codebases and frameworks
            </p>
          </a>
          <a
            href="#"
            className="group p-6 border border-border bg-card/50 hover:bg-card/80 hover:border-accent/50 transition-all"
          >
            <h4 className="font-semibold text-foreground mb-2 group-hover:text-accent transition-colors">
              GitHub
            </h4>
            <p className="text-sm text-muted-foreground">
              Contribute to histui, report issues, or suggest improvements
            </p>
          </a>
        </div>
      </div>
    </section>
  )
}
