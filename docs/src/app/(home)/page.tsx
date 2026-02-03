'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import Image from 'next/image';

const TERMINAL_DATA = {
  init: {
    title: "tharos init",
    lines: [
      { text: "Initializing Tharos AI hooks...", delay: 0 },
      { text: "✓ Found .git directory", delay: 400 },
      { text: "✓ Created .git/hooks/pre-commit", delay: 800 },
      { text: "✓ Successfully linked Go engine", delay: 1200 },
      { text: "Tharos is now shielding your repository.", delay: 1600, color: "text-cyan-neon" }
    ]
  },
  check: {
    title: "tharos check",
    lines: [
      { text: "Scanning staged files...", delay: 0 },
      { text: "• auth-service.ts", delay: 300 },
      { text: "• database.sql", delay: 500 },
      { text: "⚠️  Potential SQLi found on line 12", delay: 800, color: "text-yellow-400" },
      { text: "Commit blocked by organizational policy.", delay: 1200, color: "text-red-500" }
    ]
  },
  analyze: {
    title: "tharos analyze --ai",
    lines: [
      { text: "Running deep AI analysis...", delay: 0 },
      { text: "Sending context to Gemini 2.0...", delay: 1200 },
      { text: "🤖 AI Advice found (Risk: 84/100)", delay: 2400, color: "text-purple-neon" },
      { text: "💡 Recommendation: Use sha256 instead of md5", delay: 4800, color: "text-green-400" },
      { text: "Use --fix to apply surgical replacement.", delay: 9600, color: "text-cyan-neon" }
    ]
  },
  fix: {
    title: "tharos --fix",
    lines: [
      { text: "Applying surgical fixes...", delay: 0 },
      { text: "Targeting line 42 in utils.go", delay: 400 },
      { text: "✓ Byte-offset precision: Matched", delay: 800 },
      { text: "✅ Applied 1 fix in 12ms", delay: 1200, color: "text-green-400" },
      { text: "Codebase is now compliant.", delay: 1600, color: "text-cyan-neon" }
    ]
  }
};

function TerminalLine({ text, color, delay }: { text: string; color?: string; delay: number }) {
  const [visible, setVisible] = useState(false);
  useEffect(() => {
    const t = setTimeout(() => setVisible(true), delay);
    return () => clearTimeout(t);
  }, [delay]);

  if (!visible) return null;
  return (
    <div className={`font-mono text-sm mb-1 opacity-0 translate-y-1 animate-[fadeUp_0.3s_ease-out_forwards] ${color || 'text-white/80'}`}>
      <span className="text-cyan-neon mr-2">$</span>
      {text}
    </div>
  );
}

function InteractiveTerminal() {
  const [activeTab, setActiveTab] = useState<keyof typeof TERMINAL_DATA>('analyze');
  const [key, setKey] = useState(0);

  const handleTabChange = (tab: keyof typeof TERMINAL_DATA) => {
    setActiveTab(tab);
    setKey(prev => prev + 1);
  };

  return (
    <div className="w-full max-w-4xl mx-auto mt-12 mb-12">
      <div className="flex gap-2 mb-0 p-2 pl-4 rounded-t-xl glass border-b-0 border-white/10 overflow-x-auto no-scrollbar">
        {Object.keys(TERMINAL_DATA).map((tab) => (
          <button
            key={tab}
            onClick={() => handleTabChange(tab as any)}
            className={`px-4 py-1.5 rounded-lg text-xs font-bold transition-all whitespace-nowrap ${activeTab === tab
              ? 'bg-cyan-neon/20 text-cyan-neon border border-cyan-neon/30 shadow-[0_0_15px_rgba(0,242,255,0.2)]'
              : 'text-white/40 hover:text-white/60'
              }`}
          >
            {TERMINAL_DATA[tab as keyof typeof TERMINAL_DATA].title}
          </button>
        ))}
      </div>
      <div className="p-6 rounded-b-xl glass border-t-0 border-white/10 min-h-[200px] neon-border">
        <div key={key}>
          {TERMINAL_DATA[activeTab].lines.map((line, i) => (
            <TerminalLine key={i} {...line} />
          ))}
        </div>
      </div>
    </div>
  );
}

