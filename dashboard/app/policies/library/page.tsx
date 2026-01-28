"use client";

import { Search, Shield, ArrowRight } from "lucide-react";
import Link from "next/link";

const templates = [
    {
        id: "owasp",
        name: "OWASP Top 10",
        description: "Essential security baseline for web applications. Covers the most critical security risks.",
        category: "Security Standard",
        rules: 54,
        popuarity: "High",
    },
    {
        id: "soc2",
        name: "SOC 2 Type II",
        description: "Comprehensive controls for security, availability, and confidentiality.",
        category: "Compliance",
        rules: 42,
        popuarity: "High",
    },
    {
        id: "gdpr",
        name: "GDPR",
        description: "Data protection rules for handling EU personal data.",
        category: "Privacy",
        rules: 38,
        popuarity: "Medium",
    },
    {
        id: "pci-dss",
        name: "PCI-DSS v4.0",
        description: "Mandatory security standards for payment card processing.",
        category: "Compliance",
        rules: 48,
        popuarity: "Medium",
    },
    {
        id: "hipaa",
        name: "HIPAA",
        description: "Healthcare data protection requirements for PHI.",
        category: "Compliance",
        rules: 35,
        popuarity: "Low",
    },
    {
        id: "code-quality",
        name: "Code Quality",
        description: "General best practices for maintainability and clean code.",
        category: "Quality",
        rules: 60,
        popuarity: "High",
    },
];

export default function LibraryPage() {
    return (
        <div className="space-y-6">
            {/* Header */}
            <div>
                <div className="flex items-center gap-2 text-sm text-slate-400 mb-2">
                    <Link href="/policies" className="hover:text-white transition-colors">Policies</Link>
                    <span>/</span>
                    <span>Library</span>
                </div>
                <h1 className="text-3xl font-bold gradient-text">Policy Library</h1>
                <p className="text-slate-400 mt-1">
                    Browse and activate pre-built security templates
                </p>
            </div>

            {/* Search */}
            <div className="relative max-w-xl">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
                <input
                    type="text"
                    placeholder="Search templates (e.g. SOC2, Privacy)..."
                    className="w-full pl-10 pr-4 py-3 bg-slate-800/50 border border-slate-700/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-orange-500/50 focus:border-orange-500/50 transition-all"
                />
            </div>

            {/* Templates Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {templates.map((template) => (
                    <div
                        key={template.id}
                        className="glass rounded-xl p-6 hover:border-slate-700/50 transition-all duration-300 group flex flex-col"
                    >
                        <div className="flex items-start justify-between mb-4">
                            <div className="w-10 h-10 rounded-lg bg-blue-500/10 text-blue-500 flex items-center justify-center">
                                <Shield className="w-5 h-5" />
                            </div>
                            <span className="px-2 py-1 rounded text-xs font-medium bg-slate-800 text-slate-400 border border-slate-700">
                                {template.category}
                            </span>
                        </div>

                        <h3 className="font-semibold text-lg mb-2">{template.name}</h3>
                        <p className="text-slate-400 text-sm mb-6 flex-1">
                            {template.description}
                        </p>

                        <div className="flex items-center justify-between pt-4 border-t border-slate-800/50">
                            <span className="text-sm text-slate-400">{template.rules} Rules</span>
                            <button className="flex items-center gap-2 text-sm font-medium text-orange-500 hover:text-orange-400 transition-colors group-hover:translate-x-1 duration-200">
                                <span>View Details</span>
                                <ArrowRight className="w-4 h-4" />
                            </button>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}
