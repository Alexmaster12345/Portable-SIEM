import { Routes, Route, Navigate } from 'react-router-dom'
import Layout from '@/components/Layout'
import Overview from '@/pages/Overview'
import Events from '@/pages/Events'
import Alerts from '@/pages/Alerts'
import Search from '@/pages/Search'
import Incidents from '@/pages/Incidents'
import Rules from '@/pages/Rules'

export default function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Navigate to="/overview" replace />} />
        <Route path="/overview" element={<Overview />} />
        <Route path="/events" element={<Events />} />
        <Route path="/alerts" element={<Alerts />} />
        <Route path="/search" element={<Search />} />
        <Route path="/incidents" element={<Incidents />} />
        <Route path="/rules" element={<Rules />} />
      </Routes>
    </Layout>
  )
}
