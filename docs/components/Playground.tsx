"use client";

import { useState, useRef } from "react";
import Editor from "@monaco-editor/react";
import {
    Terminal, Shield, Sparkles, AlertTriangle,
    CheckCircle, Link as LinkIcon, RefreshCcw,
    Code2, Share2, Download
} from "lucide-react";
import Link from "next/link";

const DEFAULT_CODE = `// Tharos Security Playground
// Paste your code here to see Tharos in action.

import React from 'react';

function Dashboard() {
  const [data, setData] = React.useState('');
  const [apiKey] = React.useState('sk-live-51Mz...7j2'); // Hardcoded secret

  const handleUpdate = (input: string) => {
    // Dangerous DOM sink
    document.getElementById('display')!.innerHTML = input; 
    
    // Unsafe eval
    eval(\`console.log("Updating with: \${input}")\`);
  };

  const syncData = () => {
    // Weak crypto
    const hash = crypto.createHash('md5').update(data).digest('hex');
    
    // Potential SQLi
    const query = "SELECT * FROM logs WHERE type = '" + data + "'";
    console.log(query);
  };

  return (
    <div>
      <h1 dangerouslySetInnerHTML={{ __html: 'Dashboard' }} />
      {/* TODO: Implement secure data fetching */}
    </div>
  );
}
`;

