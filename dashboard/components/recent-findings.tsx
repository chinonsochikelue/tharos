"use client";

import { AlertTriangle, Shield, Code, ChevronRight, Clock } from "lucide-react";
import { cn } from "@/lib/utils";
import { motion } from "framer-motion";

const findings = [
    {
        id: 1,
        type: "security",
        severity: "critical",
        message: "Hardcoded API key detected in auth.ts",
        file: "src/auth.ts",
        line: 42,
        time: "2m ago",
    },
    {
        id: 2,
        type: "security",
        severity: "high",
        message: "SQL injection vulnerability in user query",
        file: "src/database/users.ts",
        line: 156,
        time: "15m ago",
    },
    {
        id: 3,
        type: "quality",
        severity: "medium",
        message: "Unused variable 'tempData' found",
        file: "src/utils/helpers.ts",
        line: 89,
        time: "1h ago",
    },
    {
        id: 4,
        type: "security",
        severity: "high",
        message: "Insecure random number generator used",
        file: "src/crypto/random.ts",
        line: 23,
        time: "2h ago",
    },
    {
        id: 5,
        type: "quality",
        severity: "low",
        message: "TODO comment found in production code",
        file: "src/components/Button.tsx",
        line: 67,
        time: "3h ago",
    },
];

const severityColors = {
    critical: "bg-red-500/10 text-red-500 border-red-500/20 shadow-[0_0_10px_-3px_rgba(239,68,68,0.3)]",
    high: "bg-orange-500/10 text-orange-500 border-orange-500/20 shadow-[0_0_10px_-3px_rgba(249,115,22,0.3)]",
    medium: "bg-yellow-500/10 text-yellow-500 border-yellow-500/20 shadow-[0_0_10px_-3px_rgba(234,179,8,0.3)]",
    low: "bg-blue-500/10 text-blue-500 border-blue-500/20 shadow-[0_0_10px_-3px_rgba(59,130,246,0.3)]",
};

export function RecentFindings() {
    return (
        <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.3 }}
            className="glass rounded-2xl p-6 border-slate-800/50"
        >
            <div className="flex items-center justify-between mb-8">
                <div>
                    <h3 className="text-xl font-bold text-white">Recent Findings</h3>
                    <p className="text-sm text-slate-400 mt-1">Real-time detection feed</p>
                </div>
                <button className="flex items-center gap-2 text-sm text-orange-500 hover:text-orange-400 font-medium group px-4 py-2 rounded-lg hover:bg-orange-500/10 transition-all">
                    View All
                    <ChevronRight className="w-4 h-4 transition-transform group-hover:translate-x-1" />
                </button>
            </div>

            <div className="space-y-3">
                {findings.map((finding, idx) => (
                    <motion.div
                        key={finding.id}
                        initial={{ opacity: 0, x: -10 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ duration: 0.3, delay: 0.4 + (idx * 0.1) }}
                        className="p-4 rounded-xl bg-slate-800/30 hover:bg-slate-800/60 border border-slate-700/30 hover:border-slate-600/50 transition-all duration-300 cursor-pointer group relative overflow-hidden"
                    >
                        {/* Hover Highlight */}
                        <div className="absolute inset-0 bg-gradient-to-r from-white/0 via-white/5 to-white/0 translate-x-[-100%] group-hover:translate-x-[100%] transition-transform duration-700 ease-in-out" />

                        <div className="flex items-start gap-4 relative z-10">
                            <div className="mt-1">
                                {finding.type === "security" ? (
                                    <div className="p-2 rounded-lg bg-orange-500/10 text-orange-500 group-hover:bg-orange-500/20 transition-colors">
                                        <Shield className="w-4 h-4" />
                                    </div>
                                ) : (
                                    <div className="p-2 rounded-lg bg-blue-500/10 text-blue-500 group-hover:bg-blue-500/20 transition-colors">
                                        <Code className="w-4 h-4" />
                                    </div>
                                )}
                            </div>
                            <div className="flex-1 min-w-0">
                                <div className="flex items-start justify-between gap-4">
                                    <p className="font-medium text-sm text-slate-200 group-hover:text-white transition-colors line-clamp-1">
                                        {finding.message}
                                    </p>
                                    <span
                                        className={cn(
                                            "px-2.5 py-0.5 rounded-full text-[10px] uppercase font-bold tracking-wide border shrink-0",
                                            severityColors[finding.severity as keyof typeof severityColors]
                                        )}
                                    >
                                        {finding.severity}
                                    </span>
                                </div>
                                <div className="flex items-center gap-3 mt-2 text-xs text-slate-500 group-hover:text-slate-400 transition-colors">
                                    <span className="font-mono bg-slate-950/30 px-1.5 py-0.5 rounded border border-slate-800/50">
                                        {finding.file}:{finding.line}
                                    </span>
                                    <div className="flex items-center gap-1">
                                        <Clock className="w-3 h-3" />
                                        <span>{finding.time}</span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </motion.div>
                ))}
            </div>
        </motion.div>
    );
}
