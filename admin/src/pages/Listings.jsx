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
      .then(r => { setItems(r?.data || []); setTotal(r?.meta?.total || 0) })
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [type, page])

  useEffect(() => { load() }, [load])

  const handleDelete = async (item) => {
    if (!confirm(`Удалить объявление "${item.title || item.cargo_type}"?`)) return
    await api.deleteListing(type, item.id); load()
  }

  const handleBlock = async (item) => {
    if (!confirm('Заблокировать объявление?')) return
    await api.blockListing(type, item.id); load()
  }

  const isCargo = type === 'cargo'

  const tabBtn = (v) => ({
    padding: '7px 16px', borderRadius: 8, fontSize: 13, fontWeight: 500,
    cursor: 'pointer', border: 'none', fontFamily: 'inherit', transition: 'all .12s',
    background: type === v ? '#2563eb' : 'transparent',
    color: type === v ? '#fff' : '#64748b',
  })

  return (
    <div style={{ padding: 28, display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <div style={{ fontSize: 20, fontWeight: 800, color: '#0f172a' }}>Объявления</div>
          <div style={{ fontSize: 13, color: '#64748b', marginTop: 3 }}>Всего: {total}</div>
        </div>
        <div style={{ display: 'flex', gap: 2, background: '#fff', border: '1px solid #e2e8f0', borderRadius: 10, padding: 3 }}>
          <button style={tabBtn('cargo')} onClick={() => { setType('cargo'); setPage(1) }}>Грузы</button>
          <button style={tabBtn('warehouse')} onClick={() => { setType('warehouse'); setPage(1) }}>Склады</button>
        </div>
      </div>

      <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
        {loading ? (
          <div style={{ textAlign: 'center', padding: '48px 0', color: '#94a3b8' }}>Загрузка...</div>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table className="data-table">
              <thead>
                <tr>
                  {isCargo ? (
                    <>
                      <th style={{ paddingLeft: 20 }}>Груз</th>
                      <th>Маршрут</th>
                      <th>Компания</th>
                      <th>Цена</th>
                    </>
                  ) : (
                    <>
                      <th style={{ paddingLeft: 20 }}>Склад</th>
                      <th>Тип</th>
                      <th>Площадь</th>
                      <th>Цена</th>
                    </>
                  )}
                  <th>Статус</th>
                  <th>Дата</th>
                  <th style={{ paddingRight: 20 }}></th>
                </tr>
              </thead>
              <tbody>
                {items.length === 0 ? (
                  <tr><td colSpan={7} style={{ textAlign: 'center', padding: '40px 0', color: '#94a3b8' }}>Нет объявлений</td></tr>
                ) : items.map(item => (
                  <tr key={item.id}>
                    {isCargo ? (
                      <>
                        <td style={{ paddingLeft: 20 }}>
                          <div style={{ fontWeight: 600, color: '#0f172a' }}>{item.cargo_type || item.category || '—'}</div>
                          <div style={{ fontSize: 12, color: '#94a3b8' }}>{item.weight_kg ? `${item.weight_kg} кг` : ''}</div>
                        </td>
                        <td style={{ color: '#475569' }}>{item.origin_city} → {item.destination_city}</td>
                        <td style={{ color: '#64748b' }}>{item.company?.name || '—'}</td>
                        <td>
                          {item.price
                            ? <span style={{ fontWeight: 600 }}>{item.currency === 'USD' ? '$' : ''}{item.price?.toLocaleString()}{item.currency === 'UZS' ? ' сум' : ''}</span>
                            : <span style={{ color: '#94a3b8' }}>Договорная</span>
                          }
                        </td>
                      </>
                    ) : (
                      <>
                        <td style={{ paddingLeft: 20 }}>
                          <div style={{ fontWeight: 600, color: '#0f172a' }}>{item.title || item.name || '—'}</div>
                          <div style={{ fontSize: 12, color: '#94a3b8' }}>{item.address}</div>
                        </td>
                        <td style={{ color: '#475569', textTransform: 'capitalize' }}>{item.warehouse_type}</td>
                        <td>{item.total_area_m2} м²</td>
                        <td>{item.price_per_m2 ? `$${item.price_per_m2}/м²` : '—'}</td>
                      </>
                    )}
                    <td>{statusBadge(item.status || (item.is_active ? 'active' : 'archived'))}</td>
                    <td style={{ fontSize: 12, color: '#94a3b8' }}>{new Date(item.created_at).toLocaleDateString('ru')}</td>
                    <td style={{ paddingRight: 20 }}>
                      <div style={{ display: 'flex', gap: 6 }}>
                        <button
                          onClick={() => handleBlock(item)}
                          style={{ padding: '4px 8px', background: '#fff7ed', color: '#c2410c', border: '1px solid #fed7aa', borderRadius: 6, fontSize: 12, cursor: 'pointer' }}
                        >Блок</button>
                        <button
                          onClick={() => handleDelete(item)}
                          style={{ padding: '4px 8px', background: '#fef2f2', color: '#dc2626', border: '1px solid #fecaca', borderRadius: 6, fontSize: 12, cursor: 'pointer' }}
                        >Удалить</button>
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
    </div>
  )
}
