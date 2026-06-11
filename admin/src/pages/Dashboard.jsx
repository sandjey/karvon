import { useState, useEffect } from 'react'
import { api } from '../api'
import { IcoUsers, IcoList, IcoBuilding, IcoCard, IcoInbox, IcoCoins, IcoPackage, IcoWarehouse } from '../icons'

function StatCard({ label, value, sub, icon: Ico, accent = '#60a5fa' }) {
  return (
    <div className="card" style={{ display: 'flex', alignItems: 'flex-start', gap: 14 }}>
      <div style={{
        width: 44, height: 44, borderRadius: 10, flexShrink: 0,
        background: `${accent}18`,
        border: `1px solid ${accent}28`,
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        color: accent,
      }}>
        <Ico size={18} />
      </div>
      <div>
        <div style={{ fontSize: 24, fontWeight: 800, color: '#dce8f5', lineHeight: 1.1 }}>{value ?? '—'}</div>
        <div style={{ fontSize: 13, color: '#4d6d90', marginTop: 3 }}>{label}</div>
        {sub && <div style={{ fontSize: 11.5, color: '#2d4a6e', marginTop: 2 }}>{sub}</div>}
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

  const periodBtn = (p) => ({
    padding: '5px 14px', borderRadius: 6, fontSize: 13, fontWeight: 500, cursor: 'pointer',
    border: 'none', transition: 'all .13s', fontFamily: 'inherit',
    background: period === p ? '#1d56d4' : 'transparent',
    color: period === p ? '#fff' : '#334d6e',
  })

  return (
    <div style={{ padding: 28, display: 'flex', flexDirection: 'column', gap: 24 }}>

      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <div style={{ fontSize: 20, fontWeight: 800, color: '#dce8f5' }}>Дашборд</div>
          <div style={{ fontSize: 13, color: '#2d4a6e', marginTop: 3 }}>Статистика платформы</div>
        </div>
        <div style={{
          display: 'flex', gap: 2, background: '#0c1526',
          border: '1px solid #19273d', borderRadius: 8, padding: 3,
        }}>
          {['7d', '30d', '90d'].map(p => (
            <button key={p} style={periodBtn(p)} onClick={() => setPeriod(p)}>{p}</button>
          ))}
        </div>
      </div>

      {loading ? (
        <div style={{ textAlign: 'center', padding: '60px 0', color: '#1e3350' }}>Загрузка данных...</div>
      ) : (
        <>
          {/* Main stats */}
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 14 }}>
            <StatCard icon={IcoUsers}    label="Пользователей"  value={d.total_users?.toLocaleString()}  sub={`+${d.new_users||0} за период`} accent="#60a5fa" />
            <StatCard icon={IcoList}     label="Объявлений"     value={d.total_listings?.toLocaleString()} sub={`Активных: ${d.active_listings||0}`} accent="#34d399" />
            <StatCard icon={IcoBuilding} label="Компаний"       value={d.total_companies?.toLocaleString()} sub={`Верифицировано: ${d.verified_companies||0}`} accent="#f59e0b" />
            <StatCard icon={IcoCard}     label="Выручка (млн ₽)"  value={d.total_revenue_uzs ? (d.total_revenue_uzs/1000000).toFixed(1)+' млн' : '0'} sub={`Транзакций: ${d.total_payments||0}`} accent="#a78bfa" />
          </div>

          {/* Secondary stats */}
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 14 }}>
            <StatCard icon={IcoInbox}    label="Ожидают модерации"  value={d.pending_companies}           accent="#f59e0b" />
            <StatCard icon={IcoCoins}    label="Токенов выдано"     value={d.tokens_issued?.toLocaleString()} accent="#60a5fa" />
            <StatCard icon={IcoPackage}  label="Грузов"             value={d.cargo_count?.toLocaleString()} accent="#34d399" />
            <StatCard icon={IcoWarehouse} label="Складов"           value={d.warehouse_count?.toLocaleString()} accent="#a78bfa" />
          </div>

          {/* Recent users */}
          <div className="card">
            <div style={{ fontSize: 14, fontWeight: 700, color: '#dce8f5', marginBottom: 16 }}>
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
                      <td style={{ color: '#c4d8ef' }}>{u.name || '—'}</td>
                      <td>{u.phone}</td>
                      <td>{u.role}</td>
                      <td style={{ color: '#60a5fa', fontWeight: 600 }}>{u.token_balance}</td>
                      <td style={{ fontSize: 12 }}>{new Date(u.created_at).toLocaleDateString('ru')}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <div style={{ textAlign: 'center', padding: '28px 0', color: '#1e3350', fontSize: 13 }}>Нет данных</div>
            )}
          </div>
        </>
      )}
    </div>
  )
}
