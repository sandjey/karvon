import { useState, useEffect, useCallback } from 'react'
import { api } from '../api'
import Pagination from '../components/Pagination'
import { statusBadge } from '../components/Badge'

export default function Listings() {
  const [items, setItems]     = useState([])
  const [total, setTotal]     = useState(0)
  const [page, setPage]       = useState(1)
  const [type, setType]       = useState('cargo')
  const [loading, setLoading] = useState(true)

  const load = useCallback(() => {
    setLoading(true)
    api.listings(type, page)
      .then(r => { setItems(r?.data?.items || []); setTotal(r?.data?.total || 0) })
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [type, page])

  useEffect(() => { load() }, [load])

  const handleDelete = async (item) => {
    if (!confirm(`Удалить объявление "${item.title || item.cargo_type}"?`)) return
    await api.deleteListing(type, item.id)
    load()
  }

  const handleBlock = async (item) => {
    if (!confirm(`Заблокировать объявление?`)) return
    await api.blockListing(type, item.id)
    load()
  }

  const isCargo = type === 'cargo'

  return (
    <div className="p-6 space-y-5">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-800">Объявления</h1>
          <p className="text-sm text-slate-500">Всего: {total}</p>
        </div>
        <div className="flex gap-1 bg-white rounded-xl border border-slate-200 p-1">
          <button
            onClick={() => { setType('cargo'); setPage(1) }}
            className={`px-5 py-1.5 rounded-lg text-sm font-medium transition ${type === 'cargo' ? 'bg-blue-600 text-white' : 'text-slate-600 hover:bg-slate-50'}`}
          >📦 Грузы</button>
          <button
            onClick={() => { setType('warehouse'); setPage(1) }}
            className={`px-5 py-1.5 rounded-lg text-sm font-medium transition ${type === 'warehouse' ? 'bg-blue-600 text-white' : 'text-slate-600 hover:bg-slate-50'}`}
          >🏭 Склады</button>
        </div>
      </div>

      <div className="card overflow-x-auto">
        {loading ? (
          <div className="text-center py-12 text-slate-400">Загрузка...</div>
        ) : (
          <>
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-slate-500 border-b border-slate-100">
                  {isCargo ? (
                    <>
                      <th className="pb-3 font-medium">Груз</th>
                      <th className="pb-3 font-medium">Маршрут</th>
                      <th className="pb-3 font-medium">Компания</th>
                      <th className="pb-3 font-medium">Цена</th>
                      <th className="pb-3 font-medium">Статус</th>
                    </>
                  ) : (
                    <>
                      <th className="pb-3 font-medium">Склад</th>
                      <th className="pb-3 font-medium">Тип</th>
                      <th className="pb-3 font-medium">Площадь</th>
                      <th className="pb-3 font-medium">Цена</th>
                      <th className="pb-3 font-medium">Статус</th>
                    </>
                  )}
                  <th className="pb-3 font-medium">Дата</th>
                  <th className="pb-3"></th>
                </tr>
              </thead>
              <tbody>
                {items.length === 0 ? (
                  <tr><td colSpan={7} className="text-center py-10 text-slate-400">Нет объявлений</td></tr>
                ) : items.map(item => (
                  <tr key={item.id} className="border-b border-slate-50 hover:bg-slate-50 group">
                    {isCargo ? (
                      <>
                        <td className="py-3">
                          <div className="font-medium">{item.cargo_type || item.category}</div>
                          <div className="text-xs text-slate-400">{item.weight_kg ? `${item.weight_kg} кг` : ''}</div>
                        </td>
                        <td className="py-3 text-slate-600">
                          {item.origin_city} → {item.destination_city}
                        </td>
                        <td className="py-3 text-slate-500">{item.company?.name || '—'}</td>
                        <td className="py-3">
                          {item.price ? (
                            <span className="font-medium">{item.currency === 'USD' ? '$' : ''}{item.price.toLocaleString()}{item.currency === 'UZS' ? ' сум' : ''}</span>
                          ) : <span className="text-slate-400">Договорная</span>}
                        </td>
                      </>
                    ) : (
                      <>
                        <td className="py-3">
                          <div className="font-medium">{item.title || item.name || '—'}</div>
                          <div className="text-xs text-slate-400">{item.address}</div>
                        </td>
                        <td className="py-3 text-slate-600 capitalize">{item.warehouse_type}</td>
                        <td className="py-3">{item.total_area_m2} м²</td>
                        <td className="py-3">
                          {item.price_per_m2 ? `$${item.price_per_m2}/м²` : '—'}
                        </td>
                      </>
                    )}
                    <td className="py-3">{statusBadge(item.status || (item.is_active ? 'active' : 'archived'))}</td>
                    <td className="py-3 text-slate-400 text-xs">{new Date(item.created_at).toLocaleDateString('ru')}</td>
                    <td className="py-3">
                      <div className="flex gap-2 opacity-0 group-hover:opacity-100 transition">
                        <button onClick={() => handleBlock(item)} className="px-2 py-1 bg-orange-50 text-orange-600 hover:bg-orange-100 rounded text-xs">Блок</button>
                        <button onClick={() => handleDelete(item)} className="px-2 py-1 bg-red-50 text-red-600 hover:bg-red-100 rounded text-xs">Удалить</button>
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
    </div>
  )
}
