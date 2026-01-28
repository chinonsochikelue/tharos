"use client";

import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Area, AreaChart } from "recharts";
import { motion } from "framer-motion";

const data = [
    { date: "Jan 21", critical: 8, high: 15, medium: 23 },
    { date: "Jan 22", critical: 6, high: 12, medium: 20 },
    { date: "Jan 23", critical: 7, high: 14, medium: 22 },
    { date: "Jan 24", critical: 5, high: 11, medium: 18 },
    { date: "Jan 25", critical: 4, high: 9, medium: 16 },
    { date: "Jan 26", critical: 3, high: 8, medium: 14 },
    { date: "Jan 27", critical: 3, high: 7, medium: 12 },
];

export function TrendChart() {
    return (
        <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.5, delay: 0.2 }}
            className="glass rounded-2xl p-6 border-slate-800/50"
        >
            <div className="mb-8 flex items-center justify-between">
                <div>
                    <h3 className="text-xl font-bold text-white">Vulnerability Trends</h3>
                    <p className="text-sm text-slate-400 mt-1">Found vs Resolved (7 Days)</p>
                </div>
                <select className="bg-slate-800/50 border border-slate-700 rounded-lg text-xs px-3 py-1.5 focus:outline-none focus:ring-1 focus:ring-orange-500 text-slate-300">
                    <option>Last 7 Days</option>
                    <option>Last 30 Days</option>
                    <option>All Time</option>
                </select>
            </div>

            <div className="h-[300px] w-full">
                <ResponsiveContainer width="100%" height="100%">
                    <AreaChart data={data}>
                        <defs>
                            <linearGradient id="colorCritical" x1="0" y1="0" x2="0" y2="1">
                                <stop offset="5%" stopColor="#ef4444" stopOpacity={0.3} />
                                <stop offset="95%" stopColor="#ef4444" stopOpacity={0} />
                            </linearGradient>
                            <linearGradient id="colorHigh" x1="0" y1="0" x2="0" y2="1">
                                <stop offset="5%" stopColor="#f97316" stopOpacity={0.3} />
                                <stop offset="95%" stopColor="#f97316" stopOpacity={0} />
                            </linearGradient>
                        </defs>
                        <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" vertical={false} />
                        <XAxis
                            dataKey="date"
                            stroke="#64748b"
                            tick={{ fill: '#64748b', fontSize: 12 }}
                            tickLine={false}
                            axisLine={false}
                            dy={10}
                        />
                        <YAxis
                            stroke="#64748b"
                            tick={{ fill: '#64748b', fontSize: 12 }}
                            tickLine={false}
                            axisLine={false}
                            dx={-10}
                        />
                        <Tooltip
                            contentStyle={{
                                backgroundColor: 'rgba(15, 23, 42, 0.9)',
                                border: '1px solid rgba(255, 255, 255, 0.1)',
                                backdropFilter: 'blur(8px)',
                                borderRadius: '12px',
                                color: '#f1f5f9',
                                boxShadow: '0 10px 15px -3px rgba(0, 0, 0, 0.5)'
                            }}
                            itemStyle={{ marginBottom: '4px' }}
                            cursor={{ stroke: '#f97316', strokeWidth: 1, strokeDasharray: '4 4' }}
                        />
                        <Area
                            type="monotone"
                            dataKey="critical"
                            stroke="#ef4444"
                            strokeWidth={3}
                            fillOpacity={1}
                            fill="url(#colorCritical)"
                        />
                        <Area
                            type="monotone"
                            dataKey="high"
                            stroke="#f97316"
                            strokeWidth={3}
                            fillOpacity={1}
                            fill="url(#colorHigh)"
                        />
                    </AreaChart>
                </ResponsiveContainer>
            </div>

            <div className="flex items-center justify-center gap-8 mt-6">
                <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-red-500/5 border border-red-500/10">
                    <div className="w-2.5 h-2.5 rounded-full bg-red-500 shadow-[0_0_10px_rgba(239,68,68,0.5)]"></div>
                    <span className="text-xs font-medium text-slate-300">Critical</span>
                </div>
                <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-orange-500/5 border border-orange-500/10">
                    <div className="w-2.5 h-2.5 rounded-full bg-orange-500 shadow-[0_0_10px_rgba(249,115,22,0.5)]"></div>
                    <span className="text-xs font-medium text-slate-300">High</span>
                </div>
            </div>
        </motion.div>
    );
}
