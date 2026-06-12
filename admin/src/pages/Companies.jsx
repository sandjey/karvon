import { useState, useEffect, useCallback } from 'react'
import { api } from '../api'
import Pagination from '../components/Pagination'
import { statusBadge } from '../components/Badge'

const STATUSES = [
  { value: '',               label: 'Все' },
  { value: 'pending',        label: 'На проверке' },
  { value: 'approved',       label: 'Одобрены' },
  { value: 'rejected',       label: 'Отклонены' },
  { value: 'docs_requested', label: 'Доп. документы' },
]

export default function Companies() {
  const [items, setItems]       = useState([])
  const [total, setTotal]       = useState(0)
  const [page, setPage]         = useState(1)
  const [status, setStatus]     = useState('')
  const [loading, setLoading]   = useState(true)
  const [expanded, setExpanded] = useState(null)

  const load = useCallback(() => {
    setLoading(true)
    api.companies(status, page)
      .then(r => { setItems(r?.data || []); setTotal(r?.meta?.total || 0) })
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [status, page])

  useEffect(() => { load() }, [load])

  const tabBtn = (v) => ({
    padding: '7px 14px', borderRadius: 8, fontSize: 13, fontWeight: 500,
    cursor: 'pointer', border: 'none', fontFamily: 'inherit', transition: 'all .12s',
    background: status === v ? '#2563eb' : 'transparent',
    color: status === v ? '#fff' : '#64748b',
  })

  return (
    <div style={{ padding: 28, display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div>
        <div style={{ fontSize: 20, fontWeight: 800, color: '#0f172a' }}>Компании</div>
        <div style={{ fontSize: 13, color: '#64748b', marginTop: 3 }}>Всего: {total}</div>
      </div>

      <div style={{ display: 'flex', gap: 2, background: '#fff', border: '1px solid #e2e8f0', borderRadius: 10, padding: 3, width: 'fit-content' }}>
        {STATUSES.map(s => (
          <button key={s.value} style={tabBtn(s.value)} onClick={() => { setStatus(s.value); setPage(1) }}>
            {s.label}
          </button>
        ))}
      </div>

      <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
        {loading ? (
          <div style={{ textAlign: 'center', padding: '48px 0', color: '#94a3b8' }}>Загрузка...</div>
        ) : (
          <>
            <table className="data-table">
              <thead>
                <tr>
                  <th style={{ paddingLeft: 20 }}>Компания</th>
                  <th>Владелец</th>
                  <th>ИНН</th>
                  <th>Статус</th>
                  <th>Дата</th>
                </tr>
              </thead>
              <tbody>
                {items.length === 0 ? (
                  <tr><td colSpan={5} style={{ textAlign: 'center', padding: '40px 0', color: '#94a3b8' }}>Нет компаний</td></tr>
                ) : items.map(c => (
                  <>
                    <tr
                      key={c.id}
                      style={{ cursor: 'pointer' }}
                      onClick={() => setExpanded(expanded === c.id ? null : c.id)}
                    >
                      <td style={{ paddingLeft: 20 }}>
                        <div style={{ fontWeight: 600, color: '#0f172a' }}>{c.name}</div>
                        <div style={{ fontSize: 12, color: '#94a3b8', marginTop: 2 }}>{c.org_type?.toUpperCase()} · {c.country}</div>
                      </td>
                      <td>
                        <div style={{ color: '#334155' }}>{c.user?.name || '—'}</div>
                        <div style={{ fontSize: 12, color: '#94a3b8', fontFamily: 'monospace' }}>{c.user?.phone}</div>
                      </td>
                      <td style={{ color: '#64748b', fontFamily: 'monospace' }}>{c.inn}</td>
                      <td>{statusBadge(c.status)}</td>
                      <td style={{ fontSize: 12, color: '#94a3b8' }}>{new Date(c.created_at).toLocaleDateString('ru')}</td>
                    </tr>
                    {expanded === c.id && (
                      <tr key={`${c.id}-exp`}>
                        <td colSpan={5} style={{ background: '#f8fafc', borderBottom: '1px solid #f1f5f9' }}>
                          <div style={{ padding: '12px 20px', display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 10, fontSize: 13 }}>
                            <div><span style={{ color: '#94a3b8' }}>Город: </span>{c.city}{c.region ? `, ${c.region}` : ''}</div>
                            <div><span style={{ color: '#94a3b8' }}>Адрес: </span>{c.street || '—'}</div>
                            <div><span style={{ color: '#94a3b8' }}>Email: </span>{c.email || '—'}</div>
                            <div><span style={{ color: '#94a3b8' }}>Телефон: </span>{c.phone || '—'}</div>
                            {c.rejection_reason && (
                              <div style={{ gridColumn: '1/-1', color: '#dc2626' }}>
                                <span style={{ fontWeight: 600 }}>Причина отказа: </span>{c.rejection_reason}
                              </div>
                            )}
                            {c.docs_request_note && (
                              <div style={{ gridColumn: '1/-1', color: '#2563eb' }}>
                                <span style={{ fontWeight: 600 }}>Запрос документов: </span>{c.docs_request_note}
                              </div>
                            )}
                            {c.reg_doc_url && (
                              <div style={{ gridColumn: '1/-1' }}>
                                <a href={c.reg_doc_url} target="_blank" rel="noreferrer" style={{ color: '#2563eb', textDecoration: 'none', fontWeight: 500 }}>
                                  Регистрационный документ →
                                </a>
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
            <div style={{ padding: '0 20px 16px' }}>
              <Pagination page={page} total={total} onChange={setPage} />
            </div>
          </>
        )}
      </div>
    </div>
  )
}
