import { useState, useEffect, useCallback } from 'react'
import { api } from '../api'
import Pagination from '../components/Pagination'
import { statusBadge } from '../components/Badge'

export default function History() {
  const [items, setItems]     = useState([])
  const [total, setTotal]     = useState(0)
  const [page, setPage]       = useState(1)
  const [loading, setLoading] = useState(true)

  const load = useCallback(() => {
    setLoading(true)
    api.history(page)
      .then(r => { setItems(r?.data?.items || []); setTotal(r?.data?.total || 0) })
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [page])

  useEffect(() => { load() }, [load])

  return (
    <div className="p-6 space-y-5">
      <div>
        <h1 className="text-xl font-bold text-slate-800">История модерации</h1>
        <p className="text-sm text-slate-500">Ваши прошлые решения по компаниям</p>
      </div>

      <div className="card overflow-x-auto">
        {loading ? (
          <div className="text-center py-12 text-slate-400">Загрузка...</div>
        ) : items.length === 0 ? (
          <div className="text-center py-16">
            <div className="text-4xl mb-3">📜</div>
            <div className="text-slate-500 font-medium">История пуста</div>
            <p className="text-slate-400 text-sm mt-1">Здесь будут отображаться ваши решения</p>
          </div>
        ) : (
          <>
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-slate-500 border-b border-slate-100">
                  <th className="pb-3 font-medium">Компания</th>
                  <th className="pb-3 font-medium">Владелец</th>
                  <th className="pb-3 font-medium">Страна</th>
                  <th className="pb-3 font-medium">Статус</th>
                  <th className="pb-3 font-medium">Дата</th>
                  <th className="pb-3 font-medium">Примечание</th>
                </tr>
              </thead>
              <tbody>
                {items.map(item => (
                  <tr key={item.id} className="border-b border-slate-50 hover:bg-slate-50">
                    <td className="py-3">
                      <div className="font-medium text-slate-800">{item.company_name}</div>
                      <div className="text-xs text-slate-400">{item.org_type?.toUpperCase()} · {item.inn}</div>
                    </td>
                    <td className="py-3">
                      <div className="text-slate-700">{item.user_name || '—'}</div>
                      <div className="text-xs text-slate-400">{item.user_phone}</div>
                    </td>
                    <td className="py-3 text-slate-500">{item.country}</td>
                    <td className="py-3">{statusBadge(item.status)}</td>
                    <td className="py-3 text-slate-400 text-xs">{new Date(item.created_at).toLocaleDateString('ru', { day:'numeric', month:'short', year:'numeric' })}</td>
                    <td className="py-3 text-slate-500 text-xs max-w-xs">
                      {item.rejection_reason || item.docs_request_note || '—'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            <Pagination page={page} total={total} onChange={setPage} />
          </>
        )}
      </div>
    </div>
  )
}
