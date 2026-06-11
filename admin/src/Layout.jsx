import { Outlet, NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from './AuthContext'
import {
  IcoDashboard, IcoUsers, IcoShield, IcoBuilding, IcoList,
  IcoCard, IcoTag, IcoInbox, IcoHistory, IcoLogout, IcoUser,
} from './icons'

const adminNav = [
  { to: '/dashboard',  label: 'Дашборд',      Icon: IcoDashboard },
  { to: '/users',      label: 'Пользователи', Icon: IcoUsers },
  { to: '/moderators', label: 'Модераторы',   Icon: IcoShield },
  { to: '/companies',  label: 'Компании',     Icon: IcoBuilding },
  { to: '/listings',   label: 'Объявления',   Icon: IcoList },
  { to: '/payments',   label: 'Платежи',      Icon: IcoCard },
  { to: '/pricing',    label: 'Тарифы',       Icon: IcoTag },
]

const modNav = [
  { to: '/queue',   label: 'Очередь',  Icon: IcoInbox },
  { to: '/history', label: 'История',  Icon: IcoHistory },
]

export default function Layout() {
  const { name, role, logout, isAdmin } = useAuth()
  const navigate = useNavigate()

  const handleLogout = () => { logout(); navigate('/login') }
  const initials = (name || 'KA').slice(0, 2).toUpperCase()

  return (
    <div style={{ display: 'flex', height: '100vh', overflow: 'hidden', background: '#070c17' }}>

      {/* ── Sidebar ── */}
      <aside style={{
        width: 228, flexShrink: 0, display: 'flex', flexDirection: 'column',
        background: '#060b15', borderRight: '1px solid #141f33',
      }}>

        {/* Logo */}
        <div style={{
          display: 'flex', alignItems: 'center', gap: 10,
          padding: '20px 16px 18px', borderBottom: '1px solid #141f33',
        }}>
          <div style={{
            width: 34, height: 34, borderRadius: 8,
            background: 'linear-gradient(135deg,#1d56d4,#2563eb)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            flexShrink: 0,
          }}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M1 3h15v13H1z"/><path d="M16 8h4l3 3v5h-7V8z"/><circle cx="5.5" cy="18.5" r="2.5"/><circle cx="18.5" cy="18.5" r="2.5"/>
            </svg>
          </div>
          <div>
            <div style={{ color: '#dce8f5', fontWeight: 700, fontSize: 14, lineHeight: 1.2 }}>KARVON</div>
            <div style={{ color: '#2d4a6e', fontSize: 11, fontWeight: 500, marginTop: 1 }}>
              {isAdmin ? 'Администрация' : 'Модерация'}
            </div>
          </div>
        </div>

        {/* Navigation */}
        <nav style={{ flex: 1, padding: '12px 8px', overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: 2 }}>
          {isAdmin && (
            <>
              <div style={{ fontSize: 10, fontWeight: 700, color: '#1e3350', letterSpacing: '.1em', textTransform: 'uppercase', padding: '8px 10px 4px' }}>
                Управление
              </div>
              {adminNav.map(({ to, label, Icon }) => (
                <NavLink key={to} to={to} className={({ isActive }) => `sidebar-link${isActive ? ' active' : ''}`}>
                  <Icon size={15} style={{ flexShrink: 0 }} />
                  {label}
                </NavLink>
              ))}
              <div style={{ height: 1, background: '#10192b', margin: '8px 0' }} />
              <div style={{ fontSize: 10, fontWeight: 700, color: '#1e3350', letterSpacing: '.1em', textTransform: 'uppercase', padding: '0 10px 4px' }}>
                Модерация
              </div>
            </>
          )}
          {modNav.map(({ to, label, Icon }) => (
            <NavLink key={to} to={to} className={({ isActive }) => `sidebar-link${isActive ? ' active' : ''}`}>
              <Icon size={15} style={{ flexShrink: 0 }} />
              {label}
            </NavLink>
          ))}
        </nav>

        {/* User footer */}
        <div style={{ padding: '12px 10px', borderTop: '1px solid #141f33' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 9 }}>
            <div style={{
              width: 32, height: 32, borderRadius: 8, flexShrink: 0,
              background: '#1d3a5c',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              color: '#60a5fa', fontSize: 12, fontWeight: 700,
            }}>{initials}</div>
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ color: '#c4d8ef', fontSize: 13, fontWeight: 600, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{name || 'Admin'}</div>
              <div style={{ color: '#2d4a6e', fontSize: 11, marginTop: 1 }}>{role?.replace('_', ' ') || 'user'}</div>
            </div>
            <button
              onClick={handleLogout}
              title="Выйти"
              style={{ color: '#2d4a6e', background: 'none', border: 'none', cursor: 'pointer', padding: 4, borderRadius: 5, display: 'flex', transition: 'color .13s' }}
              onMouseEnter={e => e.currentTarget.style.color = '#e84040'}
              onMouseLeave={e => e.currentTarget.style.color = '#2d4a6e'}
            >
              <IcoLogout size={15} />
            </button>
          </div>
        </div>
      </aside>

      {/* ── Main content ── */}
      <main style={{ flex: 1, overflowY: 'auto', background: '#070c17' }}>
        <Outlet />
      </main>
    </div>
  )
}
