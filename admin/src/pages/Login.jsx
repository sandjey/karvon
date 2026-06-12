import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../AuthContext'
import { api } from '../api'
import { IcoUser, IcoLock } from '../icons'

function parseJwt(token) {
  try { return JSON.parse(atob(token.split('.')[1])) } catch { return {} }
}

const TruckSvg = () => (
  <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M1 3h15v13H1z"/><path d="M16 8h4l3 3v5h-7V8z"/>
    <circle cx="5.5" cy="18.5" r="2.5"/><circle cx="18.5" cy="18.5" r="2.5"/>
  </svg>
)

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
      background: 'linear-gradient(135deg, #f0f4ff 0%, #f8fafc 60%, #f0fdf4 100%)',
      padding: 16,
    }}>
      <div style={{ width: '100%', maxWidth: 380 }}>

        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <div style={{
            width: 62, height: 62, borderRadius: 18, margin: '0 auto 16px',
            background: 'linear-gradient(135deg, #2563eb, #3b82f6)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            boxShadow: '0 8px 28px rgba(37,99,235,.3)',
          }}>
            <TruckSvg />
          </div>
          <div style={{ color: '#0f172a', fontSize: 26, fontWeight: 800, letterSpacing: '.01em' }}>KARVON</div>
          <div style={{ color: '#64748b', fontSize: 13.5, marginTop: 6 }}>Панель управления</div>
        </div>

        <div style={{
          background: '#ffffff', border: '1px solid #e2e8f0',
          borderRadius: 16, padding: 32,
          boxShadow: '0 4px 24px rgba(15,23,42,.08)',
        }}>
          {error && (
            <div style={{
              background: '#fef2f2', border: '1px solid #fecaca',
              color: '#dc2626', borderRadius: 8, padding: '10px 14px',
              fontSize: 13, marginBottom: 20,
            }}>{error}</div>
          )}

          <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <div>
              <label style={{ display: 'block', fontSize: 12, fontWeight: 600, color: '#374151', marginBottom: 8 }}>Логин</label>
              <div style={{ position: 'relative' }}>
                <div style={{ position: 'absolute', left: 11, top: '50%', transform: 'translateY(-50%)', color: '#9ca3af', pointerEvents: 'none' }}>
                  <IcoUser size={15} />
                </div>
                <input
                  className="input" style={{ paddingLeft: 36 }}
                  placeholder="Введите логин"
                  value={login}
                  onChange={e => setLogin(e.target.value)}
                  autoFocus autoComplete="username"
                />
              </div>
            </div>

            <div>
              <label style={{ display: 'block', fontSize: 12, fontWeight: 600, color: '#374151', marginBottom: 8 }}>Пароль</label>
              <div style={{ position: 'relative' }}>
                <div style={{ position: 'absolute', left: 11, top: '50%', transform: 'translateY(-50%)', color: '#9ca3af', pointerEvents: 'none' }}>
                  <IcoLock size={15} />
                </div>
                <input
                  type="password" className="input" style={{ paddingLeft: 36 }}
                  placeholder="Введите пароль"
                  value={password}
                  onChange={e => setPassword(e.target.value)}
                  autoComplete="current-password"
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={loading || !login || !password}
              className="btn-primary"
              style={{ marginTop: 8, width: '100%', padding: '12px', fontSize: 14, justifyContent: 'center', borderRadius: 10 }}
            >
              {loading ? 'Вход...' : 'Войти в систему'}
            </button>
          </form>
        </div>

        <div style={{ textAlign: 'center', marginTop: 20, color: '#94a3b8', fontSize: 12 }}>
          KARVON Admin Panel v1.0
        </div>
      </div>
    </div>
  )
}
