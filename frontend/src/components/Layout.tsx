import type { ReactNode } from 'react'
import Sidebar from './Sidebar'

interface Props { children: ReactNode }

export default function Layout({ children }: Props) {
  return (
    <div className="flex h-screen bg-gray-950 overflow-hidden">
      <Sidebar />
      <main className="flex-1 overflow-auto p-6">
        {children}
      </main>
    </div>
  )
}
