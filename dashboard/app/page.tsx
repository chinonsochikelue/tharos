import { MetricCard } from "@/components/metric-card";
import { TrendChart } from "@/components/trend-chart";
import { RecentFindings } from "@/components/recent-findings";
import { QuickActions } from "@/components/quick-actions";
import { Shield, AlertTriangle, CheckCircle, TrendingUp } from "lucide-react";

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-bold gradient-text">Security Dashboard</h1>
        <p className="text-slate-400 mt-1">
          Real-time overview of your security posture
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <MetricCard
          title="Security Score"
          value="87"
          suffix="/100"
          trend="+5%"
          trendUp={true}
          icon={<Shield className="w-6 h-6" />}
          color="orange"
        />
        <MetricCard
          title="Critical Issues"
          value="3"
          trend="-2"
          trendUp={true}
          icon={<AlertTriangle className="w-6 h-6" />}
          color="red"
        />
        <MetricCard
          title="Resolved Today"
          value="12"
          trend="+8"
          trendUp={true}
          icon={<CheckCircle className="w-6 h-6" />}
          color="green"
        />
        <MetricCard
          title="Total Scans"
          value="1,247"
          trend="+156"
          trendUp={true}
          icon={<TrendingUp className="w-6 h-6" />}
          color="blue"
        />
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <TrendChart />
        <QuickActions />
      </div>

      {/* Recent Findings */}
      <RecentFindings />
    </div>
  );
}
