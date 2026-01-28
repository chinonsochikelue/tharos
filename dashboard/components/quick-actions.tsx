"use client";

import { Play, FileText, Settings, Download, Zap } from "lucide-react";
import { motion } from "framer-motion";

const actions = [
    {
        title: "Run Full Scan",
        description: "Analyze all repositories",
        icon: Play,
        color: "orange",
    },
    {
        title: "Generate Report",
        description: "Export compliance report",
        icon: FileText,
        color: "blue",
    },
    {
        title: "Configure Policies",
        description: "Manage security rules",
        icon: Settings,
        color: "purple",
    },
    {
        title: "Export Data",
        description: "Download findings as CSV",
        icon: Download,
        color: "green",
    },
];

export function QuickActions() {
    return (
        <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.5, delay: 0.25 }}
            className="glass rounded-2xl p-6 border-slate-800/50 flex flex-col h-full"
        >
            <div className="mb-6 flex items-center gap-3">
                <div className="p-2 rounded-lg bg-yellow-500/10 text-yellow-500">
                    <Zap className="w-5 h-5" />
                </div>
                <div>
                    <h3 className="text-xl font-bold text-white">Quick Actions</h3>
                    <p className="text-sm text-slate-400">Common tasks</p>
                </div>
            </div>

            <div className="grid grid-cols-2 gap-4 flex-1">
                {actions.map((action, idx) => (
                    <motion.button
                        key={action.title}
                        whileHover={{ scale: 1.02, backgroundColor: "rgba(30, 41, 59, 0.8)" }}
                        whileTap={{ scale: 0.98 }}
                        className="p-4 rounded-xl bg-slate-800/40 border border-slate-700/40 flex flex-col justify-between group transition-colors hover:border-slate-600/60"
                    >
                        <div className={`w-10 h-10 rounded-lg bg-${action.color}-500/10 border border-${action.color}-500/20 flex items-center justify-center mb-3 group-hover:scale-110 transition-transform duration-300 shadow-[0_0_15px_-5px_rgba(var(--${action.color}-500-rgb),0.3)]`}>
                            <action.icon className={`w-5 h-5 text-${action.color}-500`} />
                        </div>
                        <div className="text-left">
                            <h4 className="font-semibold text-sm text-slate-200 group-hover:text-white transition-colors">{action.title}</h4>
                            <p className="text-[10px] text-slate-500 mt-1 line-clamp-1 group-hover:text-slate-400 transition-colors">{action.description}</p>
                        </div>
                    </motion.button>
                ))}
            </div>
        </motion.div>
    );
}
