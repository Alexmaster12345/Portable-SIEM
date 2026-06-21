import axios from 'axios'
import type { Alert, AlertStatus, Event, EventStats, Incident, IncidentStatus, Rule, SearchResult, TimelineEntry } from '@/types'

const http = axios.create({
  baseURL: '/api/v1',
  headers: { 'Content-Type': 'application/json' },
})

// Events
export const getEvents = (params?: Record<string, string | number>) =>
  http.get<{ events: Event[]; total: number }>('/events', { params }).then(r => r.data)

export const getEventStats = (since?: string) =>
  http.get<EventStats>('/events/stats', { params: since ? { since } : undefined }).then(r => r.data)

export const ingestEvent = (event: Partial<Event>) =>
  http.post<Event>('/events', event).then(r => r.data)

// Search
export const search = (params: Record<string, string | number>) =>
  http.get<SearchResult>('/search', { params }).then(r => r.data)

// Alerts
export const getAlerts = (params?: { status?: AlertStatus; limit?: number; offset?: number }) =>
  http.get<{ alerts: Alert[]; total: number }>('/alerts', { params }).then(r => r.data)

export const updateAlertStatus = (id: string, status: AlertStatus, notes?: string) =>
  http.patch(`/alerts/${id}/status`, { status, notes }).then(r => r.data)

// Rules
export const getRules = () =>
  http.get<{ rules: Rule[] }>('/rules').then(r => r.data)

export const createRule = (rule: Partial<Rule>) =>
  http.post<Rule>('/rules', rule).then(r => r.data)

export const updateRule = (id: string, rule: Partial<Rule>) =>
  http.put<Rule>(`/rules/${id}`, rule).then(r => r.data)

export const deleteRule = (id: string) =>
  http.delete(`/rules/${id}`).then(r => r.data)

// Incidents
export const getIncidents = (params?: { status?: IncidentStatus; limit?: number }) =>
  http.get<{ incidents: Incident[] }>('/incidents', { params }).then(r => r.data)

export const getIncident = (id: string) =>
  http.get<Incident>(`/incidents/${id}`).then(r => r.data)

export const createIncident = (inc: Partial<Incident>) =>
  http.post<Incident>('/incidents', inc).then(r => r.data)

export const updateIncidentStatus = (id: string, status: IncidentStatus) =>
  http.patch(`/incidents/${id}/status`, { status }).then(r => r.data)

export const addTimelineEntry = (incidentId: string, entry: Partial<TimelineEntry>) =>
  http.post<TimelineEntry>(`/incidents/${incidentId}/timeline`, entry).then(r => r.data)
