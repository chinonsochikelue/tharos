"use client";

import { MetricCard } from "@/components/metric-card";
import { BarChart3, TrendingUp, AlertTriangle, Shield } from "lucide-react";
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar, Cell } from "recharts";

const performanceData = [
    { date: "Jan 21", findings: 45, fixed: 30 },
    { date: "Jan 22", findings: 52, fixed: 45 },
    { date: "Jan 23", findings: 38, fixed: 35 },
    { date: "Jan 24", findings: 65, fixed: 50 },
    { date: "Jan 25", findings: 48, fixed: 55 },
    { date: "Jan 26", findings: 42, fixed: 40 },
    { date: "Jan 27", findings: 35, fixed: 38 },
];

const vulnerabilityTypes = [
    { name: "Injection", value: 35, color: "#ef4444" },
    { name: "Auth", value: 25, color: "#f59e0b" },
    { name: "Config", value: 20, color: "#3b82f6" },
    { name: "XSS", value: 15, color: "#10b981" },
    { name: "Other", value: 5, color: "#8b5cf6" },
];

export default function AnalyticsPage() {
    return (
        <div className="space-y-6">
            {/* Header */}
            <div>
                <h1 className="text-3xl font-bold gradient-text">Analytics & Insights</h1>
                <p className="text-slate-400 mt-1">
                    Deep dive into your security metrics and trends
                </p>
            </div>

            {/* Metrics */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
                <MetricCard
                    title="Mean Time to Resolve"
                    value="4.2h"
                    trend="-1.5h"
                    trendUp={true}
                    icon={<TrendingUp className="w-6 h-6" />}
                    color="green"
                />
                <MetricCard
                    title="Total Findings"
                    value="342"
                    trend="+12%"
                    trendUp={false}
                    icon={<AlertTriangle className="w-6 h-6" />}
                    color="orange"
                />
                <MetricCard
                    title="Policies Enforced"
                    value="4"
                    icon={<Shield className="w-6 h-6" />}
                    color="blue"
                />
                <MetricCard
                    title="Risk Reduction"
                    value="85%"
                    trend="+5%"
                    trendUp={true}
                    icon={<BarChart3 className="w-6 h-6" />}
                    color="blue"
                />
            </div>

            {/* Charts Grid */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                {/* Remediation Performance */}
                <div className="glass rounded-xl p-6">
                    <div className="mb-6">
                        <h3 className="text-lg font-semibold">Remediation Performance</h3>
                        <p className="text-sm text-slate-400 mt-1">Findings vs Fixed (Last 7 days)</p>
                    </div>
                    <ResponsiveContainer width="100%" height={300}>
                        <AreaChart data={performanceData}>
                            <defs>
                                <linearGradient id="colorFindings" x1="0" y1="0" x2="0" y2="1">
                                    <stop offset="5%" stopColor="#ef4444" stopOpacity={0.3} />
                                    <stop offset="95%" stopColor="#ef4444" stopOpacity={0} />
                                </linearGradient>
                                <linearGradient id="colorFixed" x1="0" y1="0" x2="0" y2="1">
                                    <stop offset="5%" stopColor="#10b981" stopOpacity={0.3} />
                                    <stop offset="95%" stopColor="#10b981" stopOpacity={0} />
                                </linearGradient>
                            </defs>
                            <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                            <XAxis dataKey="date" stroke="#64748b" style={{ fontSize: '12px' }} />
                            <YAxis stroke="#64748b" style={{ fontSize: '12px' }} />
                            <Tooltip
                                contentStyle={{
                                    backgroundColor: '#1e293b',
                                    border: '1px solid #334155',
                                    borderRadius: '8px',
                                    color: '#f1f5f9'
                                }}
                            />
                            <Area type="monotone" dataKey="findings" stroke="#ef4444" fillOpacity={1} fill="url(#colorFindings)" />
                            <Area type="monotone" dataKey="fixed" stroke="#10b981" fillOpacity={1} fill="url(#colorFixed)" />
                        </AreaChart>
                    </ResponsiveContainer>
                </div>

                {/* Top Vulnerability Types */}
                <div className="glass rounded-xl p-6">
                    <div className="mb-6">
                        <h3 className="text-lg font-semibold">Top Vulnerabilities</h3>
                        <p className="text-sm text-slate-400 mt-1">Distribution by category</p>
                    </div>
                    <ResponsiveContainer width="100%" height={300}>
                        <BarChart data={vulnerabilityTypes} layout="vertical">
                            <CartesianGrid strokeDasharray="3 3" stroke="#334155" horizontal={false} />
                            <XAxis type="number" stroke="#64748b" style={{ fontSize: '12px' }} />
                            <YAxis dataKey="name" type="category" stroke="#64748b" style={{ fontSize: '12px' }} width={80} />
                            <Tooltip
                                cursor={{ fill: 'transparent' }}
                                contentStyle={{
                                    backgroundColor: '#1e293b',
                                    border: '1px solid #334155',
                                    borderRadius: '8px',
                                    color: '#f1f5f9'
                                }}
                            />
                            <Bar dataKey="value" radius={[0, 4, 4, 0]}>
                                {vulnerabilityTypes.map((entry, index) => (
                                    <Cell key={`cell-${index}`} fill={entry.color} />
                                ))}
                            </Bar>
                        </BarChart>
                    </ResponsiveContainer>
                </div>
            </div>
        </div>
    );
}
