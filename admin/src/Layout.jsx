import { Outlet, NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from './AuthContext'

const adminNav = [
  { to: '/dashboard',  label: 'Дашборд',     icon: '📊' },
  { to: '/users',      label: 'Пользователи', icon: '👥' },
  { to: '/moderators', label: 'Модераторы',   icon: '🛡️' },
  { to: '/companies',  label: 'Компании',     icon: '🏢' },
  { to: '/listings',   label: 'Объявления',   icon: '📋' },
  { to: '/payments',   label: 'Платежи',      icon: '💳' },
  { to: '/pricing',    label: 'Тарифы',       icon: '💰' },
]

const modNav = [
  { to: '/queue',   label: 'Очередь',  icon: '📥' },
  { to: '/history', label: 'История',  icon: '📜' },
]

export default function Layout() {
  const { name, role, logout, isAdmin } = useAuth()
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const navItems = isAdmin ? [...adminNav, ...modNav] : modNav

  return (
    <div className="flex h-screen overflow-hidden">
      {/* Sidebar */}
      <aside className="w-60 bg-navy-800 flex flex-col flex-shrink-0" style={{background:'#122035'}}>
        {/* Logo */}
        <div className="flex items-center gap-2.5 px-5 py-5 border-b border-white/10">
          <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center text-white font-bold text-sm">CL</div>
          <div>
            <div className="text-white font-bold text-sm leading-tight">CargoLink</div>
            <div className="text-slate-400 text-xs">{isAdmin ? 'Super Admin' : 'Moderator'}</div>
          </div>
        </div>

        {/* Nav */}
        <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
          {isAdmin && (
            <div className="text-slate-500 text-xs font-semibold uppercase tracking-wider px-2 pb-2">Администрация</div>
          )}
          {(isAdmin ? adminNav : []).map(item => (
            <NavLink key={item.to} to={item.to} className={({isActive}) => `sidebar-link ${isActive ? 'active' : ''}`}>
              <span>{item.icon}</span> {item.label}
            </NavLink>
          ))}

          {isAdmin && (
            <div className="text-slate-500 text-xs font-semibold uppercase tracking-wider px-2 pt-4 pb-2">Модерация</div>
          )}
          {modNav.map(item => (
            <NavLink key={item.to} to={item.to} className={({isActive}) => `sidebar-link ${isActive ? 'active' : ''}`}>
              <span>{item.icon}</span> {item.label}
            </NavLink>
          ))}
        </nav>

        {/* User */}
        <div className="px-4 py-4 border-t border-white/10">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center text-white text-xs font-bold">
              {name.slice(0,2).toUpperCase()}
            </div>
            <div className="flex-1 min-w-0">
              <div className="text-white text-sm font-medium truncate">{name}</div>
              <div className="text-slate-400 text-xs capitalize">{role?.replace('_',' ')}</div>
            </div>
            <button onClick={handleLogout} className="text-slate-400 hover:text-red-400 transition text-lg" title="Выйти">⏻</button>
          </div>
        </div>
      </aside>

      {/* Main */}
      <main className="flex-1 overflow-y-auto bg-slate-100">
        <Outlet />
      </main>
    </div>
  )
}
