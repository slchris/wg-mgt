import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from './stores/auth'
import Layout from './components/Layout'
import Login from './pages/Login'
import Setup from './pages/Setup'
import Dashboard from './pages/Dashboard'
import Nodes from './pages/Nodes'
import NodeDetail from './pages/NodeDetail'
import Peers from './pages/Peers'
import Networks from './pages/Networks'
import Settings from './pages/Settings'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const { token, _hasHydrated } = useAuthStore()
  
  // Wait for hydration to complete before checking auth
  if (!_hasHydrated) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-apple-gray-50">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-apple-blue"></div>
      </div>
    )
  }
  
  if (!token) {
    return <Navigate to="/login" replace />
  }
  return <>{children}</>
}

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/setup" element={<Setup />} />
        <Route
          path="/"
          element={
            <PrivateRoute>
              <Layout />
            </PrivateRoute>
          }
        >
          <Route index element={<Dashboard />} />
          <Route path="nodes" element={<Nodes />} />
          <Route path="nodes/:id" element={<NodeDetail />} />
          <Route path="peers" element={<Peers />} />
          <Route path="networks" element={<Networks />} />
          <Route path="settings" element={<Settings />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App
