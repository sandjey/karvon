import { useState, useEffect } from 'react'
import { api } from '../api'

function StatCard({ label, value, sub, color = 'blue', icon }) {
  const colors = {
    blue:   'bg-blue-50 text-blue-600',
    green:  'bg-emerald-50 text-emerald-600',
    orange: 'bg-orange-50 text-orange-600',
    purple: 'bg-purple-50 text-purple-600',
  }
  return (
    <div className="card flex items-start gap-4">
      <div className={`w-12 h-12 rounded-xl flex items-center justify-center text-xl flex-shrink-0 ${colors[color]}`}>
        {icon}
      </div>
      <div>
        <div className="text-2xl font-bold text-slate-800">{value ?? '—'}</div>
        <div className="text-sm font-medium text-slate-600 mt-0.5">{label}</div>
        {sub && <div className="text-xs text-slate-400 mt-1">{sub}</div>}
      </div>
    </div>
  )
}

export default function Dashboard() {
  const [data, setData]     = useState(null)
  const [period, setPeriod] = useState('30d')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    api.dashboard(period)
      .then(r => setData(r?.data))
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [period])

  const d = data || {}

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-800">Дашборд</h1>
          <p className="text-sm text-slate-500">Общая статистика платформы</p>
        </div>
        <div className="flex gap-1 bg-white rounded-lg border border-slate-200 p-1">
          {['7d','30d','90d'].map(p => (
            <button
              key={p}
              onClick={() => setPeriod(p)}
              className={`px-4 py-1.5 rounded-md text-sm font-medium transition ${period === p ? 'bg-blue-600 text-white' : 'text-slate-600 hover:bg-slate-50'}`}
            >{p}</button>
          ))}
        </div>
      </div>

      {loading ? (
        <div className="text-center py-20 text-slate-400">Загрузка...</div>
      ) : (
        <>
          {/* Stats Grid */}
          <div className="grid grid-cols-2 xl:grid-cols-4 gap-4">
            <StatCard icon="👥" label="Пользователей" value={d.total_users?.toLocaleString()} sub={`+${d.new_users || 0} за период`} color="blue" />
            <StatCard icon="📋" label="Объявлений" value={d.total_listings?.toLocaleString()} sub={`Активных: ${d.active_listings || 0}`} color="green" />
            <StatCard icon="🏢" label="Компаний" value={d.total_companies?.toLocaleString()} sub={`Верифицированных: ${d.verified_companies || 0}`} color="orange" />
            <StatCard icon="💳" label="Платежей (UZS)" value={d.total_revenue_uzs ? (d.total_revenue_uzs / 1000000).toFixed(1) + ' млн' : '0'} sub={`Всего: ${d.total_payments || 0}`} color="purple" />
          </div>

          {/* Second row */}
          <div className="grid grid-cols-2 xl:grid-cols-4 gap-4">
            <StatCard icon="📥" label="Ожидают модерации" value={d.pending_companies} color="orange" />
            <StatCard icon="🪙" label="Токенов выдано" value={d.tokens_issued?.toLocaleString()} color="blue" />
            <StatCard icon="📦" label="Грузов" value={d.cargo_count?.toLocaleString()} color="green" />
            <StatCard icon="🏭" label="Складов" value={d.warehouse_count?.toLocaleString()} color="purple" />
          </div>

          {/* Recent activity placeholder */}
          <div className="card">
            <h2 className="font-semibold text-slate-700 mb-4">Последние регистрации</h2>
            {d.recent_users?.length ? (
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-left text-slate-500 border-b border-slate-100">
                    <th className="pb-2 font-medium">Имя</th>
                    <th className="pb-2 font-medium">Телефон</th>
                    <th className="pb-2 font-medium">Роль</th>
                    <th className="pb-2 font-medium">Токены</th>
                    <th className="pb-2 font-medium">Дата</th>
                  </tr>
                </thead>
                <tbody>
                  {d.recent_users.map(u => (
                    <tr key={u.id} className="border-b border-slate-50 hover:bg-slate-50">
                      <td className="py-2.5">{u.name || '—'}</td>
                      <td className="py-2.5 text-slate-500">{u.phone}</td>
                      <td className="py-2.5">{u.role}</td>
                      <td className="py-2.5">{u.token_balance}</td>
                      <td className="py-2.5 text-slate-400">{new Date(u.created_at).toLocaleDateString('ru')}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <div className="text-center py-8 text-slate-400 text-sm">Нет данных</div>
            )}
          </div>
        </>
      )}
    </div>
  )
}
