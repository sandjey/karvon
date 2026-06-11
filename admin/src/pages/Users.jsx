import { useState, useEffect, useCallback } from 'react'
import { api } from '../api'
import Pagination from '../components/Pagination'
import Modal from '../components/Modal'
import { statusBadge } from '../components/Badge'

export default function Users() {
  const [users, setUsers]     = useState([])
  const [total, setTotal]     = useState(0)
  const [page, setPage]       = useState(1)
  const [q, setQ]             = useState('')
  const [role, setRole]       = useState('')
  const [loading, setLoading] = useState(true)
  const [search, setSearch]   = useState('')

  const [topupModal, setTopupModal] = useState(null)
  const [topupAmount, setTopupAmount] = useState('')
  const [selectedUser, setSelectedUser] = useState(null)

  const load = useCallback(() => {
    setLoading(true)
    api.users(q, role, page)
      .then(r => { setUsers(r?.data?.items || []); setTotal(r?.data?.total || 0) })
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [q, role, page])

  useEffect(() => { load() }, [load])

  const handleSearch = (e) => {
    e.preventDefault()
    setQ(search)
    setPage(1)
  }

  const handleBlock = async (user) => {
    if (!confirm(`${user.is_blocked ? 'Разблокировать' : 'Заблокировать'} ${user.phone}?`)) return
    await api.blockUser(user.id, !user.is_blocked)
    load()
  }

  const handleTopup = async () => {
    const amount = parseInt(topupAmount)
    if (!amount || amount < 1) return
    await api.topupTokens(topupModal.id, amount)
    setTopupModal(null)
    setTopupAmount('')
    load()
  }

  return (
    <div className="p-6 space-y-5">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-800">Пользователи</h1>
          <p className="text-sm text-slate-500">Всего: {total}</p>
        </div>
      </div>

      {/* Filters */}
      <div className="card flex flex-wrap gap-3">
        <form onSubmit={handleSearch} className="flex gap-2 flex-1 min-w-60">
          <input
            className="input flex-1"
            placeholder="Поиск по телефону или имени..."
            value={search}
            onChange={e => setSearch(e.target.value)}
          />
          <button type="submit" className="btn-primary">Найти</button>
        </form>
        <select className="input w-44" value={role} onChange={e => { setRole(e.target.value); setPage(1) }}>
          <option value="">Все роли</option>
          <option value="user">Пользователь</option>
          <option value="moderator">Модератор</option>
          <option value="super_admin">Super Admin</option>
        </select>
      </div>

      {/* Table */}
      <div className="card overflow-x-auto">
        {loading ? (
          <div className="text-center py-12 text-slate-400">Загрузка...</div>
        ) : (
          <>
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-slate-500 border-b border-slate-100">
                  <th className="pb-3 font-medium">Пользователь</th>
                  <th className="pb-3 font-medium">Роль</th>
                  <th className="pb-3 font-medium">Токены</th>
                  <th className="pb-3 font-medium">Статус</th>
                  <th className="pb-3 font-medium">Дата</th>
                  <th className="pb-3 font-medium text-right">Действия</th>
                </tr>
              </thead>
              <tbody>
                {users.length === 0 ? (
                  <tr><td colSpan={6} className="text-center py-10 text-slate-400">Нет пользователей</td></tr>
                ) : users.map(u => (
                  <tr key={u.id} className="border-b border-slate-50 hover:bg-slate-50 group">
                    <td className="py-3">
                      <div className="font-medium text-slate-800">{u.name || '—'}</div>
                      <div className="text-slate-400 text-xs">{u.phone}</div>
                    </td>
                    <td className="py-3">{statusBadge(u.role)}</td>
                    <td className="py-3">
                      <span className="font-semibold text-blue-600">{u.token_balance}</span>
                    </td>
                    <td className="py-3">
                      {u.is_blocked
                        ? <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-700">Заблокирован</span>
                        : <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-emerald-100 text-emerald-700">Активен</span>
                      }
                    </td>
                    <td className="py-3 text-slate-400 text-xs">{new Date(u.created_at).toLocaleDateString('ru')}</td>
                    <td className="py-3">
                      <div className="flex gap-2 justify-end opacity-0 group-hover:opacity-100 transition">
                        <button
                          onClick={() => setTopupModal(u)}
                          className="px-3 py-1 bg-blue-50 text-blue-600 hover:bg-blue-100 rounded-lg text-xs font-medium"
                        >+ Токены</button>
                        <button
                          onClick={() => handleBlock(u)}
                          className={`px-3 py-1 rounded-lg text-xs font-medium ${u.is_blocked ? 'bg-emerald-50 text-emerald-600 hover:bg-emerald-100' : 'bg-red-50 text-red-600 hover:bg-red-100'}`}
                        >{u.is_blocked ? 'Разблокировать' : 'Заблокировать'}</button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            <Pagination page={page} total={total} onChange={setPage} />
          </>
        )}
      </div>

      {/* Topup Modal */}
      {topupModal && (
        <Modal title={`Пополнить токены — ${topupModal.name || topupModal.phone}`} onClose={() => setTopupModal(null)}>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">Количество токенов</label>
              <input
                type="number" min="1" max="10000"
                className="input"
                placeholder="Например: 10"
                value={topupAmount}
                onChange={e => setTopupAmount(e.target.value)}
                autoFocus
              />
              <p className="text-xs text-slate-400 mt-1">Текущий баланс: {topupModal.token_balance} токенов</p>
            </div>
            <div className="flex gap-3 justify-end">
              <button className="btn-secondary" onClick={() => setTopupModal(null)}>Отмена</button>
              <button className="btn-primary" onClick={handleTopup}>Пополнить</button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  )
}