export function Playground() {
    const [code, setCode] = useState(DEFAULT_CODE);
    const [isScanning, setIsScanning] = useState(false);
    const [isFixing, setIsFixing] = useState(false);
    const [report, setReport] = useState<null | any>(null);
    const [useAI, setUseAI] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleScan = async (fix = false) => {
        if (fix) setIsFixing(true);
        else setIsScanning(true);

        setError(null);

        try {
            const response = await fetch('/api/analyze', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ code, ai: useAI, fix }),
            });

            if (!response.ok) {
                const errData = await response.json();
                throw new Error(errData.details || 'Failed to analyze code');
            }

            const data = await response.json();
            setReport(data);

            if (fix && data.fixedCode) {
                setCode(data.fixedCode);
            }
        } catch (err: any) {
            console.error("Analysis failed:", err);
            setError(err.message);
        } finally {
            setIsScanning(false);
            setIsFixing(false);
        }
    };

    const totalIssues = report?.summary?.vulnerabilities || 0;
    const allFindings = report?.results?.flatMap((r: any) => r.findings) || [];
    const aiInsights = report?.results?.flatMap((r: any) => r.ai_insights) || [];

    return (
        <div className="flex flex-col h-screen overflow-hidden bg-fd-background text-fd-foreground font-outfit">
            {/* Header */}
            <header className="flex items-center justify-between px-6 py-4 border-b border-fd-border bg-fd-muted/30 backdrop-blur-md">
                <div className="flex items-center gap-4">
                    <Link href="/" className="flex items-center gap-2 group">
                        <div className="bg-orange-600 p-1.5 rounded-lg group-hover:bg-orange-500 transition-colors">
                            <Shield className="w-5 h-5 text-white" />
                        </div>
                        <span className="font-black italic tracking-tighter text-xl">THAROS PLAYGROUND</span>
                    </Link>
                    <div className="h-4 w-px bg-fd-border mx-2" />
                    <nav className="hidden md:flex items-center gap-6">
                        <Link href="/docs" className="text-sm font-medium text-fd-muted-foreground hover:text-fd-foreground transition-colors">Documentation</Link>
                        <Link href="/docs/architecture" className="text-sm font-medium text-fd-muted-foreground hover:text-fd-foreground transition-colors">Architecture</Link>
                    </nav>
                </div>

                <div className="flex items-center gap-3">
                    <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-fd-muted/50 border border-fd-border text-xs font-medium">
                        <div className={`w-2 h-2 rounded-full ${isScanning || isFixing ? 'bg-orange-500 animate-pulse' : 'bg-green-500'}`} />
                        Engine: Go v0.1.0
                    </div>
                </div>
            </header>

            {/* Main Editor Suite */}
            <main className="flex-1 flex overflow-hidden">
                {/* Editor Section */}
                <div className="flex-[3] flex flex-col border-r border-fd-border relative">
                    <div className="flex items-center justify-between px-4 py-2 bg-fd-muted/20 border-b border-fd-border">
                        <div className="flex items-center gap-2 text-xs font-black text-fd-muted-foreground uppercase tracking-widest">
                            <Code2 className="w-3.5 h-3.5" />
                            editor.tsx
                        </div>
                        <div className="flex items-center gap-2">
                            <label className="flex items-center gap-2 text-[10px] font-black tracking-tighter cursor-pointer hover:bg-fd-muted px-2 py-1 rounded transition-colors uppercase">
                                <input
                                    type="checkbox"
                                    checked={useAI}
                                    onChange={(e) => setUseAI(e.target.checked)}
                                    className="hidden"
                                />
                                <div className={`w-3 h-3 rounded flex items-center justify-center border ${useAI ? 'bg-purple-600 border-purple-500' : 'bg-fd-background border-fd-border'}`}>
                                    {useAI && <Sparkles className="w-2.5 h-2.5 text-white" />}
                                </div>
                                AI Insights
                            </label>
                            <button
                                onClick={() => setCode(DEFAULT_CODE)}
                                className="p-1.5 rounded hover:bg-fd-muted text-fd-muted-foreground transition-colors"
                                title="Reset Code"
                            >
                                <RefreshCcw className="w-3.5 h-3.5" />
                            </button>
                        </div>
                    </div>

                    <div className="flex-1 min-h-0 bg-[#1e1e1e]">
                        <Editor
                            height="100%"
                            defaultLanguage="typescript"
                            theme="vs-dark"
                            value={code}
                            onChange={(value) => setCode(value || "")}
                            options={{
                                fontSize: 14,
                                fontVariant: 'ligatures',
                                fontStyle: 'italic',
                                minimap: { enabled: false },
                                padding: { top: 20 },
                                smoothScrolling: true,
                                cursorSmoothCaretAnimation: 'on' as any,
                                roundedSelection: false,
                                scrollBeyondLastLine: false,
                            }}
                        />
                    </div>

                    <div className="absolute bottom-6 right-6 flex gap-3">
                        {allFindings.some((f: any) => f.replacement) && (
                            <button
                                onClick={() => handleScan(true)}
                                disabled={isScanning || isFixing}
                                className="bg-green-600 text-white px-6 py-3 rounded-xl font-bold hover:bg-green-700 transition-all flex items-center gap-2 disabled:opacity-50 shadow-2xl hover:scale-105 active:scale-95 group"
                            >
                                {isFixing ? (
                                    <div className="animate-spin h-5 w-5 border-2 border-white border-t-transparent rounded-full" />
                                ) : (
                                    <Sparkles className="w-5 h-5 group-hover:rotate-12 transition-transform" />
                                )}
                                MAGIC FIX
                            </button>
                        )}
                        <button
                            onClick={() => handleScan(false)}
                            disabled={isScanning || isFixing}
                            className="bg-orange-600 text-white px-8 py-4 rounded-xl font-black text-sm tracking-widest hover:bg-orange-500 transition-all flex items-center gap-3 disabled:opacity-50 shadow-2xl hover:scale-105 active:scale-95 uppercase italic"
                        >
                            {isScanning ? (
                                <div className="animate-spin h-5 w-5 border-2 border-white border-t-transparent rounded-full" />
                            ) : (
                                <Shield className="w-5 h-5" />
                            )}
                            {isScanning ? "Scanning Core..." : "ANALYZE CODE →"}
                        </button>
                    </div>
                </div>

                {/* Report Section */}
                <div className="flex-1 flex flex-col bg-fd-card min-w-[320px] max-w-[450px]">
                    <div className="px-6 py-4 border-b border-fd-border flex items-center justify-between">
                        <div className="flex items-center gap-2">
                            <Terminal className="w-4 h-4 text-orange-500" />
                            <span className="text-sm font-bold uppercase tracking-widest">Analysis Engine</span>
                        </div>
                    </div>

                    <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
                        {error && (
                            <div className="p-4 rounded-xl border border-red-500/20 bg-red-500/5 text-red-500 text-sm flex gap-3 items-start animate-in fade-in slide-in-from-top-2">
                                <AlertTriangle className="w-5 h-5 mt-0.5 flex-shrink-0" />
                                <div className="font-bold">{error}</div>
                            </div>
                        )}

                        {!report && !error && !isScanning && (
                            <div className="h-full flex flex-col items-center justify-center text-center p-8">
                                <div className="bg-fd-muted p-4 rounded-3xl mb-6">
                                    <Shield className="w-12 h-12 text-fd-muted-foreground/30" />
                                </div>
                                <h3 className="text-lg font-bold mb-2">Awaiting Input</h3>
                                <p className="text-sm text-fd-muted-foreground leading-relaxed">
                                    Paste your code and click analyze to start the Tharos security assessment.
                                </p>
                            </div>
                        )}

                        {isScanning && (
                            <div className="h-full flex flex-col items-center justify-center space-y-4 py-20">
                                <div className="w-12 h-12 border-4 border-orange-600/20 border-t-orange-600 rounded-full animate-spin" />
                                <div className="text-sm font-black text-orange-600 uppercase tracking-tighter">Traversing AST...</div>
                            </div>
                        )}

                        {report && (
                            <div className="space-y-6 animate-in fade-in duration-500">
                                {/* Result Summary */}
                                <div className={`p-6 rounded-2xl border transition-all ${totalIssues > 0 ? 'bg-red-500/5 border-red-500/20' : 'bg-green-500/5 border-green-500/20'
                                    }`}>
                                    <div className="flex items-center gap-4 mb-4">
                                        <div className={`p-2 rounded-lg ${totalIssues > 0 ? 'bg-red-500/20' : 'bg-green-500/20'}`}>
                                            {totalIssues > 0 ? <AlertTriangle className="w-6 h-6 text-red-500" /> : <CheckCircle className="w-6 h-6 text-green-500" />}
                                        </div>
                                        <div>
                                            <div className="text-sm font-black uppercase tracking-widest">
                                                {totalIssues > 0 ? "Commit Blocked" : "Verified Clean"}
                                            </div>
                                            <div className="text-xs opacity-50 font-medium">Scanned in {report.summary.duration}</div>
                                        </div>
                                    </div>

                                    <div className="grid grid-cols-2 gap-4">
                                        <div className="p-3 bg-fd-background rounded-xl border border-fd-border/50">
                                            <div className="text-[10px] font-bold text-fd-muted-foreground uppercase opacity-50 mb-1">Vulnerabilities</div>
                                            <div className={`text-xl font-black ${totalIssues > 0 ? 'text-red-500' : 'text-green-500'}`}>{totalIssues}</div>
                                        </div>
                                        <div className="p-3 bg-fd-background rounded-xl border border-fd-border/50">
                                            <div className="text-[10px] font-bold text-fd-muted-foreground uppercase opacity-50 mb-1">Score</div>
                                            <div className="text-xl font-black text-fd-foreground">{100 - (totalIssues * 10)}</div>
                                        </div>
                                    </div>
                                </div>

                                {/* Findings List */}
                                {allFindings.length > 0 && (
                                    <div className="space-y-3">
                                        <h4 className="text-[10px] font-black text-fd-muted-foreground uppercase tracking-[0.2em] px-1">Detailed Findings</h4>
                                        <div className="space-y-3">
                                            {allFindings.map((finding: any, i: number) => (
                                                <div key={i} className="group relative p-4 rounded-xl border border-fd-border bg-fd-card/50 hover:bg-fd-card transition-colors">
                                                    <div className="flex items-center justify-between mb-2">
                                                        <span className={`text-[9px] font-black px-2 py-0.5 rounded-md uppercase tracking-tighter ${finding.severity === 'block' ? 'bg-red-500/10 text-red-500 border border-red-500/20' : 'bg-orange-500/10 text-orange-500 border border-orange-500/20'
                                                            }`}>
                                                            {finding.type.replace('_', ' ')}
                                                        </span>
                                                        <span className="text-[10px] font-bold text-fd-muted-foreground/50">L{finding.line}</span>
                                                    </div>
                                                    <div className="text-xs font-semibold leading-relaxed group-hover:text-fd-foreground transition-colors">{finding.message}</div>
                                                </div>
                                            ))}
                                        </div>
                                    </div>
                                )}

                                {/* AI Section */}
                                {aiInsights.length > 0 && (
                                    <div className="space-y-3 pt-4 border-t border-fd-border">
                                        <div className="flex items-center gap-2 px-1">
                                            <Sparkles className="w-3.5 h-3.5 text-purple-500" />
                                            <h4 className="text-[10px] font-black text-purple-500 uppercase tracking-[0.2em]">GenAI Recommendations</h4>
                                        </div>
                                        {aiInsights.map((insight: any, i: number) => (
                                            <div key={i} className="p-4 rounded-2xl border border-purple-500/20 bg-purple-500/5 space-y-3">
                                                <div className="text-xs font-bold leading-relaxed text-purple-200/90 italic">
                                                    &quot;{insight.recommendation}&quot;
                                                </div>
                                                {insight.risk_score > 0 && (
                                                    <div className="pt-3 flex flex-col gap-2">
                                                        <div className="flex justify-between text-[9px] font-black uppercase text-purple-400">
                                                            <span>Risk Probability</span>
                                                            <span>{insight.risk_score}%</span>
                                                        </div>
                                                        <div className="w-full h-1 bg-purple-500/10 rounded-full overflow-hidden">
                                                            <div className="h-full bg-purple-500" style={{ width: `${insight.risk_score}%` }} />
                                                        </div>
                                                    </div>
                                                )}
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </div>
                        )}
                    </div>

                    {/* Footer Tools */}
                    <div className="p-4 border-t border-fd-border flex items-center justify-around bg-fd-muted/10">
                        <button className="flex flex-col items-center gap-1 group">
                            <Share2 className="w-4 h-4 text-fd-muted-foreground group-hover:text-fd-foreground" />
                            <span className="text-[9px] font-bold text-fd-muted-foreground uppercase opacity-50">Share</span>
                        </button>
                        <button className="flex flex-col items-center gap-1 group">
                            <Download className="w-4 h-4 text-fd-muted-foreground group-hover:text-fd-foreground" />
                            <span className="text-[9px] font-bold text-fd-muted-foreground uppercase opacity-50">Export</span>
                        </button>
                    </div>
                </div>
            </main>

            <style jsx global>{`
                .custom-scrollbar::-webkit-scrollbar {
                    width: 4px;
                }
                .custom-scrollbar::-webkit-scrollbar-track {
                    background: transparent;
                }
                .custom-scrollbar::-webkit-scrollbar-thumb {
                    background: var(--fd-border);
                    border-radius: 10px;
                }
            `}</style>
        </div>
    );
}
