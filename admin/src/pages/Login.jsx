import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../AuthContext'
import { api } from '../api'

function parseJwt(token) {
  try { return JSON.parse(atob(token.split('.')[1])) } catch { return {} }
}

export default function Login() {
  const [login, setLogin]       = useState('')
  const [password, setPassword] = useState('')
  const [error, setError]       = useState('')
  const [loading, setLoading]   = useState(false)
  const { login: doLogin } = useAuth()
  const navigate = useNavigate()

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const data = await api.adminLogin(login, password)
      const tokens = data.data
      const payload = parseJwt(tokens.access_token)
      doLogin(tokens.access_token, payload.role, payload.name || login)
      navigate(payload.role === 'super_admin' ? '/dashboard' : '/queue')
    } catch (err) {
      setError('Неверный логин или пароль')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center" style={{background:'#0d1b2e'}}>
      <div className="w-full max-w-sm">
        {/* Logo */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-14 h-14 bg-blue-600 rounded-2xl mb-4">
            <span className="text-white font-bold text-xl">CL</span>
          </div>
          <h1 className="text-white text-2xl font-bold">CargoLink Admin</h1>
          <p className="text-slate-400 text-sm mt-1">Панель управления</p>
        </div>

        <form onSubmit={handleSubmit} className="bg-white/5 border border-white/10 rounded-2xl p-6 space-y-4">
          {error && (
            <div className="bg-red-500/10 border border-red-500/20 text-red-400 rounded-lg px-4 py-3 text-sm">
              {error}
            </div>
          )}

          <div>
            <label className="block text-slate-300 text-sm font-medium mb-1.5">Логин</label>
            <input
              className="w-full bg-white/10 border border-white/10 rounded-lg px-4 py-2.5 text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
              placeholder="karvonadmin"
              value={login}
              onChange={e => setLogin(e.target.value)}
              required
            />
          </div>

          <div>
            <label className="block text-slate-300 text-sm font-medium mb-1.5">Пароль</label>
            <input
              type="password"
              className="w-full bg-white/10 border border-white/10 rounded-lg px-4 py-2.5 text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
              placeholder="••••••••"
              value={password}
              onChange={e => setPassword(e.target.value)}
              required
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-blue-600 hover:bg-blue-700 disabled:opacity-60 text-white font-semibold py-2.5 rounded-lg transition text-sm mt-2"
          >
            {loading ? 'Вход...' : 'Войти'}
          </button>
        </form>

        <p className="text-center text-slate-600 text-xs mt-6">
          Модераторы входят через этот же экран
        </p>
      </div>
    </div>
  )
}
