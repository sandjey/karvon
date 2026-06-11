import { useState, useEffect, useCallback } from 'react'
import { api } from '../api'
import Pagination from '../components/Pagination'
import { statusBadge } from '../components/Badge'

const TYPE_LABELS = {
  tokens: 'Токены',
  subscription: 'Подписка',
  listing: 'Объявление',
  boost: 'Продвижение',
}

export default function Payments() {
  const [items, setItems]     = useState([])
  const [total, setTotal]     = useState(0)
  const [page, setPage]       = useState(1)
  const [loading, setLoading] = useState(true)

  const load = useCallback(() => {
    setLoading(true)
    api.payments(page)
      .then(r => { setItems(r?.data?.items || []); setTotal(r?.data?.total || 0) })
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [page])

  useEffect(() => { load() }, [load])

  const totalPaid = items.filter(p => p.status === 'paid').reduce((s, p) => s + (p.amount || 0), 0)

  return (
    <div className="p-6 space-y-5">
      <div>
        <h1 className="text-xl font-bold text-slate-800">Платежи</h1>
        <p className="text-sm text-slate-500">Всего: {total}</p>
      </div>

      {/* Summary */}
      <div className="grid grid-cols-3 gap-4">
        {[
          { label: 'На этой странице оплачено (UZS)', value: totalPaid.toLocaleString() },
          { label: 'Всего транзакций', value: total },
          { label: 'Успешных', value: items.filter(p => p.status === 'paid').length },
        ].map(s => (
          <div key={s.label} className="card">
            <div className="text-lg font-bold text-slate-800">{s.value}</div>
            <div className="text-xs text-slate-500 mt-0.5">{s.label}</div>
          </div>
        ))}
      </div>

      <div className="card overflow-x-auto">
        {loading ? (
          <div className="text-center py-12 text-slate-400">Загрузка...</div>
        ) : (
          <>
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-slate-500 border-b border-slate-100">
                  <th className="pb-3 font-medium">ID</th>
                  <th className="pb-3 font-medium">Тип</th>
                  <th className="pb-3 font-medium">Сумма</th>
                  <th className="pb-3 font-medium">Статус</th>
                  <th className="pb-3 font-medium">Способ</th>
                  <th className="pb-3 font-medium">Дата</th>
                  <th className="pb-3 font-medium">Оплачено</th>
                </tr>
              </thead>
              <tbody>
                {items.length === 0 ? (
                  <tr><td colSpan={7} className="text-center py-10 text-slate-400">Нет платежей</td></tr>
                ) : items.map(p => (
                  <tr key={p.id} className="border-b border-slate-50 hover:bg-slate-50">
                    <td className="py-3 font-mono text-xs text-slate-400">{p.id?.slice(0,8)}…</td>
                    <td className="py-3">
                      <span className="font-medium text-slate-700">{TYPE_LABELS[p.payment_type] || p.payment_type}</span>
                    </td>
                    <td className="py-3 font-semibold">
                      {p.currency === 'USD' ? '$' : ''}{p.amount?.toLocaleString()}{p.currency === 'UZS' ? ' сум' : ''}
                    </td>
                    <td className="py-3">{statusBadge(p.status)}</td>
                    <td className="py-3 text-slate-500 capitalize">{p.payment_method || '—'}</td>
                    <td className="py-3 text-slate-400 text-xs">{new Date(p.created_at).toLocaleDateString('ru')}</td>
                    <td className="py-3 text-slate-400 text-xs">
                      {p.paid_at ? new Date(p.paid_at).toLocaleDateString('ru') : '—'}
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
