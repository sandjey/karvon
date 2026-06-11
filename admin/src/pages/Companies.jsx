import { useState, useEffect, useCallback } from 'react'
import { api } from '../api'
import Pagination from '../components/Pagination'
import { statusBadge } from '../components/Badge'

const STATUSES = [
  { value: '',                label: 'Все' },
  { value: 'pending',         label: '⏳ На проверке' },
  { value: 'approved',        label: '✅ Одобрены' },
  { value: 'rejected',        label: '❌ Отклонены' },
  { value: 'docs_requested',  label: '📄 Доп. документы' },
]

export default function Companies() {
  const [items, setItems]     = useState([])
  const [total, setTotal]     = useState(0)
  const [page, setPage]       = useState(1)
  const [status, setStatus]   = useState('')
  const [loading, setLoading] = useState(true)
  const [expanded, setExpanded] = useState(null)

  const load = useCallback(() => {
    setLoading(true)
    api.companies(status, page)
      .then(r => { setItems(r?.data?.items || []); setTotal(r?.data?.total || 0) })
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [status, page])

  useEffect(() => { load() }, [load])

  return (
    <div className="p-6 space-y-5">
      <div>
        <h1 className="text-xl font-bold text-slate-800">Компании</h1>
        <p className="text-sm text-slate-500">Всего: {total}</p>
      </div>

      {/* Status tabs */}
      <div className="flex gap-1 bg-white rounded-xl border border-slate-200 p-1 w-fit">
        {STATUSES.map(s => (
          <button
            key={s.value}
            onClick={() => { setStatus(s.value); setPage(1) }}
            className={`px-4 py-1.5 rounded-lg text-sm font-medium transition ${status === s.value ? 'bg-blue-600 text-white' : 'text-slate-600 hover:bg-slate-50'}`}
          >{s.label}</button>
        ))}
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
                  <th className="pb-3 font-medium">Компания</th>
                  <th className="pb-3 font-medium">Владелец</th>
                  <th className="pb-3 font-medium">ИНН</th>
                  <th className="pb-3 font-medium">Статус</th>
                  <th className="pb-3 font-medium">Дата</th>
                </tr>
              </thead>
              <tbody>
                {items.length === 0 ? (
                  <tr><td colSpan={5} className="text-center py-10 text-slate-400">Нет компаний</td></tr>
                ) : items.map(c => (
                  <>
                    <tr
                      key={c.id}
                      className="border-b border-slate-50 hover:bg-slate-50 cursor-pointer"
                      onClick={() => setExpanded(expanded === c.id ? null : c.id)}
                    >
                      <td className="py-3">
                        <div className="font-medium text-slate-800">{c.name}</div>
                        <div className="text-xs text-slate-400">{c.org_type?.toUpperCase()} · {c.country}</div>
                      </td>
                      <td className="py-3">
                        <div className="text-slate-700">{c.user?.name || '—'}</div>
                        <div className="text-xs text-slate-400">{c.user?.phone}</div>
                      </td>
                      <td className="py-3 text-slate-500">{c.inn}</td>
                      <td className="py-3">{statusBadge(c.status)}</td>
                      <td className="py-3 text-slate-400 text-xs">{new Date(c.created_at).toLocaleDateString('ru')}</td>
                    </tr>
                    {expanded === c.id && (
                      <tr key={`${c.id}-exp`} className="bg-slate-50">
                        <td colSpan={5} className="px-4 pb-4">
                          <div className="grid grid-cols-2 gap-4 pt-3 text-sm">
                            <div><span className="text-slate-500">Город:</span> {c.city}, {c.region}</div>
                            <div><span className="text-slate-500">Адрес:</span> {c.street || '—'}</div>
                            <div><span className="text-slate-500">Email:</span> {c.email || '—'}</div>
                            <div><span className="text-slate-500">Телефон:</span> {c.phone || '—'}</div>
                            {c.rejection_reason && <div className="col-span-2 text-red-600"><span className="font-medium">Причина отказа:</span> {c.rejection_reason}</div>}
                            {c.docs_request_note && <div className="col-span-2 text-blue-600"><span className="font-medium">Запрос документов:</span> {c.docs_request_note}</div>}
                            {c.reg_doc_url && (
                              <div className="col-span-2">
                                <a href={c.reg_doc_url} target="_blank" rel="noreferrer" className="text-blue-600 hover:underline text-sm">📄 Регистрационный документ</a>
                              </div>
                            )}
                          </div>
                        </td>
                      </tr>
                    )}
                  </>
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
