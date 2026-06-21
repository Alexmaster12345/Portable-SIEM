import { useEffect, useState } from 'react'
import { format } from 'date-fns'
import { getIncidents, createIncident, updateIncidentStatus } from '@/api/client'
import type { Incident, IncidentStatus } from '@/types'
import SeverityBadge from '@/components/SeverityBadge'

const STATUS_COLOR: Record<IncidentStatus, string> = {
  open:        'text-red-400',
  in_progress: 'text-yellow-400',
  resolved:    'text-green-400',
  closed:      'text-gray-500',
}

export default function Incidents() {
  const [incidents, setIncidents] = useState<Incident[]>([])
  const [creating, setCreating] = useState(false)
  const [title, setTitle] = useState('')
  const [desc, setDesc] = useState('')

  const load = () => {
    getIncidents({ limit: 50 }).then(r => setIncidents(r.incidents ?? [])).catch(console.error)
  }
  useEffect(() => { load() }, [])

  const create = async (e: React.FormEvent) => {
    e.preventDefault()
    await createIncident({ title, description: desc, severity: 'high', tags: [] })
    setTitle('')
    setDesc('')
    setCreating(false)
    load()
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Incidents</h2>
          <p className="text-gray-500 text-sm">{incidents.length} incidents</p>
        </div>
        <button
          onClick={() => setCreating(c => !c)}
          className="px-4 py-2 bg-cyan-600 hover:bg-cyan-500 text-white text-sm font-medium rounded-lg transition-colors"
        >
          + New Incident
        </button>
      </div>

      {creating && (
        <form onSubmit={create} className="card space-y-3">
          <h3 className="text-sm font-semibold text-gray-300">Create Incident</h3>
          <input
            value={title}
            onChange={e => setTitle(e.target.value)}
            placeholder="Incident title"
            required
            className="w-full bg-gray-800 border border-gray-700 text-gray-200 rounded-lg px-3 py-2 text-sm"
          />
          <textarea
            value={desc}
            onChange={e => setDesc(e.target.value)}
            placeholder="Description"
            rows={3}
            className="w-full bg-gray-800 border border-gray-700 text-gray-200 rounded-lg px-3 py-2 text-sm resize-none"
          />
          <div className="flex gap-2">
            <button type="submit" className="px-4 py-1.5 bg-cyan-600 text-white text-sm rounded-lg">Create</button>
            <button type="button" onClick={() => setCreating(false)} className="px-4 py-1.5 bg-gray-700 text-gray-300 text-sm rounded-lg">Cancel</button>
          </div>
        </form>
      )}

      <div className="space-y-3">
        {incidents.length === 0 ? (
          <div className="card text-center py-12 text-gray-500">No incidents</div>
        ) : (
          incidents.map(inc => (
            <div key={inc.id} className="card hover:border-gray-700 transition-colors">
              <div className="flex items-start justify-between gap-4">
                <div>
                  <div className="flex items-center gap-2 mb-1">
                    <SeverityBadge severity={inc.severity} />
                    <span className={`text-xs font-medium ${STATUS_COLOR[inc.status]}`}>
                      {inc.status.replace('_', ' ').toUpperCase()}
                    </span>
                  </div>
                  <h3 className="text-sm font-semibold text-gray-200">{inc.title}</h3>
                  {inc.description && <p className="text-xs text-gray-500 mt-1">{inc.description}</p>}
                  <div className="text-xs text-gray-500 mt-2">
                    Created {format(new Date(inc.created_at), 'MMM d, HH:mm')}
                    {inc.alert_ids?.length > 0 && ` · ${inc.alert_ids.length} alerts`}
                  </div>
                </div>
                <div className="flex gap-2 shrink-0">
                  {inc.status === 'open' && (
                    <button
                      onClick={() => updateIncidentStatus(inc.id, 'in_progress').then(load)}
                      className="text-xs px-2 py-1 rounded bg-yellow-500/10 text-yellow-400 border border-yellow-500/20 hover:bg-yellow-500/20"
                    >
                      Start
                    </button>
                  )}
                  {inc.status !== 'resolved' && (
                    <button
                      onClick={() => updateIncidentStatus(inc.id, 'resolved').then(load)}
                      className="text-xs px-2 py-1 rounded bg-green-500/10 text-green-400 border border-green-500/20 hover:bg-green-500/20"
                    >
                      Resolve
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
