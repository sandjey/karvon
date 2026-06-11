import { createContext, useContext, useState } from 'react'

const AuthContext = createContext(null)

function parseJwt(token) {
  try { return JSON.parse(atob(token.split('.')[1])) } catch { return {} }
}

function resolveRole() {
  const stored = localStorage.getItem('admin_role')
  if (stored && stored !== 'undefined' && stored !== 'null') return stored
  const tok = localStorage.getItem('admin_token')
  if (!tok) return null
  const payload = parseJwt(tok)
  if (payload.role) { localStorage.setItem('admin_role', payload.role); return payload.role }
  return null
}

export function AuthProvider({ children }) {
  const [token, setToken] = useState(() => localStorage.getItem('admin_token'))
  const [role, setRole]   = useState(() => resolveRole())
  const [name, setName]   = useState(() => localStorage.getItem('admin_name') || 'Admin')

  const login = (accessToken, userRole, userName) => {
    localStorage.setItem('admin_token', accessToken)
    localStorage.setItem('admin_role', userRole)
    localStorage.setItem('admin_name', userName || 'Admin')
    setToken(accessToken)
    setRole(userRole)
    setName(userName || 'Admin')
  }

  const logout = () => {
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_role')
    localStorage.removeItem('admin_name')
    setToken(null)
    setRole(null)
  }

  return (
    <AuthContext.Provider value={{
      token, role, name, login, logout,
      isAdmin: role === 'super_admin',
      isModerator: role === 'moderator' || role === 'super_admin',
    }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => useContext(AuthContext)
