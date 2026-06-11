import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider, useAuth } from './AuthContext'
import Layout from './Layout'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Users from './pages/Users'
import Moderators from './pages/Moderators'
import Listings from './pages/Listings'
import Companies from './pages/Companies'
import Payments from './pages/Payments'
import Pricing from './pages/Pricing'
import Queue from './pages/Queue'
import History from './pages/History'

function RequireAuth({ children, adminOnly = false }) {
  const { token, isAdmin } = useAuth()
  if (!token) return <Navigate to="/login" replace />
  if (adminOnly && !isAdmin) return <Navigate to="/queue" replace />
  return children
}

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter basename={import.meta.env.BASE_URL}>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/" element={<RequireAuth><Layout /></RequireAuth>}>
            <Route index element={<Navigate to="/dashboard" replace />} />
            <Route path="dashboard" element={<RequireAuth adminOnly><Dashboard /></RequireAuth>} />
            <Route path="users"      element={<RequireAuth adminOnly><Users /></RequireAuth>} />
            <Route path="moderators" element={<RequireAuth adminOnly><Moderators /></RequireAuth>} />
            <Route path="listings"   element={<RequireAuth adminOnly><Listings /></RequireAuth>} />
            <Route path="companies"  element={<RequireAuth adminOnly><Companies /></RequireAuth>} />
            <Route path="payments"   element={<RequireAuth adminOnly><Payments /></RequireAuth>} />
            <Route path="pricing"    element={<RequireAuth adminOnly><Pricing /></RequireAuth>} />
            <Route path="queue"      element={<RequireAuth><Queue /></RequireAuth>} />
            <Route path="history"    element={<RequireAuth><History /></RequireAuth>} />
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}
