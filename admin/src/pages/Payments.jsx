import { useState, useEffect, useCallback } from 'react'
import { api } from '../api'
import Pagination from '../components/Pagination'
import { statusBadge } from '../components/Badge'
import { IcoTrendUp, IcoCard } from '../icons'

const TYPE_LABELS = {
  tokens: 'Токены', subscription: 'Подписка', listing: 'Объявление', boost: 'Продвижение',
}

function SumCard({ label, value, icon: Ico, accent }) {
  return (
    <div className="card" style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
      <div style={{
        width: 44, height: 44, borderRadius: 12, flexShrink: 0,
        background: `${accent}14`, color: accent,
        display: 'flex', alignItems: 'center', justifyContent: 'center',
      }}>
        <Ico size={20} />
      </div>
      <div>
        <div style={{ fontSize: 22, fontWeight: 800, color: '#0f172a', lineHeight: 1.1 }}>{value}</div>
        <div style={{ fontSize: 12.5, color: '#64748b', marginTop: 4 }}>{label}</div>
      </div>
    </div>
  )
}

export default function Payments() {
  const [items, setItems]   = useState([])
  const [total, setTotal]   = useState(0)
  const [page, setPage]     = useState(1)
  const [loading, setLoading] = useState(true)

  const load = useCallback(() => {
    setLoading(true)
    api.payments(page)
      .then(r => { setItems(r?.data || []); setTotal(r?.meta?.total || 0) })
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [page])

  useEffect(() => { load() }, [load])

  const totalPaid  = items.filter(p => p.status === 'paid').reduce((s, p) => s + (p.amount || 0), 0)
  const paidCount  = items.filter(p => p.status === 'paid').length

  return (
    <div style={{ padding: 28, display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div>
        <div style={{ fontSize: 20, fontWeight: 800, color: '#0f172a' }}>Платежи</div>
        <div style={{ fontSize: 13, color: '#64748b', marginTop: 3 }}>Всего транзакций: {total}</div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 14 }}>
        <SumCard label="Оплачено на странице (UZS)" value={totalPaid.toLocaleString()}   icon={IcoCard}     accent="#2563eb" />
        <SumCard label="Всего транзакций"             value={total}                       icon={IcoTrendUp}  accent="#d97706" />
        <SumCard label="Успешных на странице"         value={paidCount}                   icon={IcoCard}     accent="#16a34a" />
      </div>

      <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
        {loading ? (
          <div style={{ textAlign: 'center', padding: '48px 0', color: '#94a3b8' }}>Загрузка...</div>
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
                  <tr><td colSpan={7} style={{ textAlign: 'center', padding: '40px 0', color: '#94a3b8' }}>Нет платежей</td></tr>
                ) : items.map(p => (
                  <tr key={p.id}>
                    <td style={{ paddingLeft: 20 }}>
                      <code style={{ fontSize: 11.5, color: '#94a3b8', background: '#f8fafc', padding: '2px 6px', borderRadius: 4 }}>{p.id?.slice(0, 8)}…</code>
                    </td>
                    <td style={{ color: '#0f172a', fontWeight: 600 }}>{TYPE_LABELS[p.payment_type] || p.payment_type}</td>
                    <td style={{ color: '#0f172a', fontWeight: 700 }}>
                      {p.currency === 'USD' ? '$' : ''}{p.amount?.toLocaleString()}{p.currency === 'UZS' ? ' сум' : ''}
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
