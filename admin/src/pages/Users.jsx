import { useState, useEffect, useCallback } from 'react'
import { api } from '../api'
import Pagination from '../components/Pagination'
import Modal from '../components/Modal'
import { statusBadge } from '../components/Badge'
import { IcoSearch, IcoCoins } from '../icons'

const lbl = { display: 'block', fontSize: 12, fontWeight: 600, color: '#374151', marginBottom: 8 }

export default function Users() {
  const [users, setUsers]     = useState([])
  const [total, setTotal]     = useState(0)
  const [page, setPage]       = useState(1)
  const [q, setQ]             = useState('')
  const [role, setRole]       = useState('')
  const [loading, setLoading] = useState(true)
  const [search, setSearch]   = useState('')
  const [topupModal, setTopupModal]   = useState(null)
  const [topupAmount, setTopupAmount] = useState('')

  const load = useCallback(() => {
    setLoading(true)
    api.users(q, role, page)
      .then(r => { setUsers(r?.data || []); setTotal(r?.meta?.total || 0) })
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [q, role, page])

  useEffect(() => { load() }, [load])

  const handleSearch = (e) => { e.preventDefault(); setQ(search); setPage(1) }

  const handleBlock = async (user) => {
    if (!confirm(`${user.is_blocked ? 'Разблокировать' : 'Заблокировать'} ${user.phone}?`)) return
    await api.blockUser(user.id, !user.is_blocked)
    load()
  }

  const handleTopup = async () => {
    const amount = parseInt(topupAmount)
    if (!amount || amount < 1) return
    await api.topupTokens(topupModal.id, amount)
    setTopupModal(null); setTopupAmount(''); load()
  }

  return (
    <div style={{ padding: 28, display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <div style={{ fontSize: 20, fontWeight: 800, color: '#0f172a' }}>Пользователи</div>
          <div style={{ fontSize: 13, color: '#64748b', marginTop: 3 }}>Всего: {total}</div>
        </div>
      </div>

      {/* Filters */}
      <div className="card" style={{ display: 'flex', gap: 10, flexWrap: 'wrap', padding: 14 }}>
        <form onSubmit={handleSearch} style={{ display: 'flex', gap: 8, flex: 1, minWidth: 240 }}>
          <div style={{ position: 'relative', flex: 1 }}>
            <div style={{ position: 'absolute', left: 10, top: '50%', transform: 'translateY(-50%)', color: '#94a3b8', pointerEvents: 'none' }}>
              <IcoSearch size={14} />
            </div>
            <input
              className="input" style={{ paddingLeft: 32 }}
              placeholder="Телефон или имя..."
              value={search}
              onChange={e => setSearch(e.target.value)}
            />
          </div>
          <button type="submit" className="btn-primary">Найти</button>
        </form>
        <select className="input" style={{ width: 180 }} value={role} onChange={e => { setRole(e.target.value); setPage(1) }}>
          <option value="">Все роли</option>
          <option value="user">Пользователь</option>
          <option value="moderator">Модератор</option>
          <option value="super_admin">Super Admin</option>
        </select>
      </div>

      {/* Table */}
      <div className="card" style={{ overflow: 'hidden', padding: 0 }}>
        {loading ? (
          <div style={{ textAlign: 'center', padding: '48px 0', color: '#94a3b8' }}>Загрузка...</div>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table className="data-table">
              <thead>
                <tr>
                  <th style={{ paddingLeft: 20 }}>Пользователь</th>
                  <th>Роль</th><th>Токены</th><th>Статус</th><th>Дата</th>
                  <th style={{ paddingRight: 20, textAlign: 'right' }}>Действия</th>
                </tr>
              </thead>
              <tbody>
                {users.length === 0 ? (
                  <tr><td colSpan={6} style={{ textAlign: 'center', padding: '40px 0', color: '#94a3b8' }}>Нет пользователей</td></tr>
                ) : users.map(u => (
                  <tr key={u.id}>
                    <td style={{ paddingLeft: 20 }}>
                      <div style={{ color: '#0f172a', fontWeight: 600 }}>{u.name || '—'}</div>
                      <div style={{ fontSize: 12, color: '#94a3b8', marginTop: 2, fontFamily: 'monospace' }}>{u.phone}</div>
                    </td>
                    <td>{statusBadge(u.role)}</td>
                    <td>
                      <span style={{ color: '#2563eb', fontWeight: 700 }}>{u.token_balance}</span>
                    </td>
                    <td>
                      {u.is_blocked
                        ? <span className="badge badge-red">Заблокирован</span>
                        : <span className="badge badge-green">Активен</span>
                      }
                    </td>
                    <td style={{ fontSize: 12 }}>{new Date(u.created_at).toLocaleDateString('ru')}</td>
                    <td style={{ paddingRight: 20 }}>
                      <div style={{ display: 'flex', gap: 6, justifyContent: 'flex-end' }}>
                        <button
                          onClick={() => setTopupModal(u)}
                          style={{
                            display: 'flex', alignItems: 'center', gap: 4,
                            padding: '4px 10px', fontSize: 12, fontWeight: 600,
                            background: '#eff6ff', color: '#2563eb',
                            border: '1px solid #bfdbfe', borderRadius: 6, cursor: 'pointer',
                          }}
                        ><IcoCoins size={12} /> Токены</button>
                        <button
                          onClick={() => handleBlock(u)}
                          style={{
                            padding: '4px 10px', fontSize: 12, fontWeight: 600,
                            background: u.is_blocked ? '#f0fdf4' : '#fef2f2',
                            color: u.is_blocked ? '#16a34a' : '#dc2626',
                            border: `1px solid ${u.is_blocked ? '#bbf7d0' : '#fecaca'}`,
                            borderRadius: 6, cursor: 'pointer',
                          }}
                        >{u.is_blocked ? 'Разблокировать' : 'Заблокировать'}</button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            <div style={{ padding: '0 20px 16px' }}>
              <Pagination page={page} total={total} onChange={setPage} />
            </div>
          </div>
        )}
      </div>

      {topupModal && (
        <Modal title={`Пополнить токены — ${topupModal.name || topupModal.phone}`} onClose={() => setTopupModal(null)}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <div>
              <label style={lbl}>Количество токенов</label>
              <input
                type="number" min="1" max="10000"
                className="input" placeholder="Например: 10"
                value={topupAmount}
                onChange={e => setTopupAmount(e.target.value)}
                autoFocus
              />
              <div style={{ fontSize: 12, color: '#64748b', marginTop: 6 }}>
                Баланс: <span style={{ color: '#2563eb', fontWeight: 600 }}>{topupModal.token_balance}</span> токенов
              </div>
            </div>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
              <button className="btn-secondary" onClick={() => setTopupModal(null)}>Отмена</button>
              <button className="btn-primary" onClick={handleTopup}>Пополнить</button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  )
}
