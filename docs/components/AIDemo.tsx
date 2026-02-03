"use client";

import { useState } from "react";
import { Terminal, Shield, Sparkles, AlertTriangle, CheckCircle, Info, Link as LinkIcon } from "lucide-react";
import Link from "next/link";

export function AIDemo() {
    const [code, setCode] = useState(`// Try some "bad" code
const apiKey = "sk-123456789";
function run() {
  eval("console.log(apiKey)");
}`);
    const [isScanning, setIsScanning] = useState(false);
    const [isFixing, setIsFixing] = useState(false);
    const [report, setReport] = useState<null | any>(null);
    const [useAI, setUseAI] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleScan = async (fix = false) => {
        if (fix) setIsFixing(true);
        else setIsScanning(true);

        setReport(null);
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
        <div className="my-8 border border-fd-border rounded-xl bg-fd-card overflow-hidden shadow-lg not-prose">
            <div className="bg-fd-muted p-4 border-b border-fd-border flex items-center justify-between font-outfit">
                <div className="flex items-center gap-2">
                    <Terminal className="w-4 h-4 text-orange-500" />
                    <span className="text-sm font-semibold tracking-tight">Tharos Interactive Sandbox</span>
                </div>
                <div className="flex items-center gap-4">
                    <Link
                        href="/playground"
                        className="text-[10px] font-bold text-orange-600 hover:text-orange-500 uppercase flex items-center gap-1 group"
                    >
                        Try Full Playground
                        <LinkIcon className="w-3 h-3 group-hover:translate-x-0.5 transition-transform" />
                    </Link>
                    <label className="flex items-center gap-2 text-xs font-medium cursor-pointer">
                        <input
                            type="checkbox"
                            checked={useAI}
                            onChange={(e) => setUseAI(e.target.checked)}
                            className="rounded border-fd-border text-orange-600 focus:ring-orange-500"
                        />
                        Enable AI Insights
                    </label>
                </div>
            </div>

            <div className="p-6 space-y-4 font-outfit">
                <div className="relative">
                    <textarea
                        className="w-full h-48 bg-fd-background text-fd-foreground p-4 rounded-lg border border-fd-border font-mono text-sm focus:ring-2 focus:ring-orange-500 focus:outline-none resize-none transition-all"
                        value={code}
                        onChange={(e) => setCode(e.target.value)}
                        placeholder="Paste code here..."
                    />
                    <div className="absolute bottom-4 right-4 flex gap-2">
                        {allFindings.some((f: any) => f.replacement) && (
                            <button
                                onClick={() => handleScan(true)}
                                disabled={isScanning || isFixing}
                                className="bg-green-600 text-white px-4 py-2 rounded-md hover:bg-green-700 transition-all flex items-center gap-2 disabled:opacity-50 shadow-md group"
                            >
                                {isFixing ? (
                                    <div className="animate-spin h-4 w-4 border-2 border-white border-t-transparent rounded-full" />
                                ) : (
                                    <Sparkles className="w-4 h-4 group-hover:rotate-12 transition-transform" />
                                )}
                                Magic Fix
                            </button>
                        )}
                        <button
                            onClick={() => handleScan(false)}
                            disabled={isScanning || isFixing}
                            className="bg-orange-600 text-white px-4 py-2 rounded-md hover:bg-orange-700 transition-all flex items-center gap-2 disabled:opacity-50 shadow-lg"
                        >
                            {isScanning ? (
                                <span className="flex items-center gap-2">
                                    <div className="animate-spin h-4 w-4 border-2 border-white border-t-transparent rounded-full" />
                                    Analyzing...
                                </span>
                            ) : (
                                <>
                                    <Shield className="w-4 h-4" />
                                    Analyze Code
                                </>
                            )}
                        </button>
                    </div>
                </div>

                {error && (
                    <div className="p-4 rounded-lg border border-red-500/30 bg-red-500/5 text-red-500 text-sm flex gap-2 items-center">
                        <AlertTriangle className="w-4 h-4" />
                        {error}
                    </div>
                )}

                {report && (
                    <div className="mt-4 animate-in fade-in slide-in-from-top-2 duration-300 space-y-4">
                        <div className="flex items-center justify-between p-4 rounded-lg border border-fd-border bg-fd-muted/30">
                            <div className="flex items-center gap-3">
                                {totalIssues > 0 ? (
                                    <AlertTriangle className="w-5 h-5 text-red-500" />
                                ) : (
                                    <CheckCircle className="w-5 h-5 text-green-500" />
                                )}
                                <div>
                                    <div className="font-bold text-sm">
                                        {totalIssues > 0 ? `Detected ${totalIssues} Security Issue${totalIssues > 1 ? 's' : ''}` : "Clean Code: No Issues Detected"}
                                    </div>
                                    <div className="text-xs text-fd-muted-foreground">
                                        Scan complete in {report.summary.duration}
                                    </div>
                                </div>
                            </div>
                            <div className="text-2xl font-black text-fd-muted-foreground opacity-20">
                                {totalIssues > 0 ? "BLOCKED" : "COMPLIANT"}
                            </div>
                        </div>

                        {allFindings.length > 0 && (
                            <div className="space-y-2">
                                <div className="text-xs font-bold text-fd-muted-foreground uppercase tracking-widest px-1">Detailed Findings</div>
                                <div className="grid gap-2">
                                    {allFindings.map((finding: any, i: number) => (
                                        <div key={i} className="flex flex-col p-3 rounded-lg border border-fd-border bg-fd-card/50 text-sm">
                                            <div className="flex items-center justify-between mb-1">
                                                <span className={`text-[10px] font-bold px-1.5 py-0.5 rounded uppercase ${finding.severity === 'block' ? 'bg-red-500/20 text-red-500' : 'bg-orange-500/20 text-orange-500'
                                                    }`}>
                                                    {finding.type.replace('_', ' ')}
                                                </span>
                                                <span className="text-xs text-fd-muted-foreground">Line {finding.line}</span>
                                            </div>
                                            <div className="text-fd-foreground">{finding.message}</div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        )}

                        {aiInsights.length > 0 && (
                            <div className="space-y-2">
                                <div className="text-xs font-bold text-purple-500 uppercase tracking-widest px-1 flex items-center gap-1">
                                    <Sparkles className="w-3 h-3" />
                                    AI Recommendation
                                </div>
                                {aiInsights.map((insight: any, i: number) => (
                                    <div key={i} className="p-4 rounded-lg border border-purple-500/30 bg-purple-500/5 text-sm">
                                        <div className="text-fd-foreground leading-relaxed italic">
                                            &quot;{insight.recommendation}&quot;
                                        </div>
                                        {insight.risk_score > 0 && (
                                            <div className="mt-3 flex items-center gap-2">
                                                <div className="text-[10px] font-bold text-purple-400">RISK SCORE</div>
                                                <div className="w-full h-1.5 bg-purple-500/10 rounded-full overflow-hidden">
                                                    <div className="h-full bg-purple-500" style={{ width: `${insight.risk_score}%` }} />
                                                </div>
                                                <div className="text-xs font-bold text-purple-400">{insight.risk_score}</div>
                                            </div>
                                        )}
                                    </div>
                                ))}
                            </div>
                        )}

                        {!allFindings.length && useAI && (
                            <div className="p-4 rounded-lg border border-blue-500/30 bg-blue-500/5 text-sm flex gap-2 items-start">
                                <Info className="w-4 h-4 text-blue-500 mt-0.5" />
                                <div className="text-fd-muted-foreground">
                                    AI analysis confirms this code snippet follows basic security standards.
                                </div>
                            </div>
                        )}
                    </div>
                )}
            </div>
        </div>
    );
}
