export type Severity = 'info' | 'low' | 'medium' | 'high' | 'critical'
export type AlertStatus = 'open' | 'acknowledged' | 'resolved' | 'false_positive'
export type IncidentStatus = 'open' | 'in_progress' | 'resolved' | 'closed'

export interface Event {
  id: string
  timestamp: string
  received_at: string
  host: string
  source: string
  event_type: string
  severity: Severity
  message: string
  fields: Record<string, string>
  tags: string[]
  mitre_ids?: string[]
}

export interface Alert {
  id: string
  created_at: string
  updated_at: string
  rule_id: string
  rule_name: string
  severity: Severity
  status: AlertStatus
  title: string
  description: string
  host: string
  event_ids: string[]
  mitre_ids: string[]
  assigned_to?: string
  notes?: string
  incident_id?: string
  fields: Record<string, string>
}

export interface Rule {
  id: string
  name: string
  description: string
  type: 'threshold' | 'sequence' | 'anomaly' | 'correlation'
  enabled: boolean
  severity: Severity
  source?: string
  event_type?: string
  conditions: RuleCondition[]
  threshold?: number
  window_secs?: number
  group_by?: string[]
  actions: string[]
  mitre_ids?: string[]
  tags?: string[]
  created_at: string
  updated_at: string
}

export interface RuleCondition {
  field: string
  operator: string
  value: string
}

export interface Incident {
  id: string
  created_at: string
  updated_at: string
  title: string
  description: string
  severity: Severity
  status: IncidentStatus
  assigned_to?: string
  alert_ids: string[]
  tags: string[]
  timeline: TimelineEntry[]
}

export interface TimelineEntry {
  id: string
  incident_id: string
  timestamp: string
  author: string
  type: string
  content: string
}

export interface EventStats {
  total: number
  by_severity: Record<string, number>
  by_source: Record<string, number>
  by_host: Record<string, number>
}

export interface SearchResult {
  events: Event[]
  total: number
  took_ms: number
}
