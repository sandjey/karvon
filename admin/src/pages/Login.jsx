import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../AuthContext'
import { api } from '../api'
import { IcoUser, IcoLock } from '../icons'

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
    } catch {
      setError('Неверный логин или пароль')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{
      minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center',
      background: '#070c17', padding: 16,
    }}>
      <div style={{ width: '100%', maxWidth: 380 }}>

        {/* Logo */}
        <div style={{ textAlign: 'center', marginBottom: 36 }}>
          <div style={{
            width: 52, height: 52, borderRadius: 13, margin: '0 auto 14px',
            background: 'linear-gradient(135deg,#1d56d4,#2563eb)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            boxShadow: '0 8px 32px rgba(29,86,212,0.35)',
          }}>
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M1 3h15v13H1z"/><path d="M16 8h4l3 3v5h-7V8z"/><circle cx="5.5" cy="18.5" r="2.5"/><circle cx="18.5" cy="18.5" r="2.5"/>
            </svg>
          </div>
          <div style={{ color: '#dce8f5', fontSize: 22, fontWeight: 800, letterSpacing: '.02em' }}>KARVON</div>
          <div style={{ color: '#2d4a6e', fontSize: 13, marginTop: 4 }}>Панель администратора</div>
        </div>

        {/* Form card */}
        <div style={{
          background: '#0c1526', border: '1px solid #19273d',
          borderRadius: 14, padding: 28,
          boxShadow: '0 20px 60px rgba(0,0,0,0.5)',
        }}>

          {error && (
            <div style={{
              background: 'rgba(232,64,64,0.1)', border: '1px solid rgba(232,64,64,0.2)',
              color: '#f87171', borderRadius: 8, padding: '10px 14px',
              fontSize: 13, marginBottom: 18,
            }}>{error}</div>
          )}

          <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            {/* Login field */}
            <div>
              <label style={{ display: 'block', fontSize: 12, fontWeight: 600, color: '#3d5a7d', marginBottom: 7, textTransform: 'uppercase', letterSpacing: '.06em' }}>
                Логин
              </label>
              <div style={{ position: 'relative' }}>
                <div style={{ position: 'absolute', left: 11, top: '50%', transform: 'translateY(-50%)', color: '#2d4a6e', pointerEvents: 'none' }}>
                  <IcoUser size={14} />
                </div>
                <input
                  className="input"
                  style={{ paddingLeft: 34 }}
                  placeholder="admin"
                  value={login}
                  onChange={e => setLogin(e.target.value)}
                  autoFocus
                  autoComplete="username"
                />
              </div>
            </div>

            {/* Password field */}
            <div>
              <label style={{ display: 'block', fontSize: 12, fontWeight: 600, color: '#3d5a7d', marginBottom: 7, textTransform: 'uppercase', letterSpacing: '.06em' }}>
                Пароль
              </label>
              <div style={{ position: 'relative' }}>
                <div style={{ position: 'absolute', left: 11, top: '50%', transform: 'translateY(-50%)', color: '#2d4a6e', pointerEvents: 'none' }}>
                  <IcoLock size={14} />
                </div>
                <input
                  type="password"
                  className="input"
                  style={{ paddingLeft: 34 }}
                  placeholder="••••••••"
                  value={password}
                  onChange={e => setPassword(e.target.value)}
                  autoComplete="current-password"
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={loading || !login || !password}
              style={{
                marginTop: 8, width: '100%', padding: '11px',
                background: loading ? '#132b5a' : '#1d56d4',
                color: loading ? '#2d4a6e' : '#fff',
                border: 'none', borderRadius: 9, cursor: loading ? 'not-allowed' : 'pointer',
                fontSize: 14, fontWeight: 700, letterSpacing: '.02em',
                transition: 'all .15s',
              }}
            >
              {loading ? 'Вход...' : 'Войти'}
            </button>
          </form>
        </div>

        <div style={{ textAlign: 'center', marginTop: 20, color: '#19273d', fontSize: 12 }}>
          KARVON Admin Panel v1.0
        </div>
      </div>
    </div>
  )
}
