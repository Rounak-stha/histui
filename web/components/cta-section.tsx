'use client'

export default function CTASection() {
  return (
    <section className="relative bg-background border-t border-border py-20 px-4 overflow-hidden">
      {/* Animated background */}
      <div className="absolute inset-0 pointer-events-none">
        <div
          className="absolute top-0 left-1/4 w-96 h-96 bg-accent/10 rounded-full blur-3xl mix-blend-multiply animate-blob"
          style={{
            animation: 'blob 7s infinite',
          }}
        ></div>
        <div
          className="absolute top-1/2 right-1/4 w-96 h-96 bg-accent/5 rounded-full blur-3xl mix-blend-multiply animate-blob"
          style={{
            animation: 'blob 7s infinite 2s',
          }}
        ></div>
      </div>

      <div className="max-w-4xl mx-auto text-center relative z-10">
        <h2 className="text-5xl md:text-6xl font-bold text-foreground mb-6">
          Ready to refactor with confidence?
        </h2>
        <p className="text-lg md:text-xl text-muted-foreground mb-12 max-w-2xl mx-auto">
          Start analyzing your codebase today. Uncover hidden dependencies,
          identify refactoring opportunities, and make better architectural
          decisions.
        </p>

        <div className="flex flex-col sm:flex-row gap-4 justify-center items-center mb-12">
          <button className="px-8 py-4 bg-accent text-accent-foreground font-semibold rounded-sm hover:opacity-90 transition-opacity text-lg">
            Get Started
          </button>
          <button className="px-8 py-4 border border-accent text-accent font-semibold rounded-sm hover:bg-accent/10 transition-colors text-lg">
            Read Documentation
          </button>
        </div>

        <div className="grid md:grid-cols-3 gap-8 mt-16">
          <div className="p-6 border border-border bg-card/30 rounded-sm">
            <div className="text-3xl font-bold text-accent mb-2">100K+</div>
            <div className="text-sm text-muted-foreground">Files Analyzed</div>
          </div>
          <div className="p-6 border border-border bg-card/30 rounded-sm">
            <div className="text-3xl font-bold text-accent mb-2">50+</div>
            <div className="text-sm text-muted-foreground">
              Repositories Scanned
            </div>
          </div>
          <div className="p-6 border border-border bg-card/30 rounded-sm">
            <div className="text-3xl font-bold text-accent mb-2">1000s</div>
            <div className="text-sm text-muted-foreground">
              Happy Developers
            </div>
          </div>
        </div>
      </div>

      <style jsx>{`
        @keyframes blob {
          0%,
          100% {
            transform: translate(0, 0) scale(1);
          }
          33% {
            transform: translate(30px, -50px) scale(1.1);
          }
          66% {
            transform: translate(-20px, 20px) scale(0.9);
          }
        }
      `}</style>
    </section>
  )
}
