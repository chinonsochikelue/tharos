"use client";

import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";

export function Header() {
    const pathname = usePathname();
    const pageName = pathname === "/" ? "Dashboard" : pathname.split("/").pop();

    return (
        <header className="h-20 flex items-center justify-between px-8 py-4 relative z-40">
            {/* Breadcrumb / Title */}
            <div>
                <div className="flex items-center gap-2 text-xs font-medium text-slate-500 uppercase tracking-widest mb-1">
                    <span>Menu</span>
                    <span className="text-slate-700">/</span>
                    <span className="text-orange-500">{pageName}</span>
                </div>
            </div>

            {/* Actions */}
            <div className="flex items-center gap-6">
                <div className="relative group">
                    <input
                        type="text"
                        placeholder="Search..."
                        className="w-64 pl-10 pr-4 py-2.5 bg-slate-900/50 border border-slate-800 rounded-full text-sm text-slate-200 focus:outline-none focus:ring-2 focus:ring-orange-500/50 focus:border-transparent transition-all placeholder:text-slate-600 group-hover:bg-slate-900/80"
                    />
                    <svg
                        className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500 group-hover:text-slate-400 transition-colors"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                    >
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                    </svg>
                </div>

                <div className="h-6 w-px bg-slate-800"></div>

                <div className="flex items-center gap-4">
                    <button className="relative p-2.5 rounded-full hover:bg-slate-800/50 text-slate-400 hover:text-white transition-colors group">
                        <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                        </svg>
                        <span className="absolute top-2 right-2.5 w-2 h-2 bg-orange-500 rounded-full border-2 border-slate-950 group-hover:scale-110 transition-transform"></span>
                    </button>

                    <button className="p-1 rounded-full border border-slate-800 hover:border-slate-600 transition-colors">
                        <img
                            src="https://api.dicebear.com/7.x/avataaars/svg?seed=Felix"
                            alt="User"
                            className="w-8 h-8 rounded-full bg-slate-800"
                        />
                    </button>
                </div>
            </div>
        </header>
    );
}
