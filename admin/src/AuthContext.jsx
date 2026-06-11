import { createContext, useContext, useState, useEffect } from 'react'

const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const [token, setToken]   = useState(() => localStorage.getItem('admin_token'))
  const [role, setRole]     = useState(() => localStorage.getItem('admin_role'))
  const [name, setName]     = useState(() => localStorage.getItem('admin_name') || 'Admin')

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
    <AuthContext.Provider value={{ token, role, name, login, logout, isAdmin: role === 'super_admin', isModerator: role === 'moderator' || role === 'super_admin' }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => useContext(AuthContext)
