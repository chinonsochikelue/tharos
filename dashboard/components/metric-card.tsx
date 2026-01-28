"use client";

import { cn } from "@/lib/utils";
import { motion } from "framer-motion";

interface MetricCardProps {
    title: string;
    value: string;
    suffix?: string;
    trend?: string;
    trendUp?: boolean;
    icon: React.ReactNode;
    color?: "orange" | "red" | "green" | "blue";
    index?: number;
}

const colorStyles = {
    orange: {
        bg: "bg-orange-500/10",
        text: "text-orange-500",
        border: "group-hover:border-orange-500/30",
        glow: "group-hover:shadow-[0_0_20px_-5px_rgba(249,115,22,0.3)]"
    },
    red: {
        bg: "bg-red-500/10",
        text: "text-red-500",
        border: "group-hover:border-red-500/30",
        glow: "group-hover:shadow-[0_0_20px_-5px_rgba(239,68,68,0.3)]"
    },
    green: {
        bg: "bg-green-500/10",
        text: "text-green-500",
        border: "group-hover:border-green-500/30",
        glow: "group-hover:shadow-[0_0_20px_-5px_rgba(34,197,94,0.3)]"
    },
    blue: {
        bg: "bg-blue-500/10",
        text: "text-blue-500",
        border: "group-hover:border-blue-500/30",
        glow: "group-hover:shadow-[0_0_20px_-5px_rgba(59,130,246,0.3)]"
    },
};

export function MetricCard({
    title,
    value,
    suffix,
    trend,
    trendUp,
    icon: Icon,
    color = "orange",
    index = 0,
}: MetricCardProps) {
    const styles = colorStyles[color];

    return (
        <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4, delay: index * 0.1 }}
            className={cn(
                "glass rounded-2xl p-6 transition-all duration-300 group relative overflow-hidden",
                styles.border,
                styles.glow
            )}
        >
            {/* Background Glow Effect */}
            <div className={cn(
                "absolute -top-10 -right-10 w-32 h-32 rounded-full blur-3xl opacity-0 group-hover:opacity-20 transition-opacity duration-500",
                styles.bg.replace('/10', '') // Use full color for blur
            )} />

            <div className="flex items-start justify-between relative z-10">
                <div className="flex-1">
                    <p className="text-sm text-slate-400 font-medium tracking-wide">{title}</p>
                    <div className="mt-3 flex items-baseline gap-2">
                        <h3 className="text-4xl font-bold tracking-tight text-white">{value}</h3>
                        {suffix && <span className="text-xl text-slate-500 font-medium">{suffix}</span>}
                    </div>
                    {trend && (
                        <div className="flex items-center gap-2 mt-3">
                            <span
                                className={cn(
                                    "text-xs font-semibold px-2 py-0.5 rounded-full flex items-center gap-1",
                                    trendUp
                                        ? "bg-green-500/10 text-green-400 border border-green-500/20"
                                        : "bg-red-500/10 text-red-400 border border-red-500/20"
                                )}
                            >
                                {trendUp ? "↑" : "↓"} {trend}
                            </span>
                            <span className="text-xs text-slate-500">vs last week</span>
                        </div>
                    )}
                </div>
                <div
                    className={cn(
                        "w-12 h-12 rounded-xl flex items-center justify-center transition-transform duration-300 group-hover:scale-110 group-hover:rotate-3",
                        styles.bg,
                        styles.text
                    )}
                >
                    {Icon}
                </div>
            </div>
        </motion.div>
    );
}
