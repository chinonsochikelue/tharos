import Link from 'next/link';

export default function HomePage() {
  return (
    <main className="flex flex-col flex-1">
      <div className="flex flex-col items-center justify-center p-8 text-center bg-linear-to-b from-fd-background to-fd-muted flex-1">
        <h1 className="text-5xl md:text-7xl font-bold mb-6 bg-linear-to-r from-orange-400 to-orange-600 bg-clip-text text-transparent">
          Tharos
        </h1>
        <p className="text-fd-muted-foreground text-lg md:text-2xl max-w-[600px] mb-8">
          Intelligent, Unbreakable Code Policy Enforcement.
          Bridge the gap between developer intent and organizational governance.
        </p>
        <div className="flex flex-row gap-4">
          <Link
            href="/docs"
            className="px-6 py-3 rounded-lg bg-orange-600 text-white font-semibold hover:bg-orange-700 transition-colors"
          >
            Get Started
          </Link>
          <Link
            href="https://github.com"
            className="px-6 py-3 rounded-lg bg-fd-secondary text-fd-secondary-foreground font-semibold hover:bg-fd-accent transition-colors border border-fd-border"
          >
            GitHub
          </Link>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8 p-8 md:p-16 bg-fd-background">
        <div className="p-6 rounded-xl border border-fd-border bg-fd-muted/50">
          <h3 className="text-xl font-bold mb-2">AI-Aware Hooks</h3>
          <p className="text-fd-muted-foreground">
            Semantic analysis of AST and dependencies to detect security flaws and code smells before they reach the repo.
          </p>
        </div>
        <div className="p-6 rounded-xl border border-fd-border bg-fd-muted/50">
          <h3 className="text-xl font-bold mb-2">Cloud Sync</h3>
          <p className="text-fd-muted-foreground">
            Centrally managed policies that sync automatically to all local environments. Tamper-evident and self-healing.
          </p>
        </div>
        <div className="p-6 rounded-xl border border-fd-border bg-fd-muted/50">
          <h3 className="text-xl font-bold mb-2">Performance First</h3>
          <p className="text-fd-muted-foreground">
            Core engine built in Go for lightning-fast analysis, ensuring zero impact on developer velocity.
          </p>
        </div>
      </div>
    </main>
  );
}
