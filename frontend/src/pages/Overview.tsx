import { useEffect, useState } from 'react'
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts'
import { getEventStats, getAlerts } from '@/api/client'
import type { Alert, EventStats } from '@/types'
import StatCard from '@/components/StatCard'
import SeverityBadge from '@/components/SeverityBadge'

const COLORS = { critical: '#ef4444', high: '#f97316', medium: '#eab308', low: '#3b82f6', info: '#6b7280' }

export default function Overview() {
  const [stats, setStats] = useState<EventStats | null>(null)
  const [alerts, setAlerts] = useState<Alert[]>([])

  useEffect(() => {
    getEventStats().then(setStats).catch(console.error)
    getAlerts({ status: 'open', limit: 5 }).then(r => setAlerts(r.alerts ?? [])).catch(console.error)

    const id = setInterval(() => {
      getEventStats().then(setStats).catch(console.error)
    }, 30_000)
    return () => clearInterval(id)
  }, [])

  const sourceData = Object.entries(stats?.by_source ?? {}).map(([name, value]) => ({ name, value }))
  const severityData = Object.entries(stats?.by_severity ?? {}).map(([name, value]) => ({ name, value }))

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-white">Overview</h2>
        <p className="text-gray-500 text-sm mt-1">Last 24 hours</p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard label="Total Events" value={stats?.total?.toLocaleString() ?? '—'} />
        <StatCard label="Open Alerts" value={alerts.length} accent="text-red-400" />
        <StatCard label="Critical" value={stats?.by_severity?.critical ?? 0} accent="text-red-400" />
        <StatCard label="High" value={stats?.by_severity?.high ?? 0} accent="text-orange-400" />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Events by Source */}
        <div className="card">
          <h3 className="text-sm font-semibold text-gray-300 mb-4">Events by Source</h3>
          <ResponsiveContainer width="100%" height={200}>
            <BarChart data={sourceData} margin={{ left: -20 }}>
              <XAxis dataKey="name" tick={{ fill: '#6b7280', fontSize: 11 }} />
              <YAxis tick={{ fill: '#6b7280', fontSize: 11 }} />
              <Tooltip
                contentStyle={{ background: '#111827', border: '1px solid #1f2937', borderRadius: 8 }}
                labelStyle={{ color: '#e5e7eb' }}
              />
              <Bar dataKey="value" fill="#06b6d4" radius={[3, 3, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>

        {/* Events by Severity */}
        <div className="card">
          <h3 className="text-sm font-semibold text-gray-300 mb-4">Events by Severity</h3>
          <ResponsiveContainer width="100%" height={200}>
            <PieChart>
              <Pie data={severityData} dataKey="value" nameKey="name" cx="50%" cy="50%" outerRadius={80} label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                {severityData.map(entry => (
                  <Cell key={entry.name} fill={COLORS[entry.name as keyof typeof COLORS] ?? '#6b7280'} />
                ))}
              </Pie>
              <Tooltip contentStyle={{ background: '#111827', border: '1px solid #1f2937', borderRadius: 8 }} />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Recent Alerts */}
      <div className="card">
        <h3 className="text-sm font-semibold text-gray-300 mb-4">Recent Open Alerts</h3>
        {alerts.length === 0 ? (
          <p className="text-gray-500 text-sm">No open alerts</p>
        ) : (
          <div className="space-y-3">
            {alerts.map(alert => (
              <div key={alert.id} className="flex items-start justify-between gap-4 py-2 border-b border-gray-800 last:border-0">
                <div>
                  <p className="text-sm text-gray-200 font-medium">{alert.title}</p>
                  <p className="text-xs text-gray-500 mt-0.5">{alert.rule_name} · {alert.host}</p>
                </div>
                <SeverityBadge severity={alert.severity} />
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Top Hosts */}
      {stats?.by_host && (
        <div className="card">
          <h3 className="text-sm font-semibold text-gray-300 mb-4">Top Hosts by Event Volume</h3>
          <div className="space-y-2">
            {Object.entries(stats.by_host).slice(0, 8).map(([host, count]) => (
              <div key={host} className="flex items-center gap-3">
                <span className="text-sm text-gray-300 w-40 truncate">{host}</span>
                <div className="flex-1 bg-gray-800 rounded-full h-2">
                  <div
                    className="bg-cyan-500 h-2 rounded-full"
                    style={{ width: `${Math.min(100, (count / (stats.total || 1)) * 100 * 10)}%` }}
                  />
                </div>
                <span className="text-xs text-gray-500 w-12 text-right">{count.toLocaleString()}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
