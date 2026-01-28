"use client";

import { useState } from "react";
import { Terminal, Shield, Sparkles, AlertTriangle } from "lucide-react";

export function AIDemo() {
    const [code, setCode] = useState(`// Try some "bad" code\nconst apiKey = "sk-123456789";\nfunction run() {\n  eval("console.log(apiKey)");\n}`);
    const [isScanning, setIsScanning] = useState(false);
    const [report, setReport] = useState<null | any>(null);

    const handleScan = () => {
        setIsScanning(true);
        setReport(null);
        setTimeout(() => {
            setIsScanning(false);
            setReport({
                score: 42,
                status: "Blocked",
                issues: [
                    { type: "Security", message: "Hardcoded secret detected", severity: "High" },
                    { type: "Architecture", message: "Use of 'eval' is strictly forbidden by policy #ARCH-02", severity: "Critical" }
                ],
                aiFeedback: "Tharos detected a hardcoded API key and the use of 'eval'. This violates both security best practices and organizational governance."
            });
        }, 1500);
    };

    return (
        <div className="my-8 border border-fd-border rounded-xl bg-fd-card overflow-hidden shadow-lg not-prose">
            <div className="bg-fd-muted p-4 border-b border-fd-border flex items-center gap-2">
                <Terminal className="w-4 h-4 text-orange-500" />
                <span className="text-sm font-semibold tracking-tight">Tharos Interactive Sandbox</span>
            </div>

            <div className="p-6 space-y-4">
                <div className="relative">
                    <textarea
                        className="w-full h-40 bg-fd-background text-fd-foreground p-4 rounded-lg border border-fd-border font-mono text-sm focus:ring-2 focus:ring-orange-500 focus:outline-none resize-none"
                        value={code}
                        onChange={(e) => setCode(e.target.value)}
                    />
                    <button
                        onClick={handleScan}
                        disabled={isScanning}
                        className="absolute bottom-4 right-4 bg-orange-600 text-white px-4 py-2 rounded-md hover:bg-orange-700 transition-all flex items-center gap-2 disabled:opacity-50"
                    >
                        {isScanning ? (
                            <span className="flex items-center gap-2">
                                <div className="animate-spin h-4 w-4 border-2 border-white border-t-transparent rounded-full" />
                                Scanning...
                            </span>
                        ) : (
                            <>
                                <Shield className="w-4 h-4" />
                                Scan with Tharos
                            </>
                        )}
                    </button>
                </div>

                {report && (
                    <div className="mt-4 p-4 rounded-lg border border-red-500/30 bg-red-500/5 animate-in fade-in slide-in-from-top-2 duration-300">
                        <div className="flex items-center justify-between mb-4">
                            <div className="flex items-center gap-2 text-red-500 font-bold">
                                <AlertTriangle className="w-5 h-5" />
                                <span>Commit Blocked (Score: {report.score})</span>
                            </div>
                            <Sparkles className="w-5 h-5 text-orange-500 animate-pulse" />
                        </div>
                        <ul className="space-y-2 mb-4">
                            {report.issues.map((issue: any, i: number) => (
                                <li key={i} className="flex items-start gap-2 text-sm text-fd-muted-foreground">
                                    <span className="text-red-400 font-mono">[{issue.type}]</span>
                                    <span>{issue.message}</span>
                                </li>
                            ))}
                        </ul>
                        <div className="text-sm border-t border-fd-border pt-4 text-fd-muted-foreground italic">
                            &quot;{report.aiFeedback}&quot;
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}
