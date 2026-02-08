export default function UseCasesSection() {
  const useCases = [
    {
      title: 'Refactoring Legacy Code',
      description:
        'Understand which files in your legacy system are tightly coupled, helping you prioritize refactoring efforts and reduce technical debt.',
      industry: 'Enterprise',
    },
    {
      title: 'Monorepo Architecture',
      description:
        'Detect coupling patterns across packages in your monorepo. Identify modules that should be separated or merged based on actual change history.',
      industry: 'Scaling Teams',
    },
    {
      title: 'Onboarding New Developers',
      description:
        'Help new team members understand the system architecture by showing which files work together, reducing the time to productive contributions.',
      industry: 'Team Collaboration',
    },
    {
      title: 'Code Review Standards',
      description:
        'Use coupling insights to establish code review guidelines. When high-coupling areas change, require additional scrutiny and testing.',
      industry: 'Quality Assurance',
    },
    {
      title: 'Technical Debt Tracking',
      description:
        'Track coupling scores over time to measure whether your refactoring efforts are actually improving architecture health.',
      industry: 'DevOps',
    },
    {
      title: 'Microservices Migration',
      description:
        'Identify which features can be cleanly extracted into microservices by finding naturally decoupled areas of your codebase.',
      industry: 'Architecture',
    },
  ]

  return (
    <section className="w-full bg-background py-20 px-4 border-b border-border">
      <div className="max-w-6xl mx-auto">
        <div className="mb-16">
          <h2 className="text-5xl font-bold text-foreground mb-4">Use Cases</h2>
          <p className="text-xl text-muted-foreground max-w-2xl">
            How teams use histui to improve code architecture and developer
            productivity.
          </p>
        </div>

        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          {useCases.map((useCase, idx) => (
            <div
              key={idx}
              className="group border border-border p-8 bg-card/50 hover:bg-card/80 hover:border-accent/50 transition-all duration-300"
            >
              <div className="mb-4">
                <span className="text-xs font-mono text-accent px-3 py-1 bg-accent/10 rounded">
                  {useCase.industry}
                </span>
              </div>
              <h3 className="text-lg font-semibold text-foreground mb-3">
                {useCase.title}
              </h3>
              <p className="text-sm text-muted-foreground leading-relaxed">
                {useCase.description}
              </p>
            </div>
          ))}
        </div>

        <div className="mt-16 bg-card/50 border border-border p-12 text-center">
          <h3 className="text-2xl font-semibold text-foreground mb-3">
            Your use case here?
          </h3>
          <p className="text-muted-foreground mb-6 max-w-xl mx-auto">
            Every codebase is unique. We're building histui to support diverse
            workflows and industries. Share your use case with us.
          </p>
          <a
            href="https://github.com/Rounak-stha/histui/issues"
            className="inline-block px-6 py-3 bg-accent text-accent-foreground font-semibold hover:opacity-90 transition-opacity"
          >
            Share Your Story
          </a>
        </div>
      </div>
    </section>
  )
}
