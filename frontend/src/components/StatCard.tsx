interface Props {
  label: string
  value: number | string
  sub?: string
  accent?: string
}

export default function StatCard({ label, value, sub, accent = 'text-cyan-400' }: Props) {
  return (
    <div className="card flex flex-col gap-1">
      <span className="text-xs text-gray-500 uppercase tracking-widest">{label}</span>
      <span className={`text-3xl font-bold tabular-nums ${accent}`}>{value}</span>
      {sub && <span className="text-xs text-gray-500">{sub}</span>}
    </div>
  )
}
