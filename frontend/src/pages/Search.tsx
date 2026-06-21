import { useState } from 'react'
import { format } from 'date-fns'
import { search } from '@/api/client'
import type { Event } from '@/types'
import SeverityBadge from '@/components/SeverityBadge'

export default function Search() {
  const [query, setQuery] = useState('')
  const [source, setSource] = useState('')
  const [severity, setSeverity] = useState('')
  const [events, setEvents] = useState<Event[]>([])
  const [total, setTotal] = useState(0)
  const [took, setTook] = useState(0)
  const [loading, setLoading] = useState(false)
  const [searched, setSearched] = useState(false)

  const run = (e?: React.FormEvent) => {
    e?.preventDefault()
    setLoading(true)
    const params: Record<string, string | number> = { limit: 200 }
    if (query) params.q = query
    if (source) params.source = source
    if (severity) params.severity = severity

    search(params)
      .then(r => { setEvents(r.events ?? []); setTotal(r.total); setTook(r.took_ms); setSearched(true) })
      .catch(console.error)
      .finally(() => setLoading(false))
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-white">Search</h2>
        <p className="text-gray-500 text-sm">Full-text and field-based log search</p>
      </div>

      <form onSubmit={run} className="card space-y-4">
        <div className="flex gap-3">
          <input
            value={query}
            onChange={e => setQuery(e.target.value)}
            placeholder='Search logs... e.g. "failed password" or "root"'
            className="flex-1 bg-gray-800 border border-gray-700 text-gray-200 rounded-lg px-4 py-2.5 text-sm placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-cyan-500/50"
          />
          <button
            type="submit"
            disabled={loading}
            className="px-6 py-2.5 bg-cyan-600 hover:bg-cyan-500 text-white text-sm font-medium rounded-lg transition-colors disabled:opacity-50"
          >
            {loading ? 'Searching...' : 'Search'}
          </button>
        </div>

        <div className="flex gap-3">
          <select
            value={source}
            onChange={e => setSource(e.target.value)}
            className="bg-gray-800 border border-gray-700 text-gray-300 text-sm rounded-lg px-3 py-2"
          >
            <option value="">Any source</option>
            <option value="auth">auth</option>
            <option value="journald">journald</option>
            <option value="syslog">syslog</option>
            <option value="network_syslog">network_syslog</option>
          </select>
          <select
            value={severity}
            onChange={e => setSeverity(e.target.value)}
            className="bg-gray-800 border border-gray-700 text-gray-300 text-sm rounded-lg px-3 py-2"
          >
            <option value="">Any severity</option>
            <option value="critical">Critical</option>
            <option value="high">High</option>
            <option value="medium">Medium</option>
            <option value="low">Low</option>
            <option value="info">Info</option>
          </select>
        </div>
      </form>

      {searched && (
        <div>
          <div className="text-xs text-gray-500 mb-3">
            {total.toLocaleString()} results in {took}ms
          </div>
          <div className="card overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-gray-500 border-b border-gray-800">
                  <th className="pb-3 pr-4 font-medium">Time</th>
                  <th className="pb-3 pr-4 font-medium">Host</th>
                  <th className="pb-3 pr-4 font-medium">Source</th>
                  <th className="pb-3 pr-4 font-medium">Severity</th>
                  <th className="pb-3 font-medium">Message</th>
                </tr>
              </thead>
              <tbody>
                {events.map(ev => (
                  <tr key={ev.id} className="border-b border-gray-800/50 hover:bg-gray-800/30">
                    <td className="py-2 pr-4 text-gray-400 whitespace-nowrap font-mono text-xs">
                      {format(new Date(ev.timestamp), 'MM-dd HH:mm:ss')}
                    </td>
                    <td className="py-2 pr-4 text-gray-300">{ev.host}</td>
                    <td className="py-2 pr-4 text-cyan-400">{ev.source}</td>
                    <td className="py-2 pr-4"><SeverityBadge severity={ev.severity} /></td>
                    <td className="py-2 text-gray-300 max-w-lg truncate">{ev.message}</td>
                  </tr>
                ))}
                {events.length === 0 && (
                  <tr>
                    <td colSpan={5} className="py-8 text-center text-gray-500">No results found</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}
