"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
    LayoutDashboard,
    Shield,
    BarChart3,
    Users,
    FileCheck,
    Settings,
    Github,
    LogOut
} from "lucide-react";
import { cn } from "@/lib/utils";
import { motion } from "framer-motion";

const navigation = [
    { name: "Dashboard", href: "/", icon: LayoutDashboard },
    { name: "Policies", href: "/policies", icon: Shield },
    { name: "Analytics", href: "/analytics", icon: BarChart3 },
    { name: "Team", href: "/team", icon: Users },
    { name: "Compliance", href: "/compliance", icon: FileCheck },
    { name: "Settings", href: "/settings", icon: Settings },
];

export function Sidebar() {
    const pathname = usePathname();

    return (
        <div className="w-64 glass border-r border-slate-800/50 flex flex-col relative z-50">
            {/* Logo */}
            <div className="p-8 pb-4">
                <div className="flex items-center gap-4">
                    <div className="w-10 h-10 bg-gradient-to-br from-orange-500 to-orange-600 rounded-xl flex items-center justify-center shadow-lg shadow-orange-500/20">
                        <span className="text-2xl pt-1">🦊</span>
                    </div>
                    <div>
                        <h1 className="text-xl font-bold gradient-text tracking-tight">Tharos</h1>
                        <p className="text-[10px] uppercase tracking-widest text-slate-500 font-semibold mt-0.5">Enterprise</p>
                    </div>
                </div>
            </div>

            {/* Navigation */}
            <nav className="flex-1 px-4 py-4 space-y-1.5">
                {navigation.map((item) => {
                    const isActive = pathname === item.href;
                    return (
                        <Link
                            key={item.name}
                            href={item.href}
                            className="block relative group"
                        >
                            {isActive && (
                                <motion.div
                                    layoutId="sidebar-active"
                                    className="absolute inset-0 bg-orange-500/10 rounded-xl border border-orange-500/20"
                                    initial={false}
                                    transition={{ type: "spring", stiffness: 300, damping: 30 }}
                                />
                            )}
                            <div
                                className={cn(
                                    "relative flex items-center gap-3 px-4 py-3 rounded-xl transition-colors duration-200",
                                    isActive
                                        ? "text-orange-500"
                                        : "text-slate-400 group-hover:text-slate-100 group-hover:bg-slate-800/30"
                                )}
                            >
                                <item.icon className={cn("w-5 h-5", isActive && "drop-shadow-[0_0_8px_rgba(249,115,22,0.5)]")} />
                                <span className="font-medium text-sm">{item.name}</span>
                            </div>
                        </Link>
                    );
                })}
            </nav>

            {/* Footer */}
            <div className="p-4 border-t border-slate-800/30 bg-black/10">
                <div className="flex items-center gap-3 px-4 py-3 rounded-xl bg-gradient-to-br from-slate-800/50 to-slate-900/50 border border-slate-700/30">
                    <div className="relative">
                        <div className="w-8 h-8 rounded-full bg-gradient-to-r from-blue-500 to-purple-500 flex items-center justify-center text-xs font-bold text-white">
                            JD
                        </div>
                        <div className="absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 bg-green-500 border-2 border-slate-900 rounded-full"></div>
                    </div>
                    <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-slate-200 truncate">John Doe</p>
                        <p className="text-xs text-slate-500 truncate">Admin Workspace</p>
                    </div>
                    <button className="text-slate-500 hover:text-white transition-colors">
                        <LogOut className="w-4 h-4" />
                    </button>
                </div>
            </div>
        </div>
    );
}
