import { useState, useEffect, useCallback } from 'react'
import { api } from '../api'
import Pagination from '../components/Pagination'
import Modal from '../components/Modal'
import { statusBadge } from '../components/Badge'
import { IcoCheck, IcoX, IcoDoc } from '../icons'

const TABS = [
  { value: 'pending',        label: 'На проверке' },
  { value: 'docs_requested', label: 'Доп. документы' },
  { value: 'approved',       label: 'Одобренные' },
  { value: 'rejected',       label: 'Отклонённые' },
]

const lbl = { display: 'block', fontSize: 12, fontWeight: 600, color: '#374151', marginBottom: 8 }

export default function Queue() {
  const [items, setItems]         = useState([])
  const [total, setTotal]         = useState(0)
  const [page, setPage]           = useState(1)
  const [status, setStatus]       = useState('pending')
  const [loading, setLoading]     = useState(true)
  const [detail, setDetail]       = useState(null)
  const [rejectModal, setRejectModal] = useState(null)
  const [docsModal, setDocsModal]     = useState(null)
  const [reason, setReason]   = useState('')
  const [message, setMessage] = useState('')
  const [saving, setSaving]   = useState(false)

  const load = useCallback(() => {
    setLoading(true)
    api.queue(status, page)
      .then(r => { setItems(r?.data || []); setTotal(r?.meta?.total || 0) })
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [status, page])

  useEffect(() => { load() }, [load])

  const handleApprove = async (item) => {
    if (!confirm(`Одобрить компанию "${item.company_name}"?`)) return
    setSaving(true)
    try { await api.approve(item.id); load(); if (detail?.id === item.id) setDetail(null) }
    finally { setSaving(false) }
  }

  const handleReject = async () => {
    if (!reason.trim()) return
    setSaving(true)
    try { await api.reject(rejectModal.id, reason); setRejectModal(null); setReason(''); load() }
    finally { setSaving(false) }
  }

  const handleRequestDocs = async () => {
    if (!message.trim()) return
    setSaving(true)
    try { await api.requestDocs(docsModal.id, message); setDocsModal(null); setMessage(''); load() }
    finally { setSaving(false) }
  }

  const openDetail = async (item) => {
    const r = await api.queueItem(item.id)
    setDetail(r?.data || item)
  }

  const isPending = status === 'pending' || status === 'docs_requested'

  const tabStyle = (v) => ({
    padding: '8px 16px', borderRadius: 8, fontSize: 13, fontWeight: 500,
    cursor: 'pointer', border: 'none', fontFamily: 'inherit', transition: 'all .12s',
    background: status === v ? '#2563eb' : 'transparent',
    color: status === v ? '#fff' : '#64748b',
  })

  return (
    <div style={{ padding: 28, display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div>
        <div style={{ fontSize: 20, fontWeight: 800, color: '#0f172a' }}>Очередь верификации</div>
        <div style={{ fontSize: 13, color: '#64748b', marginTop: 3 }}>Компании, ожидающие проверки</div>
      </div>

      <div style={{ display: 'flex', gap: 2, background: '#fff', border: '1px solid #e2e8f0', borderRadius: 10, padding: 3, width: 'fit-content' }}>
        {TABS.map(t => (
          <button key={t.value} style={tabStyle(t.value)} onClick={() => { setStatus(t.value); setPage(1) }}>
            {t.label}
          </button>
        ))}
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 14 }}>
        {loading ? (
          <div style={{ gridColumn: '1/-1', textAlign: 'center', padding: '48px 0', color: '#94a3b8' }}>Загрузка...</div>
        ) : items.length === 0 ? (
          <div className="card" style={{ gridColumn: '1/-1', textAlign: 'center', padding: '60px 0' }}>
            <IcoCheck size={36} style={{ color: '#e2e8f0', margin: '0 auto 12px', display: 'block' }} />
            <div style={{ color: '#64748b', fontWeight: 600 }}>Очередь пуста</div>
            <div style={{ color: '#94a3b8', fontSize: 13, marginTop: 6 }}>
              Нет компаний со статусом «{TABS.find(t => t.value === status)?.label}»
            </div>
          </div>
        ) : items.map(item => (
          <div
            key={item.id}
            className="card"
            onClick={() => openDetail(item)}
            style={{
              cursor: 'pointer', transition: 'border-color .12s, box-shadow .12s',
              borderColor: detail?.id === item.id ? '#bfdbfe' : '#e2e8f0',
              borderLeftWidth: item.is_urgent ? 3 : 1,
              borderLeftColor: item.is_urgent ? '#ef4444' : (detail?.id === item.id ? '#bfdbfe' : '#e2e8f0'),
            }}
            onMouseEnter={e => { if (detail?.id !== item.id) e.currentTarget.style.boxShadow = '0 4px 12px rgba(15,23,42,.08)' }}
            onMouseLeave={e => { e.currentTarget.style.boxShadow = '' }}
          >
            <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 10 }}>
              <div style={{ flex: 1 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
                  <span style={{ fontWeight: 700, color: '#0f172a' }}>{item.company_name}</span>
                  {item.is_urgent && (
                    <span className="badge badge-red">СРОЧНО</span>
                  )}
                </div>
                <div style={{ fontSize: 12.5, color: '#64748b', marginTop: 5 }}>
                  {item.user_name} · <span style={{ fontFamily: 'monospace' }}>{item.user_phone}</span>
                </div>
                <div style={{ display: 'flex', gap: 8, marginTop: 8, flexWrap: 'wrap', fontSize: 12, color: '#94a3b8' }}>
                  <span>{item.org_type?.toUpperCase()}</span>
                  <span>·</span>
                  <span>{item.country}</span>
                  {item.inn && <><span>·</span><span>ИНН: {item.inn}</span></>}
                </div>
              </div>
              <div style={{ flexShrink: 0 }}>{statusBadge(item.status)}</div>
            </div>

            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginTop: 14, paddingTop: 12, borderTop: '1px solid #f1f5f9' }}>
              <div style={{ fontSize: 12, color: '#94a3b8' }}>
                {new Date(item.created_at).toLocaleDateString('ru', { day: 'numeric', month: 'long', hour: '2-digit', minute: '2-digit' })}
              </div>
              {isPending && (
                <div style={{ display: 'flex', gap: 6 }} onClick={e => e.stopPropagation()}>
                  <button
                    onClick={() => handleApprove(item)} disabled={saving}
                    style={{
                      display: 'inline-flex', alignItems: 'center', gap: 4,
                      padding: '5px 10px', fontSize: 12, fontWeight: 600, cursor: 'pointer',
                      background: '#f0fdf4', color: '#16a34a',
                      border: '1px solid #bbf7d0', borderRadius: 6,
                    }}
                  ><IcoCheck size={11} /> Одобрить</button>
                  <button
                    onClick={() => setDocsModal(item)}
                    style={{ padding: '5px 10px', fontSize: 12, fontWeight: 600, cursor: 'pointer', background: '#eff6ff', color: '#2563eb', border: '1px solid #bfdbfe', borderRadius: 6, display: 'flex', alignItems: 'center' }}
                  ><IcoDoc size={12} /></button>
                  <button
                    onClick={() => setRejectModal(item)}
                    style={{ padding: '5px 10px', fontSize: 12, fontWeight: 600, cursor: 'pointer', background: '#fef2f2', color: '#dc2626', border: '1px solid #fecaca', borderRadius: 6, display: 'flex', alignItems: 'center' }}
                  ><IcoX size={12} /></button>
                </div>
              )}
            </div>
          </div>
        ))}
      </div>

      <Pagination page={page} total={total} onChange={setPage} />

      {/* Detail Panel */}
      {detail && (
        <div style={{
          position: 'fixed', top: 0, right: 0, bottom: 0, width: 380, zIndex: 40,
          background: '#fff', borderLeft: '1px solid #e2e8f0',
          display: 'flex', flexDirection: 'column',
          boxShadow: '-4px 0 24px rgba(15,23,42,.1)',
        }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '16px 20px', borderBottom: '1px solid #f1f5f9' }}>
            <span style={{ fontWeight: 700, color: '#0f172a', fontSize: 15 }}>{detail.company_name}</span>
            <button
              onClick={() => setDetail(null)}
              style={{ background: 'none', border: '1px solid #e2e8f0', borderRadius: 7, width: 28, height: 28, cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#94a3b8' }}
            ><IcoX size={13} /></button>
          </div>

          <div style={{ flex: 1, overflowY: 'auto', padding: 20, display: 'flex', flexDirection: 'column', gap: 20 }}>
            {[
              { title: 'Компания', rows: [['Тип', detail.org_type?.toUpperCase()], ['Страна', detail.country], ['Город', `${detail.city || ''}${detail.region ? `, ${detail.region}` : ''}`], ['Адрес', detail.street], ['ИНН', detail.inn]] },
              { title: 'Контакты', rows: [['Email', detail.email || '—'], ['Телефон', detail.phone || '—']] },
              { title: 'Владелец', rows: [['Имя', detail.user_name], ['Телефон', detail.user_phone]] },
            ].map(section => (
              <div key={section.title}>
                <div style={{ fontSize: 10.5, fontWeight: 700, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '.08em', marginBottom: 10 }}>
                  {section.title}
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: 7 }}>
                  {section.rows.filter(([, v]) => v).map(([k, v]) => (
                    <div key={k} style={{ display: 'flex', gap: 8 }}>
                      <span style={{ fontSize: 12.5, color: '#94a3b8', minWidth: 70 }}>{k}:</span>
                      <span style={{ fontSize: 12.5, color: '#334155' }}>{v}</span>
                    </div>
                  ))}
                </div>
              </div>
            ))}

            {detail.reg_doc_url && (
              <div>
                <div style={{ fontSize: 10.5, fontWeight: 700, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '.08em', marginBottom: 10 }}>Документы</div>
                <a href={detail.reg_doc_url} target="_blank" rel="noreferrer" style={{
                  display: 'flex', alignItems: 'center', gap: 8, padding: '10px 12px',
                  background: '#eff6ff', color: '#2563eb',
                  border: '1px solid #bfdbfe', borderRadius: 8,
                  textDecoration: 'none', fontSize: 13, fontWeight: 500,
                }}>
                  <IcoDoc size={15} /> Регистрационный документ
                </a>
              </div>
            )}

            {detail.rejection_reason && (
              <div style={{ background: '#fef2f2', border: '1px solid #fecaca', borderRadius: 8, padding: 12 }}>
                <div style={{ fontSize: 12, fontWeight: 600, color: '#dc2626', marginBottom: 4 }}>Причина отклонения:</div>
                <div style={{ fontSize: 13, color: '#991b1b' }}>{detail.rejection_reason}</div>
              </div>
            )}
          </div>

          {isPending && (
            <div style={{ padding: '12px 16px', borderTop: '1px solid #f1f5f9', display: 'flex', gap: 8 }}>
              <button
                onClick={() => handleApprove(detail)}
                style={{ flex: 1, padding: '10px', background: '#16a34a', color: '#fff', border: 'none', borderRadius: 8, cursor: 'pointer', fontSize: 13, fontWeight: 700, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6 }}
              ><IcoCheck size={14} /> Одобрить</button>
              <button
                onClick={() => setDocsModal(detail)}
                style={{ padding: '10px 14px', background: '#eff6ff', color: '#2563eb', border: '1px solid #bfdbfe', borderRadius: 8, cursor: 'pointer', display: 'flex', alignItems: 'center' }}
              ><IcoDoc size={15} /></button>
              <button
                onClick={() => setRejectModal(detail)}
                style={{ padding: '10px 14px', background: '#fef2f2', color: '#dc2626', border: '1px solid #fecaca', borderRadius: 8, cursor: 'pointer', display: 'flex', alignItems: 'center' }}
              ><IcoX size={15} /></button>
            </div>
          )}
        </div>
      )}

      {rejectModal && (
        <Modal title={`Отклонить: ${rejectModal.company_name}`} onClose={() => { setRejectModal(null); setReason('') }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <div>
              <label style={lbl}>Причина отклонения <span style={{ color: '#ef4444' }}>*</span></label>
              <textarea className="input" style={{ minHeight: 100 }} placeholder="Укажите причину..."
                value={reason} onChange={e => setReason(e.target.value)} autoFocus />
            </div>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
              <button className="btn-secondary" onClick={() => setRejectModal(null)}>Отмена</button>
              <button className="btn-danger" onClick={handleReject} disabled={saving || !reason.trim()}>Отклонить</button>
            </div>
          </div>
        </Modal>
      )}

      {docsModal && (
        <Modal title={`Запросить документы: ${docsModal.company_name}`} onClose={() => { setDocsModal(null); setMessage('') }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <div>
              <label style={lbl}>Сообщение <span style={{ color: '#ef4444' }}>*</span></label>
              <textarea className="input" style={{ minHeight: 100 }} placeholder="Какие документы нужны..."
                value={message} onChange={e => setMessage(e.target.value)} autoFocus />
            </div>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
              <button className="btn-secondary" onClick={() => setDocsModal(null)}>Отмена</button>
              <button className="btn-primary" onClick={handleRequestDocs} disabled={saving || !message.trim()}>Отправить</button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  )
}
