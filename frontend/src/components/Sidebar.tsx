import { NavLink } from 'react-router-dom'

const nav = [
  { to: '/overview',  label: 'Overview',   icon: '◉' },
  { to: '/events',    label: 'Events',      icon: '≡' },
  { to: '/alerts',    label: 'Alerts',      icon: '⚠' },
  { to: '/search',    label: 'Search',      icon: '⌕' },
  { to: '/incidents', label: 'Incidents',   icon: '⚡' },
  { to: '/rules',     label: 'Rules',       icon: '⚙' },
]

export default function Sidebar() {
  return (
    <aside className="w-56 bg-gray-900 border-r border-gray-800 flex flex-col shrink-0">
      <div className="p-5 border-b border-gray-800">
        <h1 className="text-lg font-bold text-white tracking-tight">
          <span className="text-cyan-400">⬡</span> Portable SIEM
        </h1>
        <p className="text-xs text-gray-500 mt-0.5">Security Operations</p>
      </div>

      <nav className="flex-1 p-3 space-y-1">
        {nav.map(item => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              `flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${
                isActive
                  ? 'bg-cyan-500/10 text-cyan-400 font-medium'
                  : 'text-gray-400 hover:text-gray-200 hover:bg-gray-800'
              }`
            }
          >
            <span className="text-base">{item.icon}</span>
            {item.label}
          </NavLink>
        ))}
      </nav>

      <div className="p-4 border-t border-gray-800">
        <div className="flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
          <span className="text-xs text-gray-500">Collecting events</span>
        </div>
      </div>
    </aside>
  )
}
