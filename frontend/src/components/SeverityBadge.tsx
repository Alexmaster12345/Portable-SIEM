import type { Severity } from '@/types'

const classes: Record<Severity, string> = {
  critical: 'badge-critical',
  high:     'badge-high',
  medium:   'badge-medium',
  low:      'badge-low',
  info:     'badge-info',
}

export default function SeverityBadge({ severity }: { severity: Severity }) {
  return (
    <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${classes[severity] ?? classes.info}`}>
      {severity.toUpperCase()}
    </span>
  )
}
