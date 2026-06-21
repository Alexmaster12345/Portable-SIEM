import { useEffect, useState } from 'react'
import { format } from 'date-fns'
import { getAlerts, updateAlertStatus } from '@/api/client'
import type { Alert, AlertStatus } from '@/types'
import SeverityBadge from '@/components/SeverityBadge'

const STATUS_COLORS: Record<AlertStatus, string> = {
  open:          'text-red-400',
  acknowledged:  'text-yellow-400',
  resolved:      'text-green-400',
  false_positive: 'text-gray-400',
}

export default function Alerts() {
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [total, setTotal] = useState(0)
  const [filter, setFilter] = useState<AlertStatus | ''>('open')
  const [loading, setLoading] = useState(false)

  const load = () => {
    setLoading(true)
    getAlerts({ status: filter || undefined, limit: 100 })
      .then(r => { setAlerts(r.alerts ?? []); setTotal(r.total) })
      .catch(console.error)
      .finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [filter])

  const changeStatus = async (id: string, status: AlertStatus) => {
    await updateAlertStatus(id, status)
    load()
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Alerts</h2>
          <p className="text-gray-500 text-sm">{total} alerts</p>
        </div>
        <div className="flex gap-2">
          {(['', 'open', 'acknowledged', 'resolved'] as const).map(s => (
            <button
              key={s}
              onClick={() => setFilter(s)}
              className={`px-3 py-1.5 rounded-lg text-sm transition-colors ${
                filter === s
                  ? 'bg-cyan-500/20 text-cyan-400 border border-cyan-500/30'
                  : 'bg-gray-800 text-gray-400 hover:text-gray-200'
              }`}
            >
              {s === '' ? 'All' : s.charAt(0).toUpperCase() + s.slice(1)}
            </button>
          ))}
        </div>
      </div>

      <div className="space-y-3">
        {loading ? (
          <div className="text-gray-500 text-sm py-8 text-center">Loading...</div>
        ) : alerts.length === 0 ? (
          <div className="card text-center py-12 text-gray-500">No alerts</div>
        ) : (
          alerts.map(alert => (
            <div key={alert.id} className="card hover:border-gray-700 transition-colors">
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <SeverityBadge severity={alert.severity} />
                    <span className={`text-xs font-medium ${STATUS_COLORS[alert.status]}`}>
                      {alert.status.replace('_', ' ').toUpperCase()}
                    </span>
                  </div>
                  <h3 className="text-sm font-semibold text-gray-200">{alert.title}</h3>
                  <p className="text-xs text-gray-500 mt-1">{alert.description}</p>
                  <div className="flex items-center gap-3 mt-2 text-xs text-gray-500">
                    <span>Rule: {alert.rule_name}</span>
                    <span>Host: {alert.host}</span>
                    <span>{format(new Date(alert.created_at), 'MMM d, HH:mm')}</span>
                    {alert.mitre_ids?.length > 0 && (
                      <span className="text-purple-400">{alert.mitre_ids.join(', ')}</span>
                    )}
                  </div>
                </div>
                <div className="flex gap-2 shrink-0">
                  {alert.status === 'open' && (
                    <button
                      onClick={() => changeStatus(alert.id, 'acknowledged')}
                      className="text-xs px-2 py-1 rounded bg-yellow-500/10 text-yellow-400 hover:bg-yellow-500/20 border border-yellow-500/20"
                    >
                      Ack
                    </button>
                  )}
                  {alert.status !== 'resolved' && (
                    <button
                      onClick={() => changeStatus(alert.id, 'resolved')}
                      className="text-xs px-2 py-1 rounded bg-green-500/10 text-green-400 hover:bg-green-500/20 border border-green-500/20"
                    >
                      Resolve
                    </button>
                  )}
                  {alert.status !== 'false_positive' && (
                    <button
                      onClick={() => changeStatus(alert.id, 'false_positive')}
                      className="text-xs px-2 py-1 rounded bg-gray-500/10 text-gray-400 hover:bg-gray-500/20 border border-gray-500/20"
                    >
                      FP
                    </button>
                  )}
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
