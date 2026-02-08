"use client";

import { useState, useRef, useEffect } from "react";
import Editor from "@monaco-editor/react";
import {
    Terminal, Shield, Sparkles, AlertTriangle,
    CheckCircle, Link as LinkIcon, RefreshCcw,
    Code2, Share2, Download, Copy, Check
} from "lucide-react";
import Link from "next/link";
import { useSearchParams, useRouter } from "next/navigation";
import Image from "next/image";

const DEFAULT_CODE = `const express = require("express");
const mysql = require("mysql");
const app = express();

app.use(express.json());

// ❌ Insecure CORS: Allowing any origin
app.use((req, res, next) => {
  res.header("Access-Control-Allow-Origin", "*");
  res.header("X-Content-Type-Options", "none"); // ❌ Insecure Header
  next();
});

// ❌ Hard-coded DB credentials
const db = mysql.createConnection({
  host: "localhost",
  user: "root",
  password: "root123",
  database: "test_db"
});

db.connect();

// ❌ Insecure login endpoint (SQL Injection)
app.post("/login", (req, res) => {
  const { email, password } = req.body;

  // ❌ Direct variable interpolation (SQL Injection)
  const query = \`SELECT * FROM users WHERE email = '\${email}'\`;

  db.query(query, (err, results) => {
    if (err) {
      return res.status(500).json({ error: err.message });
    }

    if (results.length > 0) {
      // ❌ DOM-like vulnerability (re-routing data)
      const message = "<h1>Welcome, " + results[0].name + "</h1>";
      // This line is client-side JS, but for demo purposes, imagine it's rendered server-side
      // document.body.innerHTML = message;
      res.json({
        success: true,
        message: "Logged in successfully",
        user: results[0] // ❌ leaking full user object
      });
    } else {
      res.status(401).json({ success: false, message: "Invalid credentials" });
    }
  });
});

// ❌ No auth, anyone can access this
app.get("/admin", (req, res) => {
  res.send("Welcome Admin 😈");
});

// ❌ Debug info exposed
app.get("/debug", (req, res) => {
  res.json({
    env: process.env,
    secretKey: "SUPER_SECRET_KEY_123"
  });
});

app.listen(3000, () => {
  console.log("🚨 Insecure server running on port 3000");
});
`;

