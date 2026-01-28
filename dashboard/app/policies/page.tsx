"use client";

import { MetricCard } from "@/components/metric-card";
import { Shield, Lock, FileText, Plus } from "lucide-react";
import Link from "next/link";

const policies = [
    {
        id: "owasp-top10",
        name: "OWASP Top 10",
        description: "Standard awareness document for developers and web application security.",
        rules: 54,
        enabled: true,
        severity: "block",
    },
    {
        id: "soc2",
        name: "SOC 2 Type II",
        description: "Security, Availability, Processing Integrity, Confidentiality, and Privacy.",
        rules: 42,
        enabled: true,
        severity: "block",
    },
    {
        id: "gdpr",
        name: "GDPR Compliance",
        description: "Data protection and privacy for individuals within the European Union.",
        rules: 38,
        enabled: false,
        severity: "warning",
    },
    {
        id: "pci-dss",
        name: "PCI-DSS v4.0",
        description: "Payment Card Industry Data Security Standard for handling credit cards.",
        rules: 48,
        enabled: false,
        severity: "block",
    },
];

export default function PoliciesPage() {
    return (
        <div className="space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold gradient-text">Policy Management</h1>
                    <p className="text-slate-400 mt-1">
                        Configure security rules and compliance frameworks
                    </p>
                </div>
                <div className="flex gap-3">
                    <Link
                        href="/policies/library"
                        className="flex items-center gap-2 px-4 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 transition-colors"
                    >
                        <FileText className="w-4 h-4" />
                        <span>Browse Library</span>
                    </Link>
                    <Link
                        href="/policies/custom"
                        className="flex items-center gap-2 px-4 py-2 rounded-lg bg-orange-500 hover:bg-orange-600 text-white transition-colors"
                    >
                        <Plus className="w-4 h-4" />
                        <span>Create Policy</span>
                    </Link>
                </div>
            </div>

            {/* Metrics */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <MetricCard
                    title="Active Policies"
                    value="2"
                    suffix="/ 4"
                    icon={<Shield className="w-6 h-6" />}
                    color="green"
                />
                <MetricCard
                    title="Total Rules Enforced"
                    value="96"
                    trend="+12"
                    trendUp={true}
                    icon={<Lock className="w-6 h-6" />}
                    color="blue"
                />
                <MetricCard
                    title="Policy Violations"
                    value="14"
                    trend="-3"
                    trendUp={true}
                    icon={<FileText className="w-6 h-6" />}
                    color="orange"
                />
            </div>

            {/* Policies Grid */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                {policies.map((policy) => (
                    <div
                        key={policy.id}
                        className="glass rounded-xl p-6 hover:border-slate-700/50 transition-all duration-300 group"
                    >
                        <div className="flex items-start justify-between mb-4">
                            <div className="flex items-center gap-3">
                                <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${policy.enabled ? 'bg-orange-500/10 text-orange-500' : 'bg-slate-800 text-slate-400'}`}>
                                    <Shield className="w-5 h-5" />
                                </div>
                                <div>
                                    <h3 className="font-semibold text-lg">{policy.name}</h3>
                                    <div className="flex items-center gap-2 mt-1">
                                        <span className={`w-2 h-2 rounded-full ${policy.enabled ? 'bg-green-500' : 'bg-slate-500'}`}></span>
                                        <span className="text-xs text-slate-400">
                                            {policy.enabled ? 'Active' : 'Disabled'}
                                        </span>
                                    </div>
                                </div>
                            </div>
                            <button className="text-slate-400 hover:text-white transition-colors">
                                Edit
                            </button>
                        </div>

                        <p className="text-slate-400 text-sm mb-6 h-10 line-clamp-2">
                            {policy.description}
                        </p>

                        <div className="flex items-center justify-between pt-4 border-t border-slate-800/50">
                            <div className="flex gap-4 text-sm text-slate-400">
                                <span>{policy.rules} Rules</span>
                                <span className="capitalize">Severity: {policy.severity}</span>
                            </div>
                            <label className="relative inline-flex items-center cursor-pointer">
                                <input type="checkbox" className="sr-only peer" checked={policy.enabled} readOnly />
                                <div className="w-11 h-6 bg-slate-700 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-orange-500/20 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-orange-500"></div>
                            </label>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}
