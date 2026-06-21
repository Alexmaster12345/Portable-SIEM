import { useEffect, useState } from 'react'
import { getRules, deleteRule } from '@/api/client'
import type { Rule } from '@/types'
import SeverityBadge from '@/components/SeverityBadge'

export default function Rules() {
  const [rules, setRules] = useState<Rule[]>([])

  const load = () => {
    getRules().then(r => setRules(r.rules ?? [])).catch(console.error)
  }
  useEffect(() => { load() }, [])

  const remove = async (id: string) => {
    await deleteRule(id)
    load()
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-white">Detection Rules</h2>
        <p className="text-gray-500 text-sm">{rules.length} rules loaded</p>
      </div>

      <div className="space-y-3">
        {rules.length === 0 ? (
          <div className="card text-center py-12 text-gray-500">No rules loaded</div>
        ) : (
          rules.map(rule => (
            <div key={rule.id} className="card hover:border-gray-700 transition-colors">
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-1">
                    <SeverityBadge severity={rule.severity} />
                    <span className={`text-xs px-2 py-0.5 rounded-full border ${rule.enabled ? 'text-green-400 bg-green-500/10 border-green-500/20' : 'text-gray-500 bg-gray-500/10 border-gray-500/20'}`}>
                      {rule.enabled ? 'ENABLED' : 'DISABLED'}
                    </span>
                    <span className="text-xs text-purple-400 bg-purple-500/10 px-2 py-0.5 rounded-full border border-purple-500/20">
                      {rule.type}
                    </span>
                  </div>
                  <h3 className="text-sm font-semibold text-gray-200">{rule.name}</h3>
                  <p className="text-xs text-gray-500 mt-1">{rule.description}</p>
                  <div className="flex items-center gap-4 mt-2 text-xs text-gray-500">
                    {rule.source && <span>Source: <span className="text-cyan-400">{rule.source}</span></span>}
                    {rule.event_type && <span>Type: <span className="text-cyan-400">{rule.event_type}</span></span>}
                    {rule.threshold && <span>Threshold: <span className="text-yellow-400">{rule.threshold}</span></span>}
                    {rule.window_secs && <span>Window: <span className="text-yellow-400">{rule.window_secs}s</span></span>}
                    {rule.mitre_ids && rule.mitre_ids.length > 0 && (
                      <span className="text-purple-400">{rule.mitre_ids.join(', ')}</span>
                    )}
                  </div>
                  {rule.group_by && rule.group_by.length > 0 && (
                    <div className="text-xs text-gray-500 mt-1">Group by: {rule.group_by.join(', ')}</div>
                  )}
                </div>
                <button
                  onClick={() => remove(rule.id)}
                  className="text-xs px-2 py-1 rounded bg-red-500/10 text-red-400 border border-red-500/20 hover:bg-red-500/20 shrink-0"
                >
                  Delete
                </button>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