export default function HomePage() {
  return (
    <main className="flex flex-col items-center text-white bg-blue-deep">
      {/* Hero Section */}
      <section className="relative w-full min-h-[95vh] flex flex-col items-center justify-center p-8 overflow-hidden">
        {/* Hero Background Asset */}
        <div className="absolute inset-0 z-0 scale-105">
          <Image
            src="/hero.png"
            alt="Futuristic Shield"
            fill
            className="object-cover opacity-40 animate-pulse-slow grayscale-[0.3]"
            priority
          />
          <div className="absolute inset-0 bg-linear-to-b from-blue-deep/10 via-blue-deep/50 to-blue-deep" />
        </div>

        <div className="relative z-10 text-center max-w-[1000px]">
          <div className="inline-block px-6 py-1.5 rounded-full glass border border-cyan-neon/40 text-cyan-neon text-xs font-bold tracking-[0.2em] mb-8 opacity-0 translate-y-2 animate-[fadeUp_0.7s_ease-out_forwards_0.2s]">
            SYSTEM ENGINE RELOADED • V0.1.5
          </div>
          <h1 className="text-7xl md:text-9xl font-black mb-6 tracking-tighter leading-[0.9] opacity-0 translate-y-4 animate-[fadeUp_0.7s_ease-out_forwards_0.4s]">
            <span className="bg-linear-to-br from-white via-cyan-neon to-purple-neon bg-clip-text text-transparent neon-text-cyan">
              THAROS
            </span>
          </h1>
          <p className="text-white/60 text-lg md:text-2xl max-w-[800px] mb-12 leading-relaxed mx-auto font-light opacity-0 translate-y-4 animate-[fadeUp_0.7s_ease-out_forwards_0.6s]">
            The architect of <span className="text-white font-medium italic">Unbreakable Code Governance</span>.
            Automated intelligence that bridges the gap between raw intent and perfect policy.
          </p>

          <div className="flex flex-col sm:flex-row gap-6 justify-center items-center opacity-0 translate-y-4 animate-[fadeUp_0.7s_ease-out_forwards_0.8s]">
            <Link
              href="/docs"
              className="group relative px-10 py-5 rounded-2xl bg-cyan-neon text-blue-deep font-black overflow-hidden transition-all hover:scale-105 active:scale-95 shadow-[0_0_50px_rgba(0,242,255,0.3)]"
            >
              <span className="relative z-10">INITIALIZE SYSTEM</span>
              <div className="absolute inset-0 bg-white opacity-0 group-hover:opacity-20 transition-opacity" />
            </Link>
            <Link
              href="https://github.com"
              className="px-10 py-5 rounded-2xl glass text-white font-bold hover:bg-white/10 transition-all border border-white/20 hover:border-white/40"
            >
              SOURCE CODE
            </Link>
          </div>
        </div>
      </section>

      {/* Interactive Feature Demo */}
      <section className="py-32 w-full p-8 max-w-6xl relative">
        <div className="absolute top-0 left-1/2 -translate-x-1/2 w-px h-24 bg-linear-to-b from-cyan-neon to-transparent opacity-50" />

        <div className="text-center mb-16 px-4">
          <h2 className="text-4xl md:text-6xl font-black mb-6 tracking-tight italic">Watch it in Action</h2>
          <p className="text-white/40 max-w-2xl mx-auto text-lg">
            Experience the surgical precision of the Tharos Go Engine.
            Click the commands below to simulate system feedback.
          </p>
        </div>

        <InteractiveTerminal />
      </section>

      {/* Scanner Visual Section */}
      <section className="py-24 w-full p-8 max-w-7xl">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
          <div className="relative group p-4 sm:p-0">
            <div className="absolute -inset-4 bg-linear-to-r from-cyan-neon/20 to-purple-neon/20 blur-xl opacity-0 group-hover:opacity-100 transition-opacity duration-1000" />
            <div className="relative rounded-3xl glass p-3 neon-border overflow-hidden">
              <Image
                src="/dashboard.png"
                alt="Scanner Interface"
                width={800}
                height={600}
                className="rounded-2xl opacity-90 group-hover:scale-105 transition-transform duration-700"
              />
              <div className="absolute top-0 left-0 w-full h-1 bg-cyan-neon shadow-[0_0_15px_#00f2ff] animate-[scan_3s_linear_infinite]" />
            </div>
          </div>

          <div className="space-y-8 px-4 sm:px-0">
            <div className="p-2 inline-block rounded-lg bg-purple-neon/10 border border-purple-neon/20 text-purple-neon text-xs font-bold px-3">
              ADVANCED SCANNER
            </div>
            <h2 className="text-4xl md:text-6xl font-black leading-[1.1]">Surgical Byte-Level <span className="text-cyan-neon underline decoration-cyan-neon/30 underline-offset-8">Precision</span></h2>
            <p className="text-white/50 text-xl leading-relaxed font-light">
              Tharos doesn't just guess. It tracks every token down to its exact byte offset,
              allowing it to apply patches that fit perfectly into your existing code architecture.
            </p>
            <ul className="space-y-4">
              {[
                "AST-Aware Token Tracking",
                "Zero-Regret Semantic Patching",
                "AI-Enhanced Validation",
                "Ultra-Fast Go Execution"
              ].map((item, i) => (
                <li key={i} className="flex items-center gap-4 text-white/80">
                  <div className="w-6 h-6 rounded-full bg-cyan-neon/20 border border-cyan-neon/40 flex items-center justify-center text-cyan-neon">✓</div>
                  <span className="font-medium">{item}</span>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </section>

      {/* Features Grid */}
      <section className="grid grid-cols-1 md:grid-cols-3 gap-8 p-8 md:p-24 w-full max-w-7xl">
        {[
          {
            title: "Policy as Code",
            desc: "Define your engineering standards in tharos.yaml and enforce them consistently across the entire organization.",
            icon: "📜"
          },
          {
            title: "AI Guardrails",
            desc: "Hybrid LLM analysis prevents complex vulnerabilities that static analyzers miss, like logic flaws or weak crypto.",
            icon: "🛡️"
          },
          {
            title: "Sub-ms Performance",
            desc: "Built in Go to ensure that your pre-commit hooks and local scans are practically instantaneous.",
            icon: "🚀"
          }
        ].map((f, i) => (
          <div key={i} className="p-10 rounded-3xl glass neon-border group hover:bg-white/5 transition-all cursor-default lg:hover:-translate-y-2">
            <div className="text-5xl mb-8 group-hover:scale-110 transition-transform inline-block group-hover:rotate-3">{f.icon}</div>
            <h3 className="text-2xl font-black mb-4 text-white group-hover:text-cyan-neon transition-colors tracking-tight">{f.title}</h3>
            <p className="text-white/50 leading-relaxed font-light">
              {f.desc}
            </p>
          </div>
        ))}
      </section>

      {/* CTA Section */}
      <section className="py-24 w-full flex flex-col items-center px-4 sm:px-8 relative overflow-hidden">
        {/* Animated Background for CTA */}
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(0,242,255,0.05)_0%,transparent_70%)] pointer-events-none" />
        <div className="absolute inset-0 opacity-10 bg-[linear-gradient(to_right,#ffffff1a_1px,transparent_1px),linear-gradient(to_bottom,#ffffff1a_1px,transparent_1px)] bg-[size:40px_40px] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_50%,#000_70%,transparent_100%)]" />

        <div className="glass p-12 md:p-24 rounded-[40px] text-center max-w-5xl w-full neon-border relative overflow-hidden group">
          <div className="absolute top-0 right-0 w-64 h-64 bg-purple-neon/10 blur-[100px] pointer-events-none group-hover:bg-purple-neon/20 transition-colors duration-700" />
          <div className="absolute bottom-0 left-0 w-64 h-64 bg-cyan-neon/10 blur-[100px] pointer-events-none group-hover:bg-cyan-neon/20 transition-colors duration-700" />

          <h2 className="text-4xl md:text-8xl font-black mb-8 italic tracking-tighter leading-[0.8] animate-in fade-in slide-in-from-bottom-4 duration-700">
            SECURE YOUR<br />FUTURE NOW.
          </h2>
          <p className="text-white/50 mb-12 text-xl max-w-2xl mx-auto font-light leading-relaxed">
            Join 2,000+ engineering teams automating security governance with Tharos AI.
            Bridge the gap between raw intent and perfect policy.
          </p>

          <div className="flex flex-col sm:flex-row gap-6 justify-center items-center">
            <Link
              href="/docs"
              className="inline-block px-12 py-5 rounded-2xl bg-white text-blue-deep font-black hover:bg-cyan-neon hover:scale-105 transition-all shadow-[0_0_40px_rgba(255,255,255,0.2)] active:scale-95"
            >
              DEPLOY THAROS →
            </Link>
            <Link
              href="/docs/getting-started"
              className="inline-block px-12 py-5 rounded-2xl glass text-white font-bold hover:bg-white/10 hover:scale-105 transition-all border border-white/20 active:scale-95"
            >
              READ THE CASE STUDY
            </Link>
          </div>
        </div>
      </section>

      {/* Premium Footer */}
      <footer className="w-full pt-24 pb-12 px-8 border-t border-white/5 bg-blue-deep/50 relative overflow-hidden">
        <div className="max-w-7xl mx-auto">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-12 mb-20">
            {/* Brand Column */}
            <div className="space-y-6">
              <div className="flex items-center gap-3">
                <Image src="/logo.png" alt="Tharos Logo" width={60} height={60} />
                <span className="text-2xl font-black tracking-tighter italic">THAROS</span>
              </div>
              <p className="text-white/40 text-sm leading-relaxed max-w-xs">
                The next generation of Intelligent Code Governance.
                Built in Go for performance, powered by AI for precision.
              </p>
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
                <span className="text-[10px] font-bold tracking-widest text-green-500 uppercase">Go Engine Online • v0.1.5-stable</span>
              </div>
            </div>

            {/* Links Columns */}
            {[
              {
                title: "PRODUCT",
                links: ["Features", "Security", "AI Analyzer", "Pricing", "Enterprise"]
              },
              {
                title: "RESOURCES",
                links: ["Documentation", "API Reference", "Benchmarks", "Case Studies", "Status"]
              },
              {
                title: "COMPANY",
                links: ["About Us", "Blog", "Careers", "Security Portal", "Contact"]
              }
            ].map((col, i) => (
              <div key={i} className="space-y-6 text-center md:text-left">
                <h4 className="text-xs font-black tracking-[0.2em] text-cyan-neon uppercase">{col.title}</h4>
                <ul className="space-y-3">
                  {col.links.map((link) => (
                    <li key={link}>
                      <Link href="#" className="text-white/40 hover:text-white transition-colors text-sm font-medium">{link}</Link>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>

          <div className="pt-12 border-t border-white/5 flex flex-col md:flex-row justify-between items-center gap-8">
            <div className="text-[10px] font-bold tracking-[0.3em] text-white/20 uppercase">
              © 2026 MODULAR POLICY ENFORCEMENT SYSTEMS INC. ALL RIGHTS RESERVED.
            </div>

            <div className="flex gap-8">
              {["X", "GitHub", "Discord", "LinkedIn"].map((social) => (
                <Link key={social} href="#" className="text-white/20 hover:text-white transition-colors text-xs font-bold uppercase tracking-widest">
                  {social}
                </Link>
              ))}
            </div>
          </div>
        </div>
      </footer>

      <style jsx global>{`
        @keyframes scan {
          0% { transform: translateY(0); opacity: 0; }
          10%, 90% { opacity: 1; }
          100% { transform: translateY(450px); opacity: 0; }
        }
        @keyframes fadeUp {
          from { opacity: 0; transform: translateY(10px); }
          to { opacity: 1; transform: translateY(0); }
        }
        .no-scrollbar::-webkit-scrollbar {
          display: none;
        }
        .no-scrollbar {
          -ms-overflow-style: none;
          scrollbar-width: none;
        }
      `}</style>
    </main>
  );
}
