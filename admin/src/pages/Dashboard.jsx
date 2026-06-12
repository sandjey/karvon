import { useState, useEffect } from 'react'
import { api } from '../api'
import { IcoUsers, IcoList, IcoBuilding, IcoCard, IcoInbox, IcoCoins, IcoPackage, IcoWarehouse } from '../icons'

function StatCard({ label, value, sub, icon: Ico, accent = '#2563eb' }) {
  return (
    <div className="card" style={{ display: 'flex', alignItems: 'flex-start', gap: 14 }}>
      <div style={{
        width: 46, height: 46, borderRadius: 12, flexShrink: 0,
        background: `${accent}14`,
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        color: accent,
      }}>
        <Ico size={20} />
      </div>
      <div>
        <div style={{ fontSize: 26, fontWeight: 800, color: '#0f172a', lineHeight: 1.1 }}>
          {value ?? <span style={{ color: '#cbd5e1' }}>—</span>}
        </div>
        <div style={{ fontSize: 13, color: '#64748b', marginTop: 4 }}>{label}</div>
        {sub && <div style={{ fontSize: 12, color: '#94a3b8', marginTop: 2 }}>{sub}</div>}
      </div>
    </div>
  )
}

export default function Dashboard() {
  const [data, setData]       = useState(null)
  const [period, setPeriod]   = useState('30d')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    api.dashboard(period)
      .then(r => setData(r?.data))
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [period])

  const d = data || {}

  const periodTab = (p) => ({
    padding: '6px 14px', borderRadius: 7, fontSize: 13, fontWeight: 500,
    cursor: 'pointer', border: 'none', fontFamily: 'inherit', transition: 'all .12s',
    background: period === p ? '#2563eb' : 'transparent',
    color: period === p ? '#fff' : '#64748b',
  })

  return (
    <div style={{ padding: 28, display: 'flex', flexDirection: 'column', gap: 22 }}>

      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <div style={{ fontSize: 20, fontWeight: 800, color: '#0f172a' }}>Дашборд</div>
          <div style={{ fontSize: 13, color: '#64748b', marginTop: 3 }}>Статистика платформы</div>
        </div>
        <div style={{
          display: 'flex', gap: 2, background: '#fff',
          border: '1px solid #e2e8f0', borderRadius: 10, padding: 3,
        }}>
          {['7d', '30d', '90d'].map(p => (
            <button key={p} style={periodTab(p)} onClick={() => setPeriod(p)}>{p}</button>
          ))}
        </div>
      </div>

      {loading ? (
        <div style={{ textAlign: 'center', padding: '60px 0', color: '#94a3b8', fontSize: 14 }}>
          Загрузка данных...
        </div>
      ) : (
        <>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 14 }}>
            <StatCard icon={IcoUsers}    label="Пользователей"     value={d.total_users?.toLocaleString()}      sub={`+${d.new_users || 0} за период`} accent="#2563eb" />
            <StatCard icon={IcoList}     label="Объявлений"        value={d.total_listings?.toLocaleString()}   sub={`Активных: ${d.active_listings || 0}`} accent="#059669" />
            <StatCard icon={IcoBuilding} label="Компаний"          value={d.total_companies?.toLocaleString()}  sub={`Верифицировано: ${d.verified_companies || 0}`} accent="#d97706" />
            <StatCard icon={IcoCard}     label="Выручка UZS"       value={d.total_revenue_uzs ? (d.total_revenue_uzs / 1000000).toFixed(1) + ' млн' : '0'} sub={`Транзакций: ${d.total_payments || 0}`} accent="#7c3aed" />
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 14 }}>
            <StatCard icon={IcoInbox}     label="Ждут модерации"  value={d.pending_companies}             accent="#d97706" />
            <StatCard icon={IcoCoins}     label="Токенов выдано"  value={d.tokens_issued?.toLocaleString()} accent="#2563eb" />
            <StatCard icon={IcoPackage}   label="Грузов"          value={d.cargo_count?.toLocaleString()}  accent="#059669" />
            <StatCard icon={IcoWarehouse} label="Складов"         value={d.warehouse_count?.toLocaleString()} accent="#7c3aed" />
          </div>

          <div className="card">
            <div style={{ fontSize: 14, fontWeight: 700, color: '#0f172a', marginBottom: 16 }}>
              Последние регистрации
            </div>
            {d.recent_users?.length ? (
              <table className="data-table">
                <thead>
                  <tr>
                    <th>Имя</th><th>Телефон</th><th>Роль</th><th>Токены</th><th>Дата</th>
                  </tr>
                </thead>
                <tbody>
                  {d.recent_users.map(u => (
                    <tr key={u.id}>
                      <td style={{ color: '#0f172a', fontWeight: 500 }}>{u.name || '—'}</td>
                      <td style={{ fontFamily: 'monospace' }}>{u.phone}</td>
                      <td>{u.role}</td>
                      <td style={{ color: '#2563eb', fontWeight: 600 }}>{u.token_balance}</td>
                      <td style={{ fontSize: 12 }}>{new Date(u.created_at).toLocaleDateString('ru')}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <div style={{ textAlign: 'center', padding: '28px 0', color: '#94a3b8', fontSize: 13 }}>Нет данных</div>
            )}
          </div>
        </>
      )}
    </div>
  )
}
