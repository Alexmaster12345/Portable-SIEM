import { useEffect, useState } from 'react'
import { format } from 'date-fns'
import { getEvents } from '@/api/client'
import type { Event } from '@/types'
import SeverityBadge from '@/components/SeverityBadge'

export default function Events() {
  const [events, setEvents] = useState<Event[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [source, setSource] = useState('')
  const [offset, setOffset] = useState(0)
  const limit = 50

  const load = (off = 0) => {
    setLoading(true)
    const params: Record<string, string | number> = { limit, offset: off }
    if (source) params.source = source
    getEvents(params)
      .then(r => { setEvents(r.events ?? []); setTotal(r.total) })
      .catch(console.error)
      .finally(() => setLoading(false))
  }

  useEffect(() => { load(0) }, [source])

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Events</h2>
          <p className="text-gray-500 text-sm">{total.toLocaleString()} total events</p>
        </div>
        <select
          value={source}
          onChange={e => setSource(e.target.value)}
          className="bg-gray-800 border border-gray-700 text-gray-200 text-sm rounded-lg px-3 py-2"
        >
          <option value="">All sources</option>
          <option value="auth">auth</option>
          <option value="journald">journald</option>
          <option value="syslog">syslog</option>
          <option value="network_syslog">network_syslog</option>
          <option value="agent">agent</option>
        </select>
      </div>

      <div className="card overflow-x-auto">
        {loading ? (
          <div className="text-gray-500 text-sm py-8 text-center">Loading...</div>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-gray-500 border-b border-gray-800">
                <th className="pb-3 pr-4 font-medium">Time</th>
                <th className="pb-3 pr-4 font-medium">Host</th>
                <th className="pb-3 pr-4 font-medium">Source</th>
                <th className="pb-3 pr-4 font-medium">Type</th>
                <th className="pb-3 pr-4 font-medium">Severity</th>
                <th className="pb-3 font-medium">Message</th>
              </tr>
            </thead>
            <tbody>
              {events.map(ev => (
                <tr key={ev.id} className="border-b border-gray-800/50 hover:bg-gray-800/30 transition-colors">
                  <td className="py-2 pr-4 text-gray-400 whitespace-nowrap font-mono text-xs">
                    {format(new Date(ev.timestamp), 'MM-dd HH:mm:ss')}
                  </td>
                  <td className="py-2 pr-4 text-gray-300 whitespace-nowrap">{ev.host}</td>
                  <td className="py-2 pr-4 text-cyan-400 whitespace-nowrap">{ev.source}</td>
                  <td className="py-2 pr-4 text-gray-400 whitespace-nowrap">{ev.event_type}</td>
                  <td className="py-2 pr-4">
                    <SeverityBadge severity={ev.severity} />
                  </td>
                  <td className="py-2 text-gray-300 max-w-md truncate">{ev.message}</td>
                </tr>
              ))}
              {events.length === 0 && (
                <tr>
                  <td colSpan={6} className="py-8 text-center text-gray-500">No events found</td>
                </tr>
              )}
            </tbody>
          </table>
        )}
      </div>

      {total > limit && (
        <div className="flex items-center justify-between text-sm text-gray-500">
          <span>{offset + 1}–{Math.min(offset + limit, total)} of {total.toLocaleString()}</span>
          <div className="flex gap-2">
            <button
              disabled={offset === 0}
              onClick={() => { setOffset(o => Math.max(0, o - limit)); load(Math.max(0, offset - limit)) }}
              className="px-3 py-1 rounded bg-gray-800 disabled:opacity-40"
            >
              Prev
            </button>
            <button
              disabled={offset + limit >= total}
              onClick={() => { setOffset(o => o + limit); load(offset + limit) }}
              className="px-3 py-1 rounded bg-gray-800 disabled:opacity-40"
            >
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
