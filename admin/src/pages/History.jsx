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
      .then(r => { setItems(r?.data || []); setTotal(r?.meta?.total || 0) })
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [page])

  useEffect(() => { load() }, [load])

  return (
    <div style={{ padding: 28, display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div>
        <div style={{ fontSize: 20, fontWeight: 800, color: '#0f172a' }}>История модерации</div>
        <div style={{ fontSize: 13, color: '#64748b', marginTop: 3 }}>Ваши прошлые решения по компаниям</div>
      </div>

      <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
        {loading ? (
          <div style={{ textAlign: 'center', padding: '48px 0', color: '#94a3b8' }}>Загрузка...</div>
        ) : items.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '60px 0' }}>
            <div style={{ fontSize: 36, marginBottom: 12 }}>📜</div>
            <div style={{ color: '#64748b', fontWeight: 600 }}>История пуста</div>
            <div style={{ color: '#94a3b8', fontSize: 13, marginTop: 6 }}>Здесь будут ваши решения</div>
          </div>
        ) : (
          <>
            <table className="data-table">
              <thead>
                <tr>
                  <th style={{ paddingLeft: 20 }}>Компания</th>
                  <th>Владелец</th>
                  <th>Страна</th>
                  <th>Статус</th>
                  <th>Дата</th>
                  <th style={{ paddingRight: 20 }}>Примечание</th>
                </tr>
              </thead>
              <tbody>
                {items.map(item => (
                  <tr key={item.id}>
                    <td style={{ paddingLeft: 20 }}>
                      <div style={{ fontWeight: 600, color: '#0f172a' }}>{item.company_name}</div>
                      <div style={{ fontSize: 12, color: '#94a3b8' }}>{item.org_type?.toUpperCase()} · {item.inn}</div>
                    </td>
                    <td>
                      <div style={{ color: '#334155' }}>{item.user_name || '—'}</div>
                      <div style={{ fontSize: 12, color: '#94a3b8', fontFamily: 'monospace' }}>{item.user_phone}</div>
                    </td>
                    <td style={{ color: '#64748b' }}>{item.country}</td>
                    <td>{statusBadge(item.status)}</td>
                    <td style={{ fontSize: 12, color: '#94a3b8' }}>{new Date(item.created_at).toLocaleDateString('ru', { day: 'numeric', month: 'short', year: 'numeric' })}</td>
                    <td style={{ fontSize: 12, color: '#64748b', paddingRight: 20, maxWidth: 200 }}>
                      {item.rejection_reason || item.docs_request_note || '—'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            <div style={{ padding: '0 20px 16px' }}>
              <Pagination page={page} total={total} onChange={setPage} />
            </div>
          </>
        )}
      </div>
    </div>
  )
}