export function Playground() {
    const [code, setCode] = useState(DEFAULT_CODE);
    const [isScanning, setIsScanning] = useState(false);
    const [isFixing, setIsFixing] = useState(false);
    const [report, setReport] = useState<null | any>(null);
    const [useAI, setUseAI] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [copied, setCopied] = useState(false);

    const searchParams = useSearchParams();
    const router = useRouter();

    // Load code from URL if present
    useEffect(() => {
        const encodedCode = searchParams.get('code');
        if (encodedCode) {
            try {
                const decoded = atob(encodedCode);
                setCode(decoded);
            } catch (e) {
                console.error("Failed to decode code from URL", e);
            }
        }
    }, [searchParams]);

    const handleShare = () => {
        const encoded = btoa(code);
        const url = `${window.location.origin}${window.location.pathname}?code=${encoded}`;
        navigator.clipboard.writeText(url);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    const handleExport = () => {
        const blob = new Blob([code], { type: 'text/typescript' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `tharos-analysis-${Date.now()}.ts`;
        a.click();
        URL.revokeObjectURL(url);
    };

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
            <header className="flex items-center justify-between px-4 md:px-6 py-4 border-b border-fd-border bg-fd-muted/30 backdrop-blur-md z-50">
                <div className="flex items-center gap-3 md:gap-4">
                    <Link href="/" className="flex items-center gap-2 group">
                        <Image src="/logo.png" alt="Tharos Logo" width={32} height={32} />
                        <span className="font-black italic tracking-tighter text-lg md:text-xl">THAROS <span className="hidden sm:inline pl-2">PLAYGROUND</span></span>
                    </Link>
                    <div className="h-4 w-px bg-fd-border mx-1 md:mx-2 hidden sm:block" />
                    <nav className="hidden lg:flex items-center gap-6">
                        <Link href="/docs" className="text-sm font-medium text-fd-muted-foreground hover:text-fd-foreground transition-colors">Documentation</Link>
                        <Link href="/docs/architecture" className="text-sm font-medium text-fd-muted-foreground hover:text-fd-foreground transition-colors">Architecture</Link>
                    </nav>
                </div>

                <div className="flex items-center gap-3">
                    <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-fd-muted/50 border border-fd-border text-[10px] md:text-xs font-medium">
                        <div className={`w-2 h-2 rounded-full ${isScanning || isFixing ? 'bg-orange-500 animate-pulse' : 'bg-green-500'}`} />
                        <span className="hidden xs:inline">Engine:</span> Go v1.2.0
                    </div>
                </div>
            </header>

            {/* Main Editor Suite */}
            <main className="flex-1 flex flex-col md:flex-row overflow-y-auto md:overflow-hidden">
                {/* Editor Section */}
                <div className="w-full md:flex-[3] flex flex-col border-b md:border-b-0 md:border-r border-fd-border relative min-h-[400px] md:min-h-0">
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
                                minimap: { enabled: false },
                                padding: { top: 20 },
                                smoothScrolling: true,
                                cursorSmoothCaretAnimation: 'on' as any,
                                roundedSelection: false,
                                scrollBeyondLastLine: false,
                            }}
                        />
                    </div>

                    <div className="absolute bottom-4 right-4 md:bottom-6 md:right-6 flex flex-wrap gap-2 md:gap-3 justify-end">
                        {allFindings.some((f: any) => f.replacement) && (
                            <button
                                onClick={() => handleScan(true)}
                                disabled={isScanning || isFixing}
                                className="bg-green-600 text-white px-4 md:px-6 py-2.5 md:py-3 rounded-xl font-bold hover:bg-green-700 transition-all flex items-center gap-2 disabled:opacity-50 shadow-2xl hover:scale-105 active:scale-95 group text-xs md:text-base"
                            >
                                {isFixing ? (
                                    <div className="animate-spin h-4 w-4 md:h-5 md:w-5 border-2 border-white border-t-transparent rounded-full" />
                                ) : (
                                    <Sparkles className="w-4 h-4 md:w-5 md:h-5 group-hover:rotate-12 transition-transform" />
                                )}
                                MAGIC <span className="hidden xs:inline">FIX</span>
                            </button>
                        )}
                        <button
                            onClick={() => handleScan(false)}
                            disabled={isScanning || isFixing}
                            className="bg-orange-600 text-white px-5 md:px-8 py-3 md:py-4 rounded-xl font-black text-xs md:text-sm tracking-widest hover:bg-orange-500 transition-all flex items-center gap-2 md:gap-3 disabled:opacity-50 shadow-2xl hover:scale-105 active:scale-95 uppercase italic"
                        >
                            {isScanning ? (
                                <div className="animate-spin h-4 w-4 md:h-5 md:w-5 border-2 border-white border-t-transparent rounded-full" />
                            ) : (
                                <Shield className="w-4 h-4 md:w-5 md:h-5" />
                            )}
                            {isScanning ? "Scanning..." : "ANALYZE →"}
                        </button>
                    </div>
                </div>

                {/* Report Section */}
                <div className="w-full md:flex-1 flex flex-col bg-fd-card md:min-w-[320px] md:max-w-[450px]">
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
                                                {allFindings.some((f: any) => f.severity === 'critical') ? "Critical Block" :
                                                    allFindings.some((f: any) => ['high', 'medium'].includes(f.severity)) ? "Issues Detected" : "Verified Clean"}
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
                                                        <div className="flex items-center gap-2">
                                                            <span className={`text-[9px] font-black px-2 py-0.5 rounded-md uppercase tracking-tighter ${['critical', 'high'].includes(finding.severity) ? 'bg-red-500/10 text-red-500 border border-red-500/20' :
                                                                finding.severity === 'medium' ? 'bg-orange-500/10 text-orange-500 border border-orange-500/20' :
                                                                    'bg-blue-500/10 text-blue-500 border border-blue-500/20'
                                                                }`}>
                                                                {finding.severity}
                                                            </span>
                                                            <span className="text-[10px] font-bold text-fd-muted-foreground/30">{finding.rule}</span>
                                                        </div>
                                                        <span className="text-[10px] font-bold text-fd-muted-foreground/50">L{finding.line}</span>
                                                    </div>
                                                    <div className="text-xs font-bold leading-relaxed group-hover:text-fd-foreground transition-colors mb-2">{finding.message}</div>

                                                    {finding.explain && (
                                                        <div className="text-[11px] text-fd-muted-foreground leading-relaxed mb-2 opacity-80 group-hover:opacity-100 transition-opacity">
                                                            {finding.explain}
                                                        </div>
                                                    )}

                                                    {finding.remediation && (
                                                        <div className="mt-3 pt-3 border-t border-fd-border/50">
                                                            <div className="text-[10px] font-black uppercase tracking-widest text-green-500/70 mb-1">Remediation</div>
                                                            <div className="text-[11px] font-medium text-fd-foreground/70">{finding.remediation}</div>
                                                        </div>
                                                    )}

                                                    {finding.confidence > 0 && (
                                                        <div className="mt-2 flex items-center gap-2">
                                                            <div className="flex-1 h-1 bg-fd-muted rounded-full overflow-hidden">
                                                                <div
                                                                    className={`h-full ${finding.confidence > 0.8 ? 'bg-green-500' : 'bg-yellow-500'}`}
                                                                    style={{ width: `${finding.confidence * 100}%` }}
                                                                />
                                                            </div>
                                                            <span className="text-[8px] font-black text-fd-muted-foreground">{(finding.confidence * 100).toFixed(0)}% CONFIDENCE</span>
                                                        </div>
                                                    )}
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
                        <button
                            onClick={handleShare}
                            className="flex flex-col items-center gap-1 group transition-all"
                        >
                            {copied ? (
                                <Check className="w-4 h-4 text-green-500" />
                            ) : (
                                <Share2 className="w-4 h-4 text-fd-muted-foreground group-hover:text-fd-foreground" />
                            )}
                            <span className={`text-[9px] font-bold uppercase opacity-50 ${copied ? 'text-green-500' : 'text-fd-muted-foreground'}`}>
                                {copied ? 'Copied' : 'Share'}
                            </span>
                        </button>
                        <button
                            onClick={handleExport}
                            className="flex flex-col items-center gap-1 group transition-all"
                        >
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