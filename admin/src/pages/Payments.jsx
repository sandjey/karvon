import { useState, useEffect, useCallback } from 'react'
import { api } from '../api'
import Pagination from '../components/Pagination'
import { statusBadge } from '../components/Badge'
import { IcoTrendUp, IcoCard } from '../icons'

const TYPE_LABELS = {
  tokens: 'Токены', subscription: 'Подписка', listing: 'Объявление', boost: 'Продвижение',
}

export default function Payments() {
  const [items, setItems]   = useState([])
  const [total, setTotal]   = useState(0)
  const [page, setPage]     = useState(1)
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
  const paidCount = items.filter(p => p.status === 'paid').length

  return (
    <div style={{ padding: 28, display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div>
        <div style={{ fontSize: 20, fontWeight: 800, color: '#dce8f5' }}>Платежи</div>
        <div style={{ fontSize: 13, color: '#2d4a6e', marginTop: 3 }}>Всего транзакций: {total}</div>
      </div>

      {/* Summary cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 14 }}>
        {[
          { label: 'Оплачено на странице (UZS)', value: totalPaid.toLocaleString(), icon: IcoCard, accent: '#60a5fa' },
          { label: 'Всего транзакций', value: total, icon: IcoTrendUp, accent: '#f59e0b' },
          { label: 'Успешных на странице', value: paidCount, icon: IcoCard, accent: '#12c678' },
        ].map((s, i) => (
          <div key={i} className="card" style={{ display: 'flex', alignItems: 'center', gap: 12, padding: 16 }}>
            <div style={{
              width: 40, height: 40, borderRadius: 9, flexShrink: 0,
              background: `${s.accent}18`, color: s.accent,
              border: `1px solid ${s.accent}28`,
              display: 'flex', alignItems: 'center', justifyContent: 'center',
            }}>
              <s.icon size={18} />
            </div>
            <div>
              <div style={{ fontSize: 20, fontWeight: 800, color: '#dce8f5' }}>{s.value}</div>
              <div style={{ fontSize: 12, color: '#2d4a6e', marginTop: 2 }}>{s.label}</div>
            </div>
          </div>
        ))}
      </div>

      <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
        {loading ? (
          <div style={{ textAlign: 'center', padding: '48px 0', color: '#1e3350' }}>Загрузка...</div>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table className="data-table">
              <thead>
                <tr>
                  <th style={{ paddingLeft: 20 }}>ID</th>
                  <th>Тип</th><th>Сумма</th><th>Статус</th>
                  <th>Способ</th><th>Дата</th><th>Оплачено</th>
                </tr>
              </thead>
              <tbody>
                {items.length === 0 ? (
                  <tr><td colSpan={7} style={{ textAlign: 'center', padding: '40px 0', color: '#1e3350' }}>Нет платежей</td></tr>
                ) : items.map(p => (
                  <tr key={p.id}>
                    <td style={{ paddingLeft: 20 }}>
                      <code style={{ fontSize: 11, color: '#2d4a6e' }}>{p.id?.slice(0, 8)}…</code>
                    </td>
                    <td style={{ color: '#c4d8ef', fontWeight: 600 }}>{TYPE_LABELS[p.payment_type] || p.payment_type}</td>
                    <td style={{ color: '#dce8f5', fontWeight: 700 }}>
                      {p.currency === 'USD' ? '$' : ''}{p.amount?.toLocaleString()}{p.currency === 'UZS' ? ' ₽' : ''}
                    </td>
                    <td>{statusBadge(p.status)}</td>
                    <td style={{ textTransform: 'capitalize', fontSize: 12 }}>{p.payment_method || '—'}</td>
                    <td style={{ fontSize: 12 }}>{new Date(p.created_at).toLocaleDateString('ru')}</td>
                    <td style={{ fontSize: 12 }}>{p.paid_at ? new Date(p.paid_at).toLocaleDateString('ru') : '—'}</td>
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
