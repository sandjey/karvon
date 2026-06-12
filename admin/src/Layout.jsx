import { Outlet, NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from './AuthContext'
import {
  IcoDashboard, IcoUsers, IcoShield, IcoBuilding, IcoList,
  IcoCard, IcoTag, IcoInbox, IcoHistory, IcoLogout,
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

const TruckSvg = () => (
  <svg width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M1 3h15v13H1z"/><path d="M16 8h4l3 3v5h-7V8z"/>
    <circle cx="5.5" cy="18.5" r="2.5"/><circle cx="18.5" cy="18.5" r="2.5"/>
  </svg>
)

function SectionLabel({ children }) {
  return (
    <div style={{
      fontSize: 10.5, fontWeight: 700, color: '#94a3b8',
      letterSpacing: '.08em', textTransform: 'uppercase',
      padding: '14px 10px 5px',
    }}>{children}</div>
  )
}

export default function Layout() {
  const { name, role, logout, isAdmin } = useAuth()
  const navigate = useNavigate()

  const handleLogout = () => { logout(); navigate('/login') }
  const initials = (name || 'KA').slice(0, 2).toUpperCase()

  return (
    <div style={{ display: 'flex', height: '100vh', overflow: 'hidden', background: '#f1f5f9' }}>

      {/* ── Sidebar ── */}
      <aside style={{
        width: 232, flexShrink: 0,
        display: 'flex', flexDirection: 'column',
        background: '#ffffff',
        borderRight: '1px solid #e2e8f0',
      }}>

        {/* Logo */}
        <div style={{
          display: 'flex', alignItems: 'center', gap: 10,
          padding: '18px 14px 16px',
          borderBottom: '1px solid #f1f5f9',
        }}>
          <div style={{
            width: 38, height: 38, borderRadius: 10, flexShrink: 0,
            background: 'linear-gradient(135deg, #2563eb, #3b82f6)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            boxShadow: '0 2px 8px rgba(37,99,235,.25)',
          }}>
            <TruckSvg />
          </div>
          <div>
            <div style={{ color: '#0f172a', fontWeight: 800, fontSize: 15.5, letterSpacing: '.01em', lineHeight: 1.2 }}>
              KARVON
            </div>
            <div style={{ color: '#94a3b8', fontSize: 11, fontWeight: 500, marginTop: 2 }}>
              {isAdmin ? 'Admin Panel' : 'Moderation'}
            </div>
          </div>
        </div>

        {/* Navigation */}
        <nav style={{ flex: 1, padding: '6px 10px', overflowY: 'auto', display: 'flex', flexDirection: 'column' }}>
          {isAdmin && (
            <>
              <SectionLabel>Управление</SectionLabel>
              {adminNav.map(({ to, label, Icon }) => (
                <NavLink key={to} to={to} className={({ isActive }) => `sidebar-link${isActive ? ' active' : ''}`}>
                  <Icon size={15} />
                  {label}
                </NavLink>
              ))}
              <SectionLabel>Модерация</SectionLabel>
            </>
          )}
          {modNav.map(({ to, label, Icon }) => (
            <NavLink key={to} to={to} className={({ isActive }) => `sidebar-link${isActive ? ' active' : ''}`}>
              <Icon size={15} />
              {label}
            </NavLink>
          ))}
        </nav>

        {/* User footer */}
        <div style={{ padding: '10px 12px', borderTop: '1px solid #f1f5f9' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 9 }}>
            <div style={{
              width: 32, height: 32, borderRadius: 8, flexShrink: 0,
              background: '#eff6ff',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              color: '#2563eb', fontSize: 12, fontWeight: 700,
            }}>{initials}</div>
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ color: '#0f172a', fontSize: 13, fontWeight: 600, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{name || 'Admin'}</div>
              <div style={{ color: '#94a3b8', fontSize: 11, marginTop: 1 }}>{role?.replace('_', ' ')}</div>
            </div>
            <button
              onClick={handleLogout}
              title="Выйти"
              style={{ color: '#94a3b8', background: 'none', border: 'none', cursor: 'pointer', padding: 5, borderRadius: 6, display: 'flex', transition: 'all .12s' }}
              onMouseEnter={e => { e.currentTarget.style.background = '#fee2e2'; e.currentTarget.style.color = '#dc2626' }}
              onMouseLeave={e => { e.currentTarget.style.background = 'none'; e.currentTarget.style.color = '#94a3b8' }}
            >
              <IcoLogout size={15} />
            </button>
          </div>
        </div>
      </aside>

      {/* ── Content ── */}
      <main style={{ flex: 1, overflowY: 'auto', background: '#f1f5f9' }}>
        <Outlet />
      </main>
    </div>
  )
}
